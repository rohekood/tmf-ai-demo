package worker_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tmf/services/product-catalog-management/internal/adapter/repository"
)

var (
	sharedDB    *gorm.DB
	pgContainer *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start Postgres container
	pg, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}
	pgContainer = pg

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres connection string: %v", err)
	}

	// 2. Connect to DB
	sharedDB, err = gorm.Open(gormPostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	// 3. Migrate Schema
	err = sharedDB.AutoMigrate(
		&repository.OutboxEventModel{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// 4. Run Tests
	code := m.Run()

	// 5. Cleanup
	if err := pg.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}

	os.Exit(code)
}
