package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"tmf/services/customer-management/internal/domain"
	"tmf/services/customer-management/internal/infrastructure/postgres"
	infraRabbit "tmf/services/customer-management/internal/infrastructure/rabbitmq"

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

var (
	sharedDB        *gorm.DB
	sharedRepo      *postgres.CustomerRepository
	sharedConn      *amqp.Connection
	sharedPublisher *infraRabbit.Publisher
	pgInstance      testcontainers.Container
	rabbitInstance  testcontainers.Container
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start Postgres container
	pg, err := pgContainer.Run(ctx,
		"postgres:15",
		pgContainer.WithDatabase("testdb"),
		pgContainer.WithUsername("postgres"),
		pgContainer.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}
	pgInstance = pg

	pgConnStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres connection string: %v", err)
	}

	// Run migrations
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "infrastructure", "postgres", "migrations")

	mig, err := migrate.New("file://"+migrationsPath, pgConnStr)
	if err != nil {
		log.Fatalf("failed to create migrate: %v", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}

	sharedDB, err = gorm.Open(gormPostgres.Open(pgConnStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	sharedRepo = postgres.NewCustomerRepository(sharedDB)

	// 2. Start RabbitMQ container
	rabbit, err := rabbitContainer.Run(ctx,
		"rabbitmq:3-management",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete").
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start rabbitmq: %v", err)
	}
	rabbitInstance = rabbit

	rabbitURL, err := rabbit.AmqpURL(ctx)
	if err != nil {
		log.Fatalf("failed to get rabbitmq URL: %v", err)
	}

	sharedConn, err = amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	sharedPublisher, err = infraRabbit.NewPublisher(sharedConn)
	if err != nil {
		log.Fatalf("failed to create publisher: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	sharedConn.Close()
	pgInstance.Terminate(ctx)
	rabbitInstance.Terminate(ctx)

	os.Exit(code)
}

func TestIntegration_OnboardCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher)

	payload := OnboardCustomerPayload{
		ID:        "cust-1",
		Name:      "Test Customer",
		PartyID:   "party-1",
		PartyType: "Individual",
		Accounts: []CustomerAccountDTO{
			{ID: "acc-1", Name: "Main Account", AccountStatus: "Active", AccountType: "Billing"},
		},
	}
	body, _ := json.Marshal(payload)

	err := handlers.HandleOnboardCustomer(ctx, amqp.Delivery{
		Body:     body,
		Exchange: "tmf.events", // Test using an event exchange for publication
	})
	require.NoError(t, err)

	// Verify DB
	saved, err := sharedRepo.GetCustomer(ctx, "cust-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Customer", saved.Name)
	assert.Equal(t, domain.CustomerStatusActive, saved.Status)
	assert.Len(t, saved.Accounts, 1)
	assert.Equal(t, "Main Account", saved.Accounts[0].Name)
}

func TestIntegration_UpdateCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher)

	// Create first
	cust := &domain.Customer{
		ID:      "cust-upd-1",
		Name:    "Original Name",
		Status:  domain.CustomerStatusActive,
		PartyID: "party-upd-1",
	}
	require.NoError(t, sharedRepo.CreateCustomer(ctx, cust))

	// Update
	payload := UpdateCustomerPayload{
		ID:     "cust-upd-1",
		Name:   "Updated Name",
		Status: domain.CustomerStatusSuspended,
	}
	body, _ := json.Marshal(payload)

	err := handlers.HandleUpdateCustomer(ctx, amqp.Delivery{Body: body})
	require.NoError(t, err)

	// Verify DB
	updated, err := sharedRepo.GetCustomer(ctx, "cust-upd-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, domain.CustomerStatusSuspended, updated.Status)
}
func TestIntegration_AuditTrail(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher)

	userID := "audit-user-123"
	payload := OnboardCustomerPayload{
		ID:   "audit-cust-1",
		Name: "Audit Test Customer",
	}
	body, _ := json.Marshal(payload)

	err := handlers.HandleOnboardCustomer(ctx, amqp.Delivery{
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
	err = sharedDB.Table("audit.logged_actions").
		Where("table_name = ? AND action = ?", "customers", "I").
		Order("action_tstamp_clk DESC").
		First(&auditLog).Error

	require.NoError(t, err)
	assert.Equal(t, userID, auditLog.UserName)
	assert.Equal(t, "customers", auditLog.TableName)
	assert.Equal(t, "I", auditLog.Action)

	// Test System User on Event
	eventPayload := PartyEventPayload{
		ID:         "party-1",
		GivenName:  "Jane",
		FamilyName: "Doe",
		Type:       "Individual",
	}
	// Pre-create customer linked to this party
	require.NoError(t, sharedRepo.CreateCustomer(ctx, &domain.Customer{
		ID:      "cust-party-1",
		Name:    "Old Name",
		PartyID: "party-1",
	}))

	evtBody, _ := json.Marshal(eventPayload)
	err = handlers.HandlePartyEvent(ctx, amqp.Delivery{
		Body:       evtBody,
		RoutingKey: EvtPartyUpdated,
	})
	require.NoError(t, err)

	var systemAudit LoggedAction
	err = sharedDB.Table("audit.logged_actions").
		Where("table_name = ? AND action = ?", "customers", "U").
		Order("action_tstamp_clk DESC").
		First(&systemAudit).Error

	require.NoError(t, err)
	assert.Equal(t, "system.customer-management", systemAudit.UserName)
}

func TestIntegration_GetCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher)

	// Pre-create customer
	custID := "get-cust-1"
	require.NoError(t, sharedRepo.CreateCustomer(ctx, &domain.Customer{
		ID:      custID,
		Name:    "Get Me",
		Status:  domain.CustomerStatusActive,
		PartyID: "p-get-1",
	}))

	// Query via handler
	payload := GetCustomerPayload{ID: custID}
	body, _ := json.Marshal(payload)

	err := handlers.HandleGetCustomer(ctx, amqp.Delivery{Body: body})
	require.NoError(t, err)
}

func TestIntegration_SearchCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher)

	// Pre-create customers
	require.NoError(t, sharedRepo.CreateCustomer(ctx, &domain.Customer{
		ID:      "s-cust-1",
		Name:    "Searchable One",
		Status:  domain.CustomerStatusActive,
		PartyID: "p-s-1",
	}))

	// Search via handler
	payload := SearchCustomerPayload{Name: "Searchable One"}
	body, _ := json.Marshal(payload)

	err := handlers.HandleSearchCustomer(ctx, amqp.Delivery{Body: body})
	require.NoError(t, err)
}

func TestIntegration_DeleteCustomer(t *testing.T) {
	ctx := context.Background()
	handlers := NewHandlers(sharedRepo, sharedPublisher)

	// Pre-create
	custID := "del-cust-int-1"
	require.NoError(t, sharedRepo.CreateCustomer(ctx, &domain.Customer{
		ID:     custID,
		Name:   "Delete Me",
		Status: domain.CustomerStatusActive,
	}))

	// Delete via handler
	payload := DeleteCustomerPayload{ID: custID}
	body, _ := json.Marshal(payload)

	err := handlers.HandleDeleteCustomer(ctx, amqp.Delivery{
		Body:     body,
		Exchange: "tmf.events",
	})
	require.NoError(t, err)

	// Verify deleted
	_, err = sharedRepo.GetCustomer(ctx, custID)
	assert.Error(t, err)
}
