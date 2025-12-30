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

	"tmf/services/party-management/internal/domain"
	"tmf/services/party-management/internal/infrastructure/postgres"
	infraRabbit "tmf/services/party-management/internal/infrastructure/rabbitmq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

// Shared test infrastructure - initialized once in TestMain
var (
	sharedDB        *gorm.DB
	sharedRepo      *postgres.PartyRepository
	sharedConn      *amqp.Connection
	sharedPublisher *infraRabbit.Publisher
	pgInstance      testcontainers.Container
	rabbitInstance  testcontainers.Container
)

// IntegrationTestSuite holds resources for a single test
type IntegrationTestSuite struct {
	DB        *gorm.DB
	Repo      *postgres.PartyRepository
	Conn      *amqp.Connection
	Publisher *infraRabbit.Publisher
	Listener  *Listener
	Handlers  *Handlers
	EventChan <-chan amqp.Delivery
	channel   *amqp.Channel
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start Postgres container
	var err error
	pg, err := pgContainer.Run(ctx,
		"postgres:15",
		pgContainer.WithDatabase("testdb"),
		pgContainer.WithUsername("postgres"),
		pgContainer.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start postgres: %s", err)
	}
	pgInstance = pg

	pgConnStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres connection string: %s", err)
	}

	// Run migrations
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

	// 2. Start RabbitMQ container
	rabbit, err := rabbitContainer.Run(ctx,
		"rabbitmq:3-management",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete").
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start rabbitmq: %s", err)
	}
	rabbitInstance = rabbit

	rabbitURL, err := rabbit.AmqpURL(ctx)
	if err != nil {
		log.Fatalf("failed to get rabbitmq URL: %s", err)
	}

	// Connect to RabbitMQ with retry
	for i := 0; i < 10; i++ {
		sharedConn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %s", err)
	}

	// Create shared Publisher
	sharedPublisher, err = infraRabbit.NewPublisher(sharedConn)
	if err != nil {
		log.Fatalf("failed to create publisher: %s", err)
	}

	// Declare exchange once
	ch, _ := sharedPublisher.GetChannel()
	if err := ch.ExchangeDeclare(EventExchange, "topic", true, false, false, false, nil); err != nil {
		log.Fatalf("failed to declare exchange: %s", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	_ = sharedConn.Close()
	_ = pgInstance.Terminate(ctx)
	_ = rabbitInstance.Terminate(ctx)

	os.Exit(code)
}

// setupTestSuite creates a per-test suite using shared containers
// It creates a fresh event queue for each test to avoid event cross-pollution
func setupTestSuite(t *testing.T) *IntegrationTestSuite {
	// Create a new channel for this test's event queue
	ch, err := sharedConn.Channel()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = ch.Close()
	})

	// Create a unique event queue for this test
	queueName := fmt.Sprintf("test.events.%s", t.Name())
	eventQueue, err := ch.QueueDeclare(queueName, false, true, false, false, nil)
	require.NoError(t, err)

	err = ch.QueueBind(eventQueue.Name, "evt.party.*", EventExchange, false, nil)
	require.NoError(t, err)

	events, err := ch.Consume(eventQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	// Create Listener for this test
	listener, err := NewListener(sharedConn)
	require.NoError(t, err)

	// Create Handlers
	handlers := NewHandlers(sharedRepo, sharedPublisher)

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

// Helper to wait for event
func (s *IntegrationTestSuite) waitForEvent(t *testing.T, timeout time.Duration) *amqp.Delivery {
	select {
	case event := <-s.EventChan:
		return &event
	case <-time.After(timeout):
		return nil
	}
}

// --- Integration Tests ---

func TestIntegration_CreateIndividual(t *testing.T) {
	suite := setupTestSuite(t)

	payload := map[string]interface{}{
		"@type":      "Individual",
		"id":         "int-ind-1",
		"givenName":  "Integration",
		"familyName": "Test",
		"href":       "http://example.com/int-ind-1",
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleCreateParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB
	saved, err := suite.Repo.GetIndividual(context.Background(), "int-ind-1")
	require.NoError(t, err)
	assert.Equal(t, "Integration", saved.GivenName)
	assert.Equal(t, "Test", saved.FamilyName)
	assert.Equal(t, "Initialized", saved.Status)

	// Verify events
	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1, "Expected evt.party.created event")
	assert.Equal(t, EvtPartyCreated, evt1.RoutingKey)

	evt2 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt2, "Expected evt.party.stateChange event")
	assert.Equal(t, EvtPartyStateChange, evt2.RoutingKey)
}

func TestIntegration_CreateOrganization(t *testing.T) {
	suite := setupTestSuite(t)

	payload := map[string]interface{}{
		"@type":         "Organization",
		"id":            "int-org-1",
		"tradingName":   "IntegrationCorp",
		"isLegalEntity": true,
		"href":          "http://example.com/int-org-1",
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleCreateParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB
	saved, err := suite.Repo.GetOrganization(context.Background(), "int-org-1")
	require.NoError(t, err)
	assert.Equal(t, "IntegrationCorp", saved.TradingName)
	assert.Equal(t, true, saved.IsLegalEntity)

	// Verify event
	evt := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt)
	assert.Equal(t, EvtPartyCreated, evt.RoutingKey)
}

func TestIntegration_UpdateIndividual(t *testing.T) {
	suite := setupTestSuite(t)

	// Create first
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

	// Update via handler
	payload := map[string]interface{}{
		"id":         "int-upd-ind-1",
		"@type":      "Individual",
		"status":     "Active",
		"givenName":  "Updated",
		"familyName": "Person",
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleUpdateParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB
	updated, err := suite.Repo.GetIndividual(context.Background(), "int-upd-ind-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.GivenName)
	assert.Equal(t, "Active", updated.Status)

	// Verify events (updated + stateChange)
	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1)
	assert.Equal(t, EvtPartyUpdated, evt1.RoutingKey)

	evt2 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt2)
	assert.Equal(t, EvtPartyStateChange, evt2.RoutingKey)
}

func TestIntegration_PatchParty(t *testing.T) {
	suite := setupTestSuite(t)

	// Create first
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

	// Patch via handler
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

	// Verify DB
	patched, err := suite.Repo.GetIndividual(context.Background(), "int-patch-1")
	require.NoError(t, err)
	assert.Equal(t, "Patched", patched.GivenName)
	assert.Equal(t, "Please", patched.FamilyName) // Unchanged
	assert.Equal(t, "Validated", patched.Status)

	// Verify events
	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1)
	assert.Equal(t, EvtPartyUpdated, evt1.RoutingKey)
}

func TestIntegration_DeleteParty_StartsSaga(t *testing.T) {
	suite := setupTestSuite(t)

	// Create first
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

	// Delete via handler
	payload := DeletePartyPayload{ID: "int-del-1"}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleDeleteParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB - should be DeletionPending
	saved, err := suite.Repo.GetIndividual(context.Background(), "int-del-1")
	require.NoError(t, err)
	assert.Equal(t, "DeletionPending", saved.Status)

	// Verify events
	// 1. Deletion Initiated
	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1, "Expected initiation event")
	assert.Equal(t, EvtPartyDeletionInitiated, evt1.RoutingKey)

	// 2. State Change
	evt2 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt2, "Expected state change event")
	assert.Equal(t, EvtPartyStateChange, evt2.RoutingKey)
}

func TestIntegration_FinalizeDeletion(t *testing.T) {
	suite := setupTestSuite(t)

	// Create Pending Party
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

	// Finalize via handler
	payload := DeletePartyPayload{ID: "int-final-1"}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleFinalizeDeletion(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB - Status Deleted (Soft Delete)
	saved, err := suite.Repo.GetIndividual(context.Background(), "int-final-1")
	require.NoError(t, err)
	assert.Equal(t, "Deleted", saved.Status)

	// Verify events
	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1)
	assert.Equal(t, EvtPartyDeleted, evt1.RoutingKey)

	evt2 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt2)
	assert.Equal(t, EvtPartyStateChange, evt2.RoutingKey)
}

func TestIntegration_GetParty(t *testing.T) {
	suite := setupTestSuite(t)

	// Create test data
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

	// Query via handler (no ReplyTo, just verify no error)
	payload := GetPartyPayload{ID: "int-get-1"}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleGetParty(context.Background(), amqp.Delivery{Body: body})
	require.NoError(t, err)
}

func TestIntegration_SearchParty(t *testing.T) {
	suite := setupTestSuite(t)

	// Create test data - Individual and Organization
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

	// Create reply queue for RPC
	replyQueue, err := suite.channel.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	replies, err := suite.channel.Consume(replyQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	// Search by givenName - should return Individual with complete data
	givenName := "SearchAlice"
	payload := SearchPartyPayload{GivenName: &givenName}
	body, _ := json.Marshal(payload)

	err = suite.Handlers.HandleSearchParty(context.Background(), amqp.Delivery{
		Body:          body,
		ReplyTo:       replyQueue.Name,
		CorrelationId: "search-test-1",
	})
	require.NoError(t, err)

	// Receive and parse reply
	select {
	case reply := <-replies:
		assert.Equal(t, "search-test-1", reply.CorrelationId)

		// Parse response as array of interfaces
		var results []map[string]interface{}
		err := json.Unmarshal(reply.Body, &results)
		require.NoError(t, err, "Failed to parse search response")
		require.Len(t, results, 1, "Expected exactly one search result")

		// Verify complete Individual data is returned (not base Party)
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

	// Create test organization
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

	// Create reply queue for RPC
	replyQueue, err := suite.channel.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	replies, err := suite.channel.Consume(replyQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	// Search by tradingName
	tradingName := "CompleteDataCorp"
	payload := SearchPartyPayload{TradingName: &tradingName}
	body, _ := json.Marshal(payload)

	err = suite.Handlers.HandleSearchParty(context.Background(), amqp.Delivery{
		Body:          body,
		ReplyTo:       replyQueue.Name,
		CorrelationId: "search-org-test-1",
	})
	require.NoError(t, err)

	// Receive and parse reply
	select {
	case reply := <-replies:
		var results []map[string]interface{}
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

	// Create both Individual and Organization
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

	// Create reply queue for RPC
	replyQueue, err := suite.channel.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	replies, err := suite.channel.Consume(replyQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	// Search all parties (no filter)
	payload := SearchPartyPayload{}
	body, _ := json.Marshal(payload)

	err = suite.Handlers.HandleSearchParty(context.Background(), amqp.Delivery{
		Body:          body,
		ReplyTo:       replyQueue.Name,
		CorrelationId: "search-mixed-test",
	})
	require.NoError(t, err)

	// Receive and parse reply
	select {
	case reply := <-replies:
		var results []map[string]interface{}
		err := json.Unmarshal(reply.Body, &results)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(results), 2, "Expected at least 2 results")

		// Find our test parties and verify they have complete data
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
	payload := CreatePartyPayload{
		Type: "Individual",
		Individual: &CreateIndividualPayload{
			ID:         "audit-ind-1",
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

	// Verify Audit Log
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

	// Start Listener
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := suite.Listener.Start(ctx, suite.Handlers)
		if err != nil && err != context.Canceled {
			log.Printf("Listener stopped with error: %v", err)
		}
	}()

	// Wait for listener to be ready (declaration of queues/exchanges)
	time.Sleep(100 * time.Millisecond)

	// Publish Command to Exchange
	// Fix: Use flat map for TMF polymorphism
	payload := map[string]interface{}{
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

	// Verify DB (Retry a few times as it's async)
	var saved *domain.Individual
	for i := 0; i < 10; i++ {
		saved, err = suite.Repo.GetIndividual(context.Background(), "route-ind-1")
		if err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	require.NoError(t, err, "Failed to find routed individual in DB")
	assert.Equal(t, "Routed", saved.GivenName)
}

// Regression Test for Bug 1: Header Propagation
func TestIntegration_HeaderPropagation(t *testing.T) {
	suite := setupTestSuite(t)

	// Setup: Mock Exchange/Queue to catch published event
	ch, err := suite.Conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	err = ch.QueueBind(q.Name, EvtPartyDeletionInitiated, EventExchange, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	require.NoError(t, err)

	// Execute: HandleDelete with Auth Headers
	ctx := context.Background()
	// Simulate headers coming from AMQP
	headers := amqp.Table{
		"user":          "test-user",
		"Authorization": "Bearer test-token",
	}

	// Create a party first
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

	// Verify: Headers propagated to Event
	select {
	case msg := <-msgs:
		assert.Equal(t, "test-user", msg.Headers["user"])
		assert.Equal(t, "Bearer test-token", msg.Headers["Authorization"])
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

// Regression Test for Bug 2: Stuck Deletion / Idempotency
func TestIntegration_DeleteParty_Idempotency(t *testing.T) {
	suite := setupTestSuite(t)

	// Setup: Party already in DeletionPending
	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "idemp-test-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "DeletionPending", // Already pending
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Stuck",
		FamilyName: "Delete",
	}
	require.NoError(t, suite.Repo.CreateIndividual(context.Background(), ind))

	// Setup: Reply Queue
	ch, err := suite.Conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	replyQ, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(replyQ.Name, "", true, true, false, false, nil)
	require.NoError(t, err)

	// Execute: Call HandleDeleteParty again
	payload := DeletePartyPayload{ID: "idemp-test-1"}
	body, _ := json.Marshal(payload)

	err = suite.Handlers.HandleDeleteParty(context.Background(), amqp.Delivery{
		Body:          body,
		ReplyTo:       replyQ.Name,
		CorrelationId: "corr-idemp-1",
	})
	require.NoError(t, err)

	// Verify: Should receive success reply (idempotent)
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
