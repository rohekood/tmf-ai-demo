package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/pocv/internal/adapter/handler"
	"tmf/services/pocv/internal/adapter/repository"
	"tmf/services/pocv/internal/adapter/rpc"
	"tmf/services/pocv/internal/adapter/worker"
	"tmf/services/pocv/internal/usecase"

	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config holds all connection configuration for the application.
type Config struct {
	DatabaseURL string
	RabbitMQURL string
	Exchange    string
	QueueName   string
}

// DefaultConfig returns a Config with default local values.
func DefaultConfig() Config {
	return Config{
		DatabaseURL: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable",
		RabbitMQURL: "amqp://guest:guest@localhost:5672/",
		Exchange:    "ex.domain.order",
		QueueName:   "q.pocv.saga",
	}
}

// Run wires up all application components and runs until ctx is cancelled.
// Returns an error if any component fails to initialize.
func Run(ctx context.Context, cfg Config) error {
	// Database
	db, err := gorm.Open(gormPostgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(&repository.SagaTable{}, &repository.OutboxTable{}); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	// RabbitMQ Publisher
	rmqPub, err := rabbitmq.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}
	defer func() { _ = rmqPub.Close() }()

	// RPC Client
	rpcClient, err := rabbitmq.NewRPCClient(cfg.RabbitMQURL, rabbitmq.WithExchange("ex.domain.commerce"), rabbitmq.WithReplyQueue(rabbitmq.DirectReplyToQueue))
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	defer func() { _ = rpcClient.Close() }()

	// Application layers
	repo := repository.NewSagaRepository(db)
	cartClient := rpc.NewCartClient(rpcClient)
	uc := usecase.NewSagaUseCase(repo, cartClient)
	h := handler.NewRabbitMQHandler(uc, rmqPub)
	rpcHandler := handler.NewRPCHandler(uc, rmqPub, slog.Default())

	// Outbox Worker
	outboxWorker := worker.NewOutboxWorker(db, rmqPub, cfg.Exchange)
	go outboxWorker.Start(ctx)

	// Consumer
	consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQURL, cfg.Exchange, cfg.QueueName, rabbitmq.WithMessageTimeout(30*time.Second))
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer func() { _ = consumer.Close() }()

	if err := consumer.Subscribe("#", h.HandleSagaEvent); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// RPC Consumer for queries
	rpcConsumer, err := rabbitmq.NewConsumer(cfg.RabbitMQURL, cfg.Exchange, "q.pocv.rpc", rabbitmq.WithMessageTimeout(30*time.Second))
	if err != nil {
		return fmt.Errorf("failed to create rpc consumer: %w", err)
	}
	defer func() { _ = rpcConsumer.Close() }()

	if err := rpcHandler.BindRPCHandlers(rpcConsumer); err != nil {
		return fmt.Errorf("failed to bind rpc handlers: %w", err)
	}

	slog.Info("POCV Service Started")
	<-ctx.Done()
	slog.Info("POCV Service stopped")
	return nil
}
