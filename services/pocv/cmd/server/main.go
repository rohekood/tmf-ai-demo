package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger.Info("Starting POCV Service...")

	// 1. Database Connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	}
	db, err := gorm.Open(gormPostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Auto Migrate
	if err := db.AutoMigrate(&repository.SagaTable{}, &repository.OutboxTable{}); err != nil {
		logger.Error("Failed to auto migrate", "error", err)
		os.Exit(1)
	}

	// 2. RabbitMQ Connection
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	// 3. Components
	// Publisher (for Worker)
	rmqPub, err := rabbitmq.NewPublisher(rabbitURL)
	if err != nil {
		logger.Error("Failed to create publisher", "error", err)
		os.Exit(1)
	}
	defer func() { _ = rmqPub.Close() }()
	// Declare explicitly
	if err := rmqPub.DeclareTopicExchange("ex.domain.order", true, false, false, false); err != nil {
		logger.Error("Failed to declare exchange", "error", err)
		os.Exit(1)
	}

	// RPC Client for Cart (Commerce Domain)
	rpcClient, err := rabbitmq.NewRPCClient(rabbitURL, rabbitmq.WithExchange("ex.domain.commerce"))
	if err != nil {
		logger.Error("Failed to create RPC client", "error", err)
		os.Exit(1)
	}
	defer func() { _ = rpcClient.Close() }()

	// Layers
	repo := repository.NewSagaRepository(db)
	cartClient := rpc.NewCartClient(rpcClient)
	uc := usecase.NewSagaUseCase(repo, cartClient)
	h := handler.NewRabbitMQHandler(uc)

	// 4. Outbox Worker
	outboxWorker := worker.NewOutboxWorker(db, rmqPub, "ex.domain.order")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go outboxWorker.Start(ctx)

	// 5. Consumers
	// Consumer for Saga Events
	consumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.order", "q.pocv.saga")
	if err != nil {
		logger.Error("Failed to create consumer", "error", err)
		os.Exit(1)
	}
	defer func() { _ = consumer.Close() }()

	// Subscriptions
	// Subscriptions
	// Use Dispatcher Pattern to avoid Multiple Consumers on same Queue
	slog.Info("Subscribing to all events with Dispatcher")
	if err := consumer.Subscribe("#", h.HandleSagaEvent); err != nil {
		logger.Error("Failed to subscribe dispatcher", "error", err)
		os.Exit(1)
	}

	logger.Info("POCV Service Started")

	// 6. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down POCV Service...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	<-shutdownCtx.Done()
	logger.Info("POCV Service Stopped.")
}
