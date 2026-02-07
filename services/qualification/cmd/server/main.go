package main

import (
	"database/sql"
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
	// 0. Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Starting Qualification Service...")

	// 1. Config
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/tmf_qualification_db?sslmode=disable"
	}

	// 3. Initialize Shared Publisher for Events (Exchange: ex.domain.market)
	// Using ex.domain.market as per EDA plan
	rmqPub, err := rabbitmq.NewPublisher(rabbitURL)
	if err != nil {
		logger.Error("Failed to create Publisher", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := rmqPub.Close(); err != nil {
			logger.Error("Failed to close publisher", "error", err)
		}
	}()
	// Declare exchange explicitly
	if err := rmqPub.DeclareTopicExchange("ex.domain.market", true, false, false, false); err != nil {
		logger.Error("Failed to declare exchange", "error", err)
		os.Exit(1)
	}
	// Declare dependency exchanges to ensure they exist before RPC clients use them
	if err := rmqPub.DeclareTopicExchange("catalog_events", true, false, false, false); err != nil {
		logger.Error("Failed to declare catalog_events", "error", err)
		os.Exit(1)
	}
	if err := rmqPub.DeclareTopicExchange("customer.events", true, false, false, false); err != nil {
		logger.Error("Failed to declare customer.events", "error", err)
		os.Exit(1)
	}

	eventPub := publisher.NewEventPublisher(rmqPub, "ex.domain.market")

	// 2. Database Migrations
	runMigrations(dbURL, logger)

	// 3. Database Connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database", "error", err)
		}
	}()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	logger.Info("Connected to PostgreSQL")

	// Initialize session repository
	sessionRepo := postgres.NewSessionRepository(db)

	// 4. Initialize Infrastructure Adapters
	// 4.1 Redis Cache
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisClient, err := cache.NewRedisClient(redisAddr, redisPassword, 0)
	if err != nil {
		logger.Warn("Redis not available, continuing without cache", "error", err)
	} else {
		defer func() {
			if err := redisClient.Close(); err != nil {
				logger.Error("Failed to close redis client", "error", err)
			}
		}()
		logger.Info("Connected to Redis", "addr", redisAddr)
	}

	// 4.2 Clients
	// Use REAL RPC clients (via pkg/rabbitmq/rpc.go) instead of Mocks
	// Need to create RPC Client from pkg
	// FIX: Must specify Exchange, otherwise it uses Default Exchange and routing fails
	rpcClient, err := rabbitmq.NewRPCClient(rabbitURL, rabbitmq.WithExchange("ex.domain.market"))
	if err != nil {
		logger.Error("Failed to create RPC client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := rpcClient.Close(); err != nil {
			logger.Error("Failed to close RPC client", "error", err)
		}
	}()

	// Use real RabbitMQ RPC clients
	gisClient := rpc.NewGISClient(rpcClient)
	if redisClient != nil {
		// Wrap with Cache
		gisClient = rpc.NewCachedGISClient(gisClient, redisClient.Client, logger)
	}

	invClient := rpc.NewInventoryClient(rpcClient)

	// Real Catalog Client
	catClient, err := rpc.NewCatalogRPCClient(rabbitURL)
	if err != nil {
		logger.Error("Failed to create catalog RPC client", "error", err)
		os.Exit(1)
	}

	// Create dedicated RPC client for Catalog Pricing (must use catalog_events exchange)
	catalogPricingRPCClient, err := rabbitmq.NewRPCClient(rabbitURL, rabbitmq.WithExchange("catalog_events"))
	if err != nil {
		logger.Error("Failed to create catalog pricing RPC client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := catalogPricingRPCClient.Close(); err != nil {
			logger.Error("Failed to close catalog pricing RPC client", "error", err)
		}
	}()

	// Create dedicated RPC client for Customer Pricing (must use customer.events exchange)
	customerPricingRPCClient, err := rabbitmq.NewRPCClient(rabbitURL, rabbitmq.WithExchange("customer.events"))
	if err != nil {
		logger.Error("Failed to create customer pricing RPC client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := customerPricingRPCClient.Close(); err != nil {
			logger.Error("Failed to close customer pricing RPC client", "error", err)
		}
	}()

	// Pricing clients for session creation
	catalogPricingClient := rpc.NewCatalogPricingClient(catalogPricingRPCClient)
	customerPricingClient := rpc.NewCustomerPricingClient(customerPricingRPCClient)

	// 5. Initialize UseCase
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

	// 6. Initialize Handler
	h := handler.NewRabbitMQHandler(checkUC, logger)

	// 6.1 Initialize RPC Handler for session queries
	rpcHandler := handler.NewRPCHandler(sessionRepo, rmqPub, logger)

	// 7. Start Command Consumer
	// Exchange: ex.domain.market
	// Queue: q.qual.command
	commandConsumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.market", "q.qual.command")
	if err != nil {
		logger.Error("Failed to create command Consumer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := commandConsumer.Close(); err != nil {
			logger.Error("Failed to close command consumer", "error", err)
		}
	}()

	err = commandConsumer.Subscribe(rabbitmq.CmdQualEligibilityCheck, h.HandleCheckCommand)
	if err != nil {
		logger.Error("Failed to subscribe command", "error", err)
		os.Exit(1)
	}

	// 7.1 Start RPC Consumer
	// Queue: q.qual.rpc
	rpcConsumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.market", "q.qual.rpc")
	if err != nil {
		logger.Error("Failed to create RPC Consumer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := rpcConsumer.Close(); err != nil {
			logger.Error("Failed to close RPC consumer", "error", err)
		}
	}()

	// Bind RPC handlers for session queries
	if err := rpcHandler.BindRPCHandlers(rpcConsumer); err != nil {
		logger.Error("Failed to bind RPC handlers", "error", err)
		os.Exit(1)
	}

	logger.Info("Qualification Service Ready. Waiting for commands...")

	// 8. Wait for SigTerm
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down...")
}

func runMigrations(dbURL string, logger *slog.Logger) {
	m, err := migrate.New(
		"file://internal/infrastructure/postgres/migrations",
		dbURL,
	)
	if err != nil {
		logger.Error("failed to create migration instance", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Migrations completed successfully.")
}
