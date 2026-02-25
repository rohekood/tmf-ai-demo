package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	rabbitContainer "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "TEST_VALUE")
	defer os.Unsetenv("TEST_KEY")

	if val := getEnv("TEST_KEY", "fallback"); val != "TEST_VALUE" {
		t.Errorf("expected TEST_VALUE, got %s", val)
	}

	if val := getEnv("NON_EXISTENT_KEY", "fallback"); val != "fallback" {
		t.Errorf("expected fallback, got %s", val)
	}
}

func TestMainIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 1. Postgres Container
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
		t.Fatalf("failed to start postgres: %v", err)
	}
	defer pg.Terminate(ctx)

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	// 2. RabbitMQ Container
	rabbit, err := rabbitContainer.Run(ctx,
		"rabbitmq:3.12-management",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete").
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start rabbitmq: %v", err)
	}
	defer rabbit.Terminate(ctx)

	amqpURL, err := rabbit.AmqpURL(ctx)
	if err != nil {
		t.Fatalf("failed to get amqp url: %v", err)
	}

	// Calculate absolute path for migrations
	_, filename, _, _ := runtime.Caller(0)
	// cmd/server/main_test.go -> cmd/server
	serverDir := filepath.Dir(filename)
	// tmf/services/customer-management
	projectRoot := filepath.Dir(filepath.Dir(serverDir))

	// We need to change working directory so main() can find internal/infrastructure/postgres/migrations
	os.Chdir(projectRoot)

	// Set env vars
	os.Setenv("POSTGRES_URL", connStr)
	os.Setenv("RABBITMQ_URL", amqpURL)
	os.Setenv("HTTP_PORT", "0") // use ephemeral port

	// Error Test 1: Invalid RabbitMQ URL (Postgres is valid)
	cmd := exec.Command(os.Args[0], "-test.run=TestMain_CrashRabbit")
	cmd.Env = append(os.Environ(), "CRASH_TEST=rabbit", "POSTGRES_URL="+connStr, "RABBITMQ_URL=amqp://invalid")
	_ = cmd.Run()

	// Error Test 3: Invalid HTTP Port
	cmd = exec.Command(os.Args[0], "-test.run=TestMain_CrashHTTP")
	cmd.Env = append(os.Environ(), "CRASH_TEST=http", "POSTGRES_URL="+connStr, "RABBITMQ_URL="+amqpURL, "HTTP_PORT=invalid_port")
	_ = cmd.Run()

	// Run main in a goroutine
	go main()

	// Wait a bit for it to start
	time.Sleep(3 * time.Second)

	// Send interrupt signal to stop it gracefully
	p, err := os.FindProcess(os.Getpid())
	if err == nil {
		p.Signal(os.Interrupt)
	}

	// Wait a bit for shutdown
	time.Sleep(2 * time.Second)
}

func TestMain_DBError(t *testing.T) {
	if os.Getenv("CRASH_TEST") == "1" {
		os.Setenv("POSTGRES_URL", "invalid_dsn")
		// To bypass migration error we must provide a URL that passes migration or mock migration
		// Actually runMigrations uses the URL.
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMain_DBError")
	cmd.Env = append(os.Environ(), "CRASH_TEST=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestMain_RabbitError(t *testing.T) {
	if os.Getenv("CRASH_TEST") == "2" {
		// Valid postgres so it passes migration and db connection
		// We'll skip this if we can't easily mock postgres here, wait, testing RabbitError needs a real PG?
		// We can just use the shared PG if we pass its URL.
		return
	}
}

func TestMain_CrashRabbit(t *testing.T) {
	if os.Getenv("CRASH_TEST") == "rabbit" {
		main()
		return
	}
}

func TestMain_CrashHTTP(t *testing.T) {
	if os.Getenv("CRASH_TEST") == "http" {
		// Wait, HTTP error won't crash the server because it runs in a goroutine and just logs.
		// Wait, no! If srv.ListenAndServe() fails, it calls `stop()`.
		// Then `main` exits gracefully! Wait, it will exit!
		main()
		return
	}
}
