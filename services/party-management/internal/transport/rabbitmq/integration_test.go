package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/party-management/internal/domain"
	"tmf/services/party-management/internal/infrastructure/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	rabbitContainer "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	sharedDB        *gorm.DB
	sharedRepo      *postgres.PartyRepository
	sharedConn      *amqp.Connection
	sharedPublisher rabbitmq.Publisher
	pgInstance      testcontainers.Container
	rabbitInstance  testcontainers.Container
)

type IntegrationTestSuite struct {
	DB        *gorm.DB
	Repo      *postgres.PartyRepository
	Conn      *amqp.Connection
	Publisher rabbitmq.Publisher
	Listener  *Listener
	Handlers  *Handlers
	EventChan <-chan amqp.Delivery
	channel   *amqp.Channel
}

func runPGContainer(ctx context.Context) (*pgContainer.PostgresContainer, error) {
	type outcome struct {
		val *pgContainer.PostgresContainer
		err error
	}
	ch := make(chan outcome, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- outcome{err: fmt.Errorf("testcontainers panic: %v", r)}
			}
		}()
		v, e := pgContainer.Run(ctx,
			"postgres:15",
			pgContainer.WithDatabase("testdb"),
			pgContainer.WithUsername("postgres"),
			pgContainer.WithPassword("postgres"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second)),
		)
		ch <- outcome{val: v, err: e}
	}()
	o := <-ch
	return o.val, o.err
}

func runRabbitContainer(ctx context.Context) (*rabbitContainer.RabbitMQContainer, error) {
	type outcome struct {
		val *rabbitContainer.RabbitMQContainer
		err error
	}
	ch := make(chan outcome, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- outcome{err: fmt.Errorf("testcontainers panic: %v", r)}
			}
		}()
		v, e := rabbitContainer.Run(ctx,
			"rabbitmq:3-management",
			testcontainers.WithWaitStrategy(
				wait.ForLog("Server startup complete").
					WithStartupTimeout(60*time.Second)),
		)
		ch <- outcome{val: v, err: e}
	}()
	o := <-ch
	return o.val, o.err
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := runPGContainer(ctx)
	if err != nil {
		fmt.Printf("Skipping testcontainers setup due to postgres error: %s\n", err)
		os.Exit(m.Run())
	}
	pgInstance = pg

	pgConnStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres connection string: %s", err)
	}

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "infrastructure", "postgres", "migrations")

	mig, err := migrate.New("file://"+migrationsPath, pgConnStr)
	if err != nil {
		log.Fatalf("failed to create migrate: %s", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %s", err)
	}

	sharedDB, err = gorm.Open(gormPostgres.Open(pgConnStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %s", err)
	}

	sharedRepo = postgres.NewPartyRepository(sharedDB)

	rabbit, err := runRabbitContainer(ctx)
	if err != nil {
		fmt.Printf("Skipping testcontainers setup due to rabbitmq error: %s\n", err)
		os.Exit(m.Run())
	}
	rabbitInstance = rabbit

	rabbitURL, err := rabbit.AmqpURL(ctx)
	if err != nil {
		log.Fatalf("failed to get rabbitmq URL: %s", err)
	}

	for range 10 {
		sharedConn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %s", err)
	}

	sharedPublisher, err = rabbitmq.NewPublisherWithConnection(sharedConn)
	if err != nil {
		log.Fatalf("failed to create publisher: %s", err)
	}

	// Declare the exchanges and queues that K8s CRDs provide in production.
	// The listener now uses QueueDeclarePassive and expects these to pre-exist.
	ch, err := sharedConn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %s", err)
	}
	for _, exchange := range []struct{ name, kind string }{
		{EventExchange, "topic"},
		{CommandExchange, "topic"},
		{DeadLetterExchange, "fanout"},
	} {
		if err := ch.ExchangeDeclare(exchange.name, exchange.kind, true, false, false, false, nil); err != nil {
			log.Fatalf("failed to declare exchange %s: %s", exchange.name, err)
		}
	}
	for _, queue := range []string{DeadLetterQueue, PartyQueue, "party.events"} {
		if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
			log.Fatalf("failed to declare queue %s: %s", queue, err)
		}
	}
	_ = ch.Close()

	code := m.Run()

	_ = sharedConn.Close()
	_ = pgInstance.Terminate(ctx)
	_ = rabbitInstance.Terminate(ctx)

	os.Exit(code)
}

func setupTestSuite(t *testing.T) *IntegrationTestSuite {
	if sharedConn == nil || sharedDB == nil {
		t.Skip("Skipping integration test: testcontainers unavailable")
	}

	ch, err := sharedConn.Channel()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = ch.Close()
	})

	queueName := fmt.Sprintf("test.events.%s", t.Name())
	eventQueue, err := ch.QueueDeclare(queueName, false, true, false, false, nil)
	require.NoError(t, err)

	err = ch.QueueBind(eventQueue.Name, "evt.party.*", EventExchange, false, nil)
	require.NoError(t, err)

	events, err := ch.Consume(eventQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	listener, err := NewListener(sharedConn)
	require.NoError(t, err)

	tm := postgres.NewTransactionManager(sharedDB)
	outboxRepo := postgres.NewOutboxRepository(sharedDB)
	outboxPublisher := postgres.NewOutboxPublisher(outboxRepo)

	// Clear any pending outbox events left over from previous tests so their
	// workers don't publish stale events into this test's event queue.
	sharedDB.Exec("DELETE FROM outbox_events WHERE status = 'PENDING'")

	worker := postgres.NewOutboxWorker(outboxRepo, sharedPublisher)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go worker.Start(ctx)

	handlers := NewHandlers(sharedRepo, outboxPublisher, sharedPublisher, tm)

	return &IntegrationTestSuite{
		DB:        sharedDB,
		Repo:      sharedRepo,
		Conn:      sharedConn,
		Publisher: sharedPublisher,
		Listener:  listener,
		Handlers:  handlers,
		EventChan: events,
		channel:   ch,
	}
}

func (s *IntegrationTestSuite) waitForEvent(_ *testing.T, timeout time.Duration) *amqp.Delivery {
	select {
	case event := <-s.EventChan:
		return &event
	case <-time.After(timeout):
		return nil
	}
}

func TestIntegration_CreateIndividual(t *testing.T) {
	suite := setupTestSuite(t)

	payload := map[string]any{
		"@type":      "Individual",
		"id":         "int-ind-1",
		"givenName":  "Integration",
		"familyName": "Test",
		"href":       "http://example.com/int-ind-1",
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleCreateParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	saved, err := suite.Repo.GetIndividual(context.Background(), "int-ind-1")
	require.NoError(t, err)
	assert.Equal(t, "Integration", saved.GivenName)
	assert.Equal(t, "Test", saved.FamilyName)
	assert.Equal(t, "Initialized", saved.Status)

	evt := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt, "Expected evt.party.created event")
	assert.Equal(t, EvtPartyCreated, evt.RoutingKey)
}

func TestIntegration_CreateOrganization(t *testing.T) {
	suite := setupTestSuite(t)

	payload := map[string]any{
		"@type":         "Organization",
		"id":            "int-org-1",
		"tradingName":   "IntegrationCorp",
		"isLegalEntity": true,
		"href":          "http://example.com/int-org-1",
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleCreateParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	saved, err := suite.Repo.GetOrganization(context.Background(), "int-org-1")
	require.NoError(t, err)
	assert.Equal(t, "IntegrationCorp", saved.TradingName)
	assert.Equal(t, true, saved.IsLegalEntity)

	evt := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt)
	assert.Equal(t, EvtPartyCreated, evt.RoutingKey)
}

func TestIntegration_UpdateIndividual(t *testing.T) {
	suite := setupTestSuite(t)

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "int-upd-ind-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Initialized",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Original",
		FamilyName: "Name",
	}
	require.NoError(t, suite.Repo.CreateIndividual(context.Background(), ind))

	payload := map[string]any{
		"id":         "int-upd-ind-1",
		"@type":      "Individual",
		"status":     "Active",
		"givenName":  "Updated",
		"familyName": "Person",
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleUpdateParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	updated, err := suite.Repo.GetIndividual(context.Background(), "int-upd-ind-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.GivenName)
	assert.Equal(t, "Active", updated.Status)

	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1)
	evt2 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt2)

	keys := []string{evt1.RoutingKey, evt2.RoutingKey}
	assert.Contains(t, keys, EvtPartyUpdated)
	assert.Contains(t, keys, EvtPartyStateChange)
}

func TestIntegration_PatchParty(t *testing.T) {
	suite := setupTestSuite(t)

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "int-patch-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Initialized",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "PatchMe",
		FamilyName: "Please",
	}
	require.NoError(t, suite.Repo.CreateIndividual(context.Background(), ind))

	newName := "Patched"
	newStatus := "Validated"
	payload := PatchPartyPayload{
		ID:        "int-patch-1",
		GivenName: &newName,
		Status:    &newStatus,
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandlePatchParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	patched, err := suite.Repo.GetIndividual(context.Background(), "int-patch-1")
	require.NoError(t, err)
	assert.Equal(t, "Patched", patched.GivenName)
	assert.Equal(t, "Please", patched.FamilyName)
	assert.Equal(t, "Validated", patched.Status)

	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1)
	evt2 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt2)

	keys := []string{evt1.RoutingKey, evt2.RoutingKey}
	assert.Contains(t, keys, EvtPartyUpdated)
	assert.Contains(t, keys, EvtPartyStateChange)
}

func TestIntegration_DeleteParty_StartsSaga(t *testing.T) {
	suite := setupTestSuite(t)

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "int-del-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "DeleteMe",
		FamilyName: "Saga",
	}
	require.NoError(t, suite.Repo.CreateIndividual(context.Background(), ind))

	payload := DeletePartyPayload{ID: "int-del-1"}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleDeleteParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	saved, err := suite.Repo.GetIndividual(context.Background(), "int-del-1")
	require.NoError(t, err)
	assert.Equal(t, "DeletionPending", saved.Status)

	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1, "Expected initiation event")
	assert.Equal(t, EvtPartyDeletionInitiated, evt1.RoutingKey)

	evt2 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt2, "Expected state change event")
	assert.Equal(t, EvtPartyStateChange, evt2.RoutingKey)
}

func TestIntegration_FinalizeDeletion(t *testing.T) {
	suite := setupTestSuite(t)

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "int-final-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "DeletionPending",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Finalize",
		FamilyName: "Me",
	}
	require.NoError(t, suite.Repo.CreateIndividual(context.Background(), ind))

	payload := DeletePartyPayload{ID: "int-final-1"}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleFinalizeDeletion(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	saved, err := suite.Repo.GetIndividual(context.Background(), "int-final-1")
	require.NoError(t, err)
	assert.Equal(t, "Deleted", saved.Status)

	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1)
	assert.Equal(t, EvtPartyDeleted, evt1.RoutingKey)

	evt2 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt2)
	assert.Equal(t, EvtPartyStateChange, evt2.RoutingKey)
}

func TestIntegration_GetParty(t *testing.T) {
	suite := setupTestSuite(t)

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "int-get-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "GetMe",
		FamilyName: "Query",
	}
	require.NoError(t, suite.Repo.CreateIndividual(context.Background(), ind))

	payload := GetPartyPayload{ID: "int-get-1"}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleGetParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)
}

func TestIntegration_SearchParty(t *testing.T) {
	suite := setupTestSuite(t)

	ind1 := &domain.Individual{
		Party: domain.Party{
			ID:        "int-search-ind-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "SearchAlice",
		FamilyName: "TestFamily",
	}
	org1 := &domain.Organization{
		Party: domain.Party{
			ID:        "int-search-org-1",
			Type:      domain.PartyTypeOrganization,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "SearchCorp",
		IsLegalEntity: true,
	}
	require.NoError(t, suite.Repo.CreateIndividual(context.Background(), ind1))
	require.NoError(t, suite.Repo.CreateOrganization(context.Background(), org1))

	replyQueue, err := suite.channel.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	replies, err := suite.channel.Consume(replyQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	givenName := "SearchAlice"
	payload := SearchPartyPayload{GivenName: &givenName}
	body, _ := json.Marshal(payload)

	err = suite.Handlers.HandleSearchParty(context.Background(), amqp.Delivery{
		Body:          body,
		ReplyTo:       replyQueue.Name,
		CorrelationId: "search-test-1",
	})
	require.NoError(t, err)

	select {
	case reply := <-replies:
		assert.Equal(t, "search-test-1", reply.CorrelationId)

		var results []map[string]any
		err := json.Unmarshal(reply.Body, &results)
		require.NoError(t, err, "Failed to parse search response")
		require.Len(t, results, 1, "Expected exactly one search result")

		result := results[0]
		assert.Equal(t, "int-search-ind-1", result["id"])
		assert.Equal(t, "Individual", result["@type"])
		assert.Equal(t, "SearchAlice", result["givenName"], "givenName should be present in response")
		assert.Equal(t, "TestFamily", result["familyName"], "familyName should be present in response")

	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for search reply")
	}
}

func TestIntegration_SearchParty_ReturnsCompleteOrganizationData(t *testing.T) {
	suite := setupTestSuite(t)

	org := &domain.Organization{
		Party: domain.Party{
			ID:        "int-search-org-data-1",
			Type:      domain.PartyTypeOrganization,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "CompleteDataCorp",
		IsLegalEntity: true,
	}
	require.NoError(t, suite.Repo.CreateOrganization(context.Background(), org))

	replyQueue, err := suite.channel.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	replies, err := suite.channel.Consume(replyQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	tradingName := "CompleteDataCorp"
	payload := SearchPartyPayload{TradingName: &tradingName}
	body, _ := json.Marshal(payload)

	err = suite.Handlers.HandleSearchParty(context.Background(), amqp.Delivery{
		Body:          body,
		ReplyTo:       replyQueue.Name,
		CorrelationId: "search-org-test-1",
	})
	require.NoError(t, err)

	select {
	case reply := <-replies:
		var results []map[string]any
		err := json.Unmarshal(reply.Body, &results)
		require.NoError(t, err)
		require.Len(t, results, 1)

		result := results[0]
		assert.Equal(t, "int-search-org-data-1", result["id"])
		assert.Equal(t, "Organization", result["@type"])
		assert.Equal(t, "CompleteDataCorp", result["tradingName"], "tradingName should be present in response")
		assert.Equal(t, true, result["isLegalEntity"], "isLegalEntity should be present in response")

	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for search reply")
	}
}

func TestIntegration_SearchParty_MixedTypes_ReturnsAllFields(t *testing.T) {
	suite := setupTestSuite(t)

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "int-search-mixed-ind",
			Type:      domain.PartyTypeIndividual,
			Status:    "Validated",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "MixedTest",
		FamilyName: "Person",
	}
	org := &domain.Organization{
		Party: domain.Party{
			ID:        "int-search-mixed-org",
			Type:      domain.PartyTypeOrganization,
			Status:    "Validated",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "MixedTestCorp",
		IsLegalEntity: false,
	}
	require.NoError(t, suite.Repo.CreateIndividual(context.Background(), ind))
	require.NoError(t, suite.Repo.CreateOrganization(context.Background(), org))

	replyQueue, err := suite.channel.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	replies, err := suite.channel.Consume(replyQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	payload := SearchPartyPayload{}
	body, _ := json.Marshal(payload)

	err = suite.Handlers.HandleSearchParty(context.Background(), amqp.Delivery{
		Body:          body,
		ReplyTo:       replyQueue.Name,
		CorrelationId: "search-mixed-test",
	})
	require.NoError(t, err)

	select {
	case reply := <-replies:
		var results []map[string]any
		err := json.Unmarshal(reply.Body, &results)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(results), 2, "Expected at least 2 results")

		foundInd := false
		foundOrg := false
		for _, result := range results {
			if result["id"] == "int-search-mixed-ind" {
				foundInd = true
				assert.Equal(t, "MixedTest", result["givenName"], "Individual should have givenName")
				assert.Equal(t, "Person", result["familyName"], "Individual should have familyName")
			}
			if result["id"] == "int-search-mixed-org" {
				foundOrg = true
				assert.Equal(t, "MixedTestCorp", result["tradingName"], "Organization should have tradingName")
				assert.Equal(t, false, result["isLegalEntity"], "Organization should have isLegalEntity")
			}
		}
		assert.True(t, foundInd, "Should find the test Individual")
		assert.True(t, foundOrg, "Should find the test Organization")

	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for search reply")
	}
}
func TestIntegration_AuditTrail(t *testing.T) {
	suite := setupTestSuite(t)

	userID := "party-audit-123"
	partyID := uuid.New().String()
	payload := CreatePartyPayload{
		Type: "Individual",
		Individual: &CreateIndividualPayload{
			ID:         partyID,
			GivenName:  "Audit",
			FamilyName: "Test",
		},
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleCreateParty(context.Background(), amqp.Delivery{
		Body:    body,
		Headers: amqp.Table{"user": userID},
	})
	require.NoError(t, err)

	type LoggedAction struct {
		TableName string `gorm:"column:table_name"`
		UserName  string `gorm:"column:user_name"`
		Action    string `gorm:"column:action"`
	}

	var auditLog LoggedAction
	err = suite.DB.Table("audit.logged_actions").
		Where("table_name = ? AND action = ?", "parties", "I").
		Order("action_tstamp_clk DESC").
		First(&auditLog).Error

	require.NoError(t, err)
	assert.Equal(t, userID, auditLog.UserName)
	assert.Equal(t, "I", auditLog.Action)
}

func TestListener_Routing(t *testing.T) {
	suite := setupTestSuite(t)

	ctx := t.Context()

	go func() {
		err := suite.Listener.Start(ctx, suite.Handlers)
		if err != nil && err != context.Canceled {
			log.Printf("Listener stopped with error: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	payload := map[string]any{
		"@type":      "Individual",
		"id":         "route-ind-1",
		"givenName":  "Routed",
		"familyName": "User",
		"href":       "http://example.com/route-ind-1",
	}
	body, _ := json.Marshal(payload)

	err := suite.channel.PublishWithContext(context.Background(), CommandExchange, CmdPartyCreate, false, false, amqp.Publishing{
		ContentType: "application/json",
		Headers:     amqp.Table{"Authorization": "Bearer test-token"},
		Body:        body,
	})
	require.NoError(t, err)

	var saved *domain.Individual
	for range 10 {
		saved, err = suite.Repo.GetIndividual(context.Background(), "route-ind-1")
		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	require.NoError(t, err, "Failed to find routed individual in DB")
	assert.Equal(t, "Routed", saved.GivenName)
}

func TestIntegration_HeaderPropagation(t *testing.T) {
	suite := setupTestSuite(t)

	ch, err := suite.Conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	err = ch.QueueBind(q.Name, EvtPartyDeletionInitiated, EventExchange, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	require.NoError(t, err)

	ctx := context.Background()
	headers := amqp.Table{
		"user":          "test-user",
		"Authorization": "Bearer test-token",
	}

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "prop-test-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Test",
		FamilyName: "Prop",
	}
	require.NoError(t, suite.Repo.CreateIndividual(ctx, ind))

	payload := DeletePartyPayload{ID: "prop-test-1"}
	body, _ := json.Marshal(payload)

	err = suite.Handlers.HandleDeleteParty(ctx, amqp.Delivery{
		Body:    body,
		Headers: headers,
	})
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.Equal(t, "test-user", msg.Headers["user"])
		assert.Equal(t, "Bearer test-token", msg.Headers["Authorization"])
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestIntegration_DeleteParty_Idempotency(t *testing.T) {
	suite := setupTestSuite(t)

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "idemp-test-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "DeletionPending",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Stuck",
		FamilyName: "Delete",
	}
	require.NoError(t, suite.Repo.CreateIndividual(context.Background(), ind))

	ch, err := suite.Conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	replyQ, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(replyQ.Name, "", true, true, false, false, nil)
	require.NoError(t, err)

	payload := DeletePartyPayload{ID: "idemp-test-1"}
	body, _ := json.Marshal(payload)

	err = suite.Handlers.HandleDeleteParty(context.Background(), amqp.Delivery{
		Body:          body,
		ReplyTo:       replyQ.Name,
		CorrelationId: "corr-idemp-1",
	})
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.Equal(t, "corr-idemp-1", msg.CorrelationId)
		var resp map[string]string
		_ = json.Unmarshal(msg.Body, &resp)
		assert.Equal(t, "deletion_initiated", resp["status"])
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for reply")
	}
}
