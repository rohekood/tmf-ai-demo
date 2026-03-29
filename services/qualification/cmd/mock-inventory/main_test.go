package main

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	go func() {
		time.Sleep(100 * time.Millisecond)
		conn, err := amqp.Dial(amqpURL)
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

	err = run(ctx, func(k string) string { return amqpURL })
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
