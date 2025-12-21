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
	sharedConn.Close()
	pgInstance.Terminate(ctx)
	rabbitInstance.Terminate(ctx)

	os.Exit(code)
}

// setupTestSuite creates a per-test suite using shared containers
// It creates a fresh event queue for each test to avoid event cross-pollution
func setupTestSuite(t *testing.T) *IntegrationTestSuite {
	// Create a new channel for this test's event queue
	ch, err := sharedConn.Channel()
	require.NoError(t, err)

	t.Cleanup(func() {
		ch.Close()
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

	payload := CreatePartyPayload{
		Type: "Individual",
		Individual: &CreateIndividualPayload{
			ID:         "int-ind-1",
			GivenName:  "Integration",
			FamilyName: "Test",
			Href:       "http://example.com/int-ind-1",
		},
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleCreateParty(amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB
	saved, err := suite.Repo.GetIndividual("int-ind-1")
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

	payload := CreatePartyPayload{
		Type: "Organization",
		Organization: &CreateOrganizationPayload{
			ID:            "int-org-1",
			TradingName:   "IntegrationCorp",
			IsLegalEntity: true,
			Href:          "http://example.com/int-org-1",
		},
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleCreateParty(amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB
	saved, err := suite.Repo.GetOrganization("int-org-1")
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
	require.NoError(t, suite.Repo.CreateIndividual(ind))

	// Update via handler
	payload := UpdatePartyPayload{
		ID:     "int-upd-ind-1",
		Type:   "Individual",
		Status: "Active",
		Individual: &CreateIndividualPayload{
			ID:         "int-upd-ind-1",
			GivenName:  "Updated",
			FamilyName: "Person",
		},
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleUpdateParty(amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB
	updated, err := suite.Repo.GetIndividual("int-upd-ind-1")
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
	require.NoError(t, suite.Repo.CreateIndividual(ind))

	// Patch via handler
	newName := "Patched"
	newStatus := "Validated"
	payload := PatchPartyPayload{
		ID:        "int-patch-1",
		GivenName: &newName,
		Status:    &newStatus,
	}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandlePatchParty(amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB
	patched, err := suite.Repo.GetIndividual("int-patch-1")
	require.NoError(t, err)
	assert.Equal(t, "Patched", patched.GivenName)
	assert.Equal(t, "Please", patched.FamilyName) // Unchanged
	assert.Equal(t, "Validated", patched.Status)

	// Verify events
	evt1 := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt1)
	assert.Equal(t, EvtPartyUpdated, evt1.RoutingKey)
}

func TestIntegration_DeleteParty(t *testing.T) {
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
		FamilyName: "Soon",
	}
	require.NoError(t, suite.Repo.CreateIndividual(ind))

	// Delete via handler
	payload := DeletePartyPayload{ID: "int-del-1"}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleDeleteParty(amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB - should not exist
	_, err = suite.Repo.GetIndividual("int-del-1")
	assert.Error(t, err)

	// Verify event
	evt := suite.waitForEvent(t, 2*time.Second)
	require.NotNil(t, evt)
	assert.Equal(t, EvtPartyDeleted, evt.RoutingKey)
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
	require.NoError(t, suite.Repo.CreateIndividual(ind))

	// Query via handler (no ReplyTo, just verify no error)
	payload := GetPartyPayload{ID: "int-get-1"}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleGetParty(amqp.Delivery{Body: body})
	require.NoError(t, err)
}

func TestIntegration_SearchParty(t *testing.T) {
	suite := setupTestSuite(t)

	// Create test data
	ind1 := &domain.Individual{
		Party: domain.Party{
			ID:        "int-search-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "SearchAlice",
		FamilyName: "Test",
	}
	ind2 := &domain.Individual{
		Party: domain.Party{
			ID:        "int-search-2",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "SearchBob",
		FamilyName: "Test",
	}
	require.NoError(t, suite.Repo.CreateIndividual(ind1))
	require.NoError(t, suite.Repo.CreateIndividual(ind2))

	// Search via handler (no ReplyTo, just verify no error)
	givenName := "SearchAlice"
	payload := SearchPartyPayload{GivenName: &givenName}
	body, _ := json.Marshal(payload)

	err := suite.Handlers.HandleSearchParty(amqp.Delivery{Body: body})
	require.NoError(t, err)
}
