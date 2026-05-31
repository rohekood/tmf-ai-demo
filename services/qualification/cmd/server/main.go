package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"tmf/pkg/rabbitmq"
	"tmf/services/qualification/internal/adapter/cache"
	"tmf/services/qualification/internal/adapter/handler"
	"tmf/services/qualification/internal/adapter/publisher"
	"tmf/services/qualification/internal/adapter/rpc"
	"tmf/services/qualification/internal/infrastructure/postgres"
	"tmf/services/qualification/internal/usecase"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger.Info("Starting Qualification Service...")

	rabbitURL := getEnv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	dbURL := getEnv("POSTGRES_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/tmf_qualification_db?sslmode=disable"
	}

	rmqPub, err := rabbitmq.NewPublisher(rabbitURL)
	if err != nil {
		return fmt.Errorf("failed to create Publisher: %w", err)
	}
	defer func() { _ = rmqPub.Close() }()

	if err := rmqPub.DeclareTopicExchange("ex.domain.market", true, false, false, false); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	if err := rmqPub.DeclareTopicExchange("catalog_events", true, false, false, false); err != nil {
		return fmt.Errorf("failed to declare catalog_events: %w", err)
	}
	if err := rmqPub.DeclareTopicExchange("customer.events", true, false, false, false); err != nil {
		return fmt.Errorf("failed to declare customer.events: %w", err)
	}

	eventPub := publisher.NewEventPublisher(rmqPub, "ex.domain.market")

	if err := runMigrations(dbURL, logger); err != nil {
		return err
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	logger.Info("Connected to PostgreSQL")

	sessionRepo := postgres.NewSessionRepository(db)

	redisAddr := getEnv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisUser := getEnv("REDIS_USER")
	redisPassword := getEnv("REDIS_PASSWORD")

	redisClient, err := cache.NewRedisClient(redisAddr, redisUser, redisPassword)
	if err != nil {
		logger.Warn("Redis not available, continuing without cache", "error", err)
	} else {
		defer func() { _ = redisClient.Close() }()
		logger.Info("Connected to Redis", "addr", redisAddr)
	}

	rpcClient, err := rabbitmq.NewRPCClient(rabbitURL, rabbitmq.WithExchange("ex.domain.market"))
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	defer func() { _ = rpcClient.Close() }()

	gisClient := rpc.NewGISClient(rpcClient)
	if redisClient != nil {
		gisClient = rpc.NewCachedGISClient(gisClient, redisClient.Client, logger, "qualification:")
	}

	invClient := rpc.NewInventoryClient(rpcClient)

	catalogPricingRPCClient, err := rabbitmq.NewRPCClient(rabbitURL, rabbitmq.WithExchange("catalog_events"))
	if err != nil {
		return fmt.Errorf("failed to create catalog pricing RPC client: %w", err)
	}
	defer func() { _ = catalogPricingRPCClient.Close() }()

	catClient := rpc.NewCatalogRPCClient(catalogPricingRPCClient)

	customerPricingRPCClient, err := rabbitmq.NewRPCClient(rabbitURL, rabbitmq.WithExchange("customer.events"))
	if err != nil {
		return fmt.Errorf("failed to create customer pricing RPC client: %w", err)
	}
	defer func() { _ = customerPricingRPCClient.Close() }()

	catalogPricingClient := rpc.NewCatalogPricingClient(catalogPricingRPCClient)
	customerPricingClient := rpc.NewCustomerPricingClient(customerPricingRPCClient)

	checkUC := usecase.NewCheckEligibility(
		gisClient,
		invClient,
		catClient,
		eventPub,
		sessionRepo,
		customerPricingClient,
		catalogPricingClient,
		logger,
	)

	h := handler.NewRabbitMQHandler(checkUC, logger)
	rpcHandler := handler.NewRPCHandler(sessionRepo, rmqPub, logger)

	commandConsumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.market", "q.qual.command")
	if err != nil {
		return fmt.Errorf("failed to create command Consumer: %w", err)
	}
	defer func() { _ = commandConsumer.Close() }()

	err = commandConsumer.Subscribe(rabbitmq.CmdQualEligibilityCheck, h.HandleCheckCommand)
	if err != nil {
		return fmt.Errorf("failed to subscribe command: %w", err)
	}

	rpcConsumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.market", "q.qual.rpc")
	if err != nil {
		return fmt.Errorf("failed to create RPC Consumer: %w", err)
	}
	defer func() { _ = rpcConsumer.Close() }()

	if err := rpcHandler.BindRPCHandlers(rpcConsumer); err != nil {
		return fmt.Errorf("failed to bind RPC handlers: %w", err)
	}

	logger.Info("Qualification Service Ready. Waiting for commands...")

	<-ctx.Done()
	logger.Info("Shutting down...")
	return nil
}

func runMigrations(dbURL string, logger *slog.Logger) error {
	m, err := migrate.New(
		"file://internal/infrastructure/postgres/migrations",
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	logger.Info("Migrations completed successfully.")
	return nil
}
