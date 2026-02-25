package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"os/exec"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMainApp_Success(t *testing.T) {
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
		t.Fatalf("failed to start postgres: %v", err)
	}
	defer pg.Terminate(ctx)

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	// 2. Start RabbitMQ container
	rmq, err := rabbitmq.Run(ctx,
		"rabbitmq:3.12-management",
		rabbitmq.WithAdminPassword("guest"),
		rabbitmq.WithAdminUsername("guest"),
	)
	if err != nil {
		t.Fatalf("failed to start rabbitmq: %v", err)
	}
	defer rmq.Terminate(ctx)

	amqpURL, err := rmq.AmqpURL(ctx)
	if err != nil {
		t.Fatalf("failed to get amqp url: %v", err)
	}

	// Set env vars for main()
	os.Setenv("POSTGRES_URL", connStr)
	os.Setenv("RABBITMQ_URL", amqpURL)

	// We need to run migrations. `main.go` uses `file://internal/adapter/repository/migrations`
	// but when running from `cmd/`, the relative path is `../internal/adapter/repository/migrations`.
	// Let's copy or set the working directory to `..` so the path works, or temporarily create a symlink.
	os.Chdir("..")

	// Run main in a goroutine
	go func() {
		main()
	}()

	// Wait a bit to let it start and do migrations
	time.Sleep(5 * time.Second)

	// Send an interrupt to stop it gracefully
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find process: %v", err)
	}
	
	if err := p.Signal(syscall.SIGTERM); err != nil {
		t.Logf("Failed to send SIGTERM: %v", err)
	}
	
	// Wait a bit for shutdown
	time.Sleep(2 * time.Second)
}



func TestMainApp_DBFailure(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		os.Setenv("POSTGRES_URL", "host=invalid")
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainApp_DBFailure")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestMainApp_RabbitMQFailure(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "2" {
		os.Setenv("RABBITMQ_URL", "amqp://invalid")
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainApp_RabbitMQFailure")
	// Copy POSTGRES_URL to test env so it passes DB
	cmd.Env = append(os.Environ(), "BE_CRASHER=2", "POSTGRES_URL="+os.Getenv("POSTGRES_URL"))
	err := cmd.Run()
	if err != nil {
		return
	}
	t.Fatalf("process ran with success, want exit status 1")
}

func TestMainApp_MigrationFailure(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "3" {
		// Valid postgres, but invalid migrations path
		os.Chdir("/tmp") 
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainApp_MigrationFailure")
	cmd.Env = append(os.Environ(), "BE_CRASHER=3", "POSTGRES_URL="+os.Getenv("POSTGRES_URL"))
	err := cmd.Run()
	if err != nil {
		return
	}
	t.Fatalf("process ran with success, want exit status 1")
}

