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
	eventPub := publisher.NewEventPublisher(rmqPub, "ex.domain.market")

	// 2. Database Connection
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
	catClient := rpc.NewMockCatalogClient() // Catalog still mocked as it's not critical for this E2E flow yet or not implemented

	// Pricing clients for session creation
	catalogPricingClient := rpc.NewCatalogPricingClient(rpcClient)
	customerPricingClient := rpc.NewCustomerPricingClient(rpcClient)

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

	// 7. Start Consumer
	// Exchange: ex.domain.market
	// Queue: q.qual.command
	// Binding: cmd.qual.eligibility.check
	consumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.market", "q.qual.command")
	if err != nil {
		logger.Error("Failed to create Consumer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error("Failed to close consumer", "error", err)
		}
	}()

	err = consumer.Subscribe(rabbitmq.CmdQualEligibilityCheck, h.HandleCheckCommand)
	if err != nil {
		logger.Error("Failed to subscribe", "error", err)
		os.Exit(1)
	}

	// Bind RPC handlers for session queries
	if err := rpcHandler.BindRPCHandlers(consumer); err != nil {
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
