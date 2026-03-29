package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDBURL     string
	testRabbitURL string
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../../"); err != nil {
		panic("Failed to chdir: " + err.Error())
	}

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithDatabase("cart_db"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err == nil {
		connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err == nil {
			testDBURL = connStr
		}
	}

	rmqContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.12-management-alpine",
		testcontainers.WithEnv(map[string]string{
			"RABBITMQ_DEFAULT_USER": "admin",
			"RABBITMQ_DEFAULT_PASS": "admin",
		}),
	)
	if err == nil {
		amqpURL, err := rmqContainer.AmqpURL(ctx)
		if err == nil {
			testRabbitURL = strings.Replace(amqpURL, "guest:guest", "admin:admin", 1)
		}
	}
	if testDBURL == "" || testRabbitURL == "" {
		panic("Failed to initialize testcontainers: DB or RabbitMQ URL is empty")
	}

	code := m.Run()

	if pgContainer != nil {
		_ = pgContainer.Terminate(ctx)
	}
	if rmqContainer != nil {
		_ = rmqContainer.Terminate(ctx)
	}

	os.Exit(code)
}

func TestMainExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	err := run(ctx, func(k string) string {
		if k == "DB_URL" {
			return testDBURL
		}
		if k == "RABBITMQ_URL" {
			return testRabbitURL
		}
		return ""
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRun_NoDBURL(t *testing.T) {
	err := run(context.Background(), func(k string) string { return "" })
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestRun_BadDBURL(t *testing.T) {
	err := run(context.Background(), func(k string) string {
		if k == "DB_URL" {
			return "postgres://invalid:invalid@localhost:5432/invalid?sslmode=disable"
		}
		return ""
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestRun_BadRabbitURL(t *testing.T) {
	err := run(context.Background(), func(k string) string {
		if k == "DB_URL" {
			return testDBURL
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

func TestMain_Signal(t *testing.T) {
	_ = os.Setenv("DB_URL", testDBURL)
	_ = os.Setenv("RABBITMQ_URL", testRabbitURL)

	go func() {
		time.Sleep(500 * time.Millisecond)
		p, err := os.FindProcess(os.Getpid())
		if err == nil {
			_ = p.Signal(os.Interrupt)
		}
	}()

	main()
}
