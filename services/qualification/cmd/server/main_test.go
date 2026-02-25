package main

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRun_Success(t *testing.T) {
	_ = os.Chdir("../../") // change to qualification root
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_ = os.Setenv("POSTGRES_URL", "postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable")
	_ = os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	err := run(ctx, os.Getenv)
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

func TestMainExecution(t *testing.T) {
	_ = os.Setenv("POSTGRES_URL", "postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable")
	_ = os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	go func() {
		time.Sleep(1 * time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err == nil {
			_ = p.Signal(os.Interrupt)
		}
	}()

	main()
}

func TestRun_BadRedisURL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := run(ctx, func(k string) string {
		if k == "POSTGRES_URL" {
			return "postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable"
		}
		if k == "RABBITMQ_URL" {
			return "amqp://guest:guest@localhost:5672/"
		}
		if k == "REDIS_ADDR" {
			return "invalid:0"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("expected error, got nil")
	}
}
