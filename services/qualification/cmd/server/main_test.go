package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRun_Success(t *testing.T) {
	_ = os.Chdir("../../")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		t.Skipf("Skipping integration test (testcontainers error): %v", err)
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
		t.Skipf("Skipping integration test (testcontainers error): %v", err)
	}
	defer func() { _ = rmq.Terminate(ctx) }()

	amqpURL, err := rmq.AmqpURL(ctx)
	if err != nil {
		t.Fatalf("failed to get amqp url: %v", err)
	}

	_ = os.Setenv("POSTGRES_URL", connStr)
	_ = os.Setenv("RABBITMQ_URL", amqpURL)

	go func() {
		time.Sleep(3 * time.Second)
		cancel()
	}()

	err = run(ctx, os.Getenv)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRun_BadDBURL(t *testing.T) {
	err := run(context.Background(), func(k string) string {
		if k == "POSTGRES_URL" {
			return "postgres://invalid:invalid@localhost:5432/invalid?sslmode=disable"
		}
		if k == "RABBITMQ_URL" {
			return "amqp://guest:guest@localhost:5672/"
		}
		return ""
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestRun_BadRabbitURL(t *testing.T) {
	err := run(context.Background(), func(k string) string {
		if k == "POSTGRES_URL" {
			return "postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable"
		}
		if k == "RABBITMQ_URL" {
			return "amqp://invalid:invalid@localhost:5672/"
		}
		return ""
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
