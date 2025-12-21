package rabbitmq

import (
	"context"
	"encoding/json"
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

// IntegrationTestSuite holds all resources for integration tests
type IntegrationTestSuite struct {
	DB        *gorm.DB
	Repo      *postgres.PartyRepository
	RabbitURL string
	Conn      *amqp.Connection
	Publisher *infraRabbit.Publisher
	Listener  *Listener
	Handlers  *Handlers
	EventChan <-chan amqp.Delivery
}

func setupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	ctx := context.Background()

	// 1. Start Postgres container
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
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %s", err)
		}
	})

	pgConnStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Run migrations
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "infrastructure", "postgres", "migrations")

	m, err := migrate.New("file://"+migrationsPath, pgConnStr)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	db, err := gorm.Open(gormPostgres.Open(pgConnStr), &gorm.Config{})
	require.NoError(t, err)

	repo := postgres.NewPartyRepository(db)

	// 2. Start RabbitMQ container
	rabbit, err := rabbitContainer.Run(ctx,
		"rabbitmq:3-management",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete").
				WithStartupTimeout(60*time.Second)),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := rabbit.Terminate(ctx); err != nil {
			t.Logf("failed to terminate rabbitmq container: %s", err)
		}
	})

	rabbitURL, err := rabbit.AmqpURL(ctx)
	require.NoError(t, err)

	// Connect to RabbitMQ
	var conn *amqp.Connection
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.NoError(t, err)

	t.Cleanup(func() {
		conn.Close()
	})

	// Create Publisher
	publisher, err := infraRabbit.NewPublisher(conn)
	require.NoError(t, err)

	// Declare exchange
	ch, _ := publisher.GetChannel()
	err = ch.ExchangeDeclare(EventExchange, "topic", true, false, false, false, nil)
	require.NoError(t, err)

	// Create Listener
	listener, err := NewListener(conn)
	require.NoError(t, err)

	// Create Handlers
	handlers := NewHandlers(repo, publisher)

	// Setup event consumer to verify events
	eventQueue, err := ch.QueueDeclare("test.events", false, true, false, false, nil)
	require.NoError(t, err)
	err = ch.QueueBind(eventQueue.Name, "evt.party.*", EventExchange, false, nil)
	require.NoError(t, err)
	events, err := ch.Consume(eventQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	return &IntegrationTestSuite{
		DB:        db,
		Repo:      repo,
		RabbitURL: rabbitURL,
		Conn:      conn,
		Publisher: publisher,
		Listener:  listener,
		Handlers:  handlers,
		EventChan: events,
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
	suite := setupIntegrationTest(t)

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
	suite := setupIntegrationTest(t)

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
	suite := setupIntegrationTest(t)

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
	suite := setupIntegrationTest(t)

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
	suite := setupIntegrationTest(t)

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
	suite := setupIntegrationTest(t)

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
	suite := setupIntegrationTest(t)

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
