package postgres

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	sharedDB      *gorm.DB
	sharedConnStr string
	pgContainer   *postgres.PostgresContainer
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
		// Note: To reuse across test runs (not just within package), we would need 'testcontainers.WithReuse(true)'
		// and a fixed container name. For now, we reuse within the package.
	)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}
	pgContainer = pg

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get postgres connection string: %v", err)
	}
	sharedConnStr = connStr

	// Run migrations
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "migrations")

	mig, err := migrate.New("file://"+migrationsPath, connStr)
	if err != nil {
		log.Fatalf("failed to create migrate: %v", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}

	sharedDB, err = gorm.Open(gormPostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	code := m.Run()

	// Cleanup
	if err := pg.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}

	os.Exit(code)
}
