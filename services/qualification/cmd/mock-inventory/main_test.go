package main

import (
	"context"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		// Connect and publish a test message
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err == nil {
			defer func() { _ = conn.Close() }()
			ch, err := conn.Channel()
			if err == nil {
				defer func() { _ = ch.Close() }()
				_ = ch.Publish("ex.domain.market", "query.inventory.resource.capacity", false, false, amqp.Publishing{
					ReplyTo: "dummy_reply_q",
					Body:    []byte(`{"address":{"city":"Nowhere"}}`),
				})
			}
		}
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := run(ctx, func(k string) string { return "amqp://guest:guest@localhost:5672/" })
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRunError(t *testing.T) {
	err := run(context.Background(), func(k string) string { return "amqp://invalid" })
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMainExecution(t *testing.T) {
	_ = os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	go func() {
		time.Sleep(500 * time.Millisecond)
		p, err := os.FindProcess(os.Getpid())
		if err == nil {
			_ = p.Signal(os.Interrupt)
		}
	}()

	main()
}
