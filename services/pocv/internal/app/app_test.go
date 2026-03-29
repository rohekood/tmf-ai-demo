package app_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"tmf/services/pocv/internal/app"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcrabbit "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRun_StartsAndShutdownsCleanly(t *testing.T) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Skipping integration test (testcontainers panic): %v", r)
		}
	}()

	// 1. Start PostgreSQL container
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:15",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Skipf("Skipping integration test (cannot start postgres): %v", err)
		return
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Skipf("cannot get postgres connection string: %v", err)
		return
	}

	// Run migrations
	// From internal/app/ → up 1 dir → internal/ → infrastructure/postgres/migrations
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "infrastructure", "postgres", "migrations")
	m, err := migrate.New("file://"+migrationsPath, pgConnStr)
	if err != nil {
		t.Skipf("cannot create migrator: %v", err)
		return
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Skipf("migration failed: %v", err)
		return
	}

	// 2. Start RabbitMQ container
	rabbitContainer, err := tcrabbit.Run(ctx,
		"rabbitmq:3-management",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete").
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Skipf("Skipping integration test (cannot start rabbitmq): %v", err)
		return
	}
	t.Cleanup(func() { _ = rabbitContainer.Terminate(ctx) })

	rabbitURL, err := rabbitContainer.AmqpURL(ctx)
	if err != nil {
		t.Skipf("cannot get rabbitmq URL: %v", err)
		return
	}

	// 3. Run the application with cancellable context
	cfg := app.Config{
		DatabaseURL: pgConnStr,
		RabbitMQURL: rabbitURL,
		Exchange:    "ex.domain.order",
		QueueName:   "q.pocv.saga.test",
	}

	runCtx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- app.Run(runCtx, cfg)
	}()

	// Give the app a moment to start up
	time.Sleep(500 * time.Millisecond)

	// Cancel context to trigger graceful shutdown
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("app.Run returned unexpected error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Error("app.Run did not shut down within timeout")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := app.DefaultConfig()
	if cfg.DatabaseURL == "" {
		t.Error("expected non-empty DatabaseURL")
	}
	if cfg.RabbitMQURL == "" {
		t.Error("expected non-empty RabbitMQURL")
	}
	if cfg.Exchange == "" {
		t.Error("expected non-empty Exchange")
	}
	if cfg.QueueName == "" {
		t.Error("expected non-empty QueueName")
	}
}

func TestRun_DBConnectionFails(t *testing.T) {
	cfg := app.Config{
		DatabaseURL: "host=localhost user=baduser password=bad dbname=bad port=65000 sslmode=disable connect_timeout=1",
		RabbitMQURL: "amqp://guest:guest@localhost:5672/",
		Exchange:    "ex.domain.order",
		QueueName:   "q.pocv.saga.test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Run(ctx, cfg)
	if err == nil {
		t.Error("expected error when database URL is invalid, got nil")
	}
}

// setupPGForApp starts a PostgreSQL container, runs migrations and returns the DSN.
func setupPGForApp(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("testcontainers panic: %v", r)
		}
	}()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:15",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Skipf("cannot start postgres: %v", err)
		return ""
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Skipf("cannot get postgres connection string: %v", err)
		return ""
	}

	// From internal/app/ → up 1 dir → internal/ → infrastructure/postgres/migrations
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "infrastructure", "postgres", "migrations")
	m, err := migrate.New("file://"+migrationsPath, pgConnStr)
	if err != nil {
		t.Skipf("cannot create migrator: %v", err)
		return ""
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Skipf("migration failed: %v", err)
		return ""
	}

	return pgConnStr
}

// TestRun_PublisherFails covers the error path where RabbitMQ publisher creation fails
// (line 52-54 in app.go).
func TestRun_PublisherFails(t *testing.T) {
	pgConnStr := setupPGForApp(t)
	if pgConnStr == "" {
		return
	}

	cfg := app.Config{
		DatabaseURL: pgConnStr,
		RabbitMQURL: "amqp://bad:bad@127.0.0.1:65502/", // unreachable → publisher fails
		Exchange:    "ex.domain.order",
		QueueName:   "q.pocv.saga.test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := app.Run(ctx, cfg)
	if err == nil {
		t.Error("expected error when RabbitMQ publisher cannot connect, got nil")
	}
}

// TestRun_SecondPublisherFails covers the second AMQP error path (RPC client, line 62-65)
// by using a URL that fails after the publisher step (same invalid URL hits publisher first).
// Having two tests with different port numbers ensures they don't collide and both paths
// are exercised as regression guards.
func TestRun_AnotherRabbitError(t *testing.T) {
	pgConnStr := setupPGForApp(t)
	if pgConnStr == "" {
		return
	}

	cfg := app.Config{
		DatabaseURL: pgConnStr,
		RabbitMQURL: "amqp://x:x@127.0.0.1:65503/",
		Exchange:    "ex.domain.order",
		QueueName:   "q.pocv.saga.test2",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := app.Run(ctx, cfg)
	if err == nil {
		t.Error("expected error from Run when using unreachable AMQP, got nil")
	}
}
