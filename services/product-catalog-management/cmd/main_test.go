package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"os/exec"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMainApp_Success(t *testing.T) {
	ctx := context.Background()

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
	defer func() { _ = pg.Terminate(ctx) }()

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	rmq, err := rabbitmq.Run(ctx,
		"rabbitmq:3.12-management",
		rabbitmq.WithAdminPassword("guest"),
		rabbitmq.WithAdminUsername("guest"),
	)
	if err != nil {
		t.Fatalf("failed to start rabbitmq: %v", err)
	}
	defer func() { _ = rmq.Terminate(ctx) }()

	amqpURL, err := rmq.AmqpURL(ctx)
	if err != nil {
		t.Fatalf("failed to get amqp url: %v", err)
	}

	_ = os.Setenv("POSTGRES_URL", connStr)
	_ = os.Setenv("RABBITMQ_URL", amqpURL)

	_ = os.Chdir("..")

	go func() {
		main()
	}()

	time.Sleep(5 * time.Second)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find process: %v", err)
	}

	if err := p.Signal(syscall.SIGTERM); err != nil {
		t.Logf("Failed to send SIGTERM: %v", err)
	}

	time.Sleep(2 * time.Second)
}

func TestMainApp_DBFailure(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		_ = os.Setenv("POSTGRES_URL", "host=invalid")
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
		_ = os.Setenv("RABBITMQ_URL", "amqp://invalid")
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainApp_RabbitMQFailure")
	cmd.Env = append(os.Environ(), "BE_CRASHER=2", "POSTGRES_URL="+os.Getenv("POSTGRES_URL"))
	err := cmd.Run()
	if err != nil {
		return
	}
	t.Fatalf("process ran with success, want exit status 1")
}

func TestMainApp_MigrationFailure(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "3" {
		_ = os.Chdir("/tmp")
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
