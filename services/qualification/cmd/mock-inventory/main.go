package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"tmf/pkg/rabbitmq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Getenv); err != nil {
		slog.Error("Fatal error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getEnv func(string) string) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	url := getEnv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer func() { _ = ch.Close() }()

	// Declare Exchange
	exchange := "ex.domain.market"
	err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Declare Queue
	q, err := ch.QueueDeclare(
		"q.mock.inventory.rpc", // name
		false,                  // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return err
	}

	// Bind
	err = ch.QueueBind(q.Name, rabbitmq.QueryInventoryResourceCapacity, exchange, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	logger.Info("Mock Inventory Started. Waiting for RPC queries...")

	go func() {
		for d := range msgs {
			response := map[string]interface{}{
				"free": 100, // Return 100 ports available
			}
			body, _ := json.Marshal(response)

			err = ch.PublishWithContext(context.Background(),
				"",        // default exchange
				d.ReplyTo, // routing key = reply queue
				false,
				false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          body,
				})

			if err != nil {
				logger.Error("Failed to reply", "error", err)
				_ = d.Nack(false, false)
			} else {
				_ = d.Ack(false)
			}
		}
	}()

	<-ctx.Done()
	return nil
}
