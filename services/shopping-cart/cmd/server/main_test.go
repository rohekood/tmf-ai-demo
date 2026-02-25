package main

import (
	"context"
	"os"
	"testing"
)

func TestMainExecution(t *testing.T) {
	os.Chdir("/home/raino/IdeaProject/tmf/services/shopping-cart")
	os.Setenv("DB_URL", "postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable")
	os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	// Cancel immediately so workers don't run in background and steal rows
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = run(ctx, os.Getenv)
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

func TestMain_Signal(t *testing.T) {
	os.Setenv("DB_URL", "postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable")
	os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	go func() {
		p, err := os.FindProcess(os.Getpid())
		if err == nil {
			p.Signal(os.Interrupt)
		}
	}()

	main()
}
