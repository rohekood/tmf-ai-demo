package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"tmf/pkg/rabbitmq"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	consumer, err := rabbitmq.NewConsumer(url, "ex.domain.order", "q.mock.billing")
	if err != nil {
		logger.Error("Failed to create consumer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error("Failed to close consumer", "error", err)
		}
	}()

	publisher, err := rabbitmq.NewPublisher(url)
	if err != nil {
		logger.Error("Failed to create publisher", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			logger.Error("Failed to close publisher", "error", err)
		}
	}()
	if err := publisher.DeclareTopicExchange("ex.domain.order", true, false, false, false); err != nil {
		logger.Error("Failed to declare exchange", "error", err)
		os.Exit(1)
	}

	handler := func(ctx context.Context, payload []byte) error {
		var cmd struct {
			SagaID string  `json:"sagaId"`
			Amount float64 `json:"amount"`
		}
		if err := json.Unmarshal(payload, &cmd); err != nil {
			return err
		}

		topic := rabbitmq.EvtPaymentTransactionAuthorized
		if cmd.Amount > 1000 {
			topic = rabbitmq.EvtPaymentTransactionDeclined
		}

		evt := map[string]string{"sagaId": cmd.SagaID}

		logger.Info("Processing Payment", "sagaId", cmd.SagaID, "amount", cmd.Amount, "outcome", topic)

		return publisher.Publish(ctx, "ex.domain.order", topic, evt)
	}

	if err := consumer.Subscribe(rabbitmq.CmdPaymentTransactionAuthorize, handler); err != nil {
		logger.Error("Failed to subscribe", "error", err)
		os.Exit(1)
	}

	logger.Info("Mock Billing Started")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
