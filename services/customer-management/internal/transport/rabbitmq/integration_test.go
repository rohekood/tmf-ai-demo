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
		"postgres:15-alpine",
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
