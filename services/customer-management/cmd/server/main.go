package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/customer-management/internal/infrastructure/postgres"
	"tmf/services/customer-management/internal/infrastructure/telemetry"
	transportHttp "tmf/services/customer-management/internal/transport/http"
	transportRabbit "tmf/services/customer-management/internal/transport/rabbitmq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	amqp "github.com/rabbitmq/amqp091-go"

	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("starting customer management service")

	// Initialize OpenTelemetry
	shutdown, err := telemetry.InitTracer("customer-management")
	if err != nil {
		slog.Error("failed to initialize tracer", "error", err)
	} else {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				slog.Error("Failed to shutdown", "error", err)
			}
		}()
	}

	// Configuration (using defaults or env vars)
	dbURL := getEnv("POSTGRES_URL", "postgres://postgres:password@localhost:5432/tmf_customer_db?sslmode=disable")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	// 1. Database Migrations
	runMigrations(dbURL)

	// 2. Database Connection
	db, err := gorm.Open(gormPostgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	repo := postgres.NewCustomerRepository(db)
	tm := postgres.NewTransactionManager(db)
	outboxRepo := postgres.NewOutboxRepository(db)
	eventPublisher := postgres.NewOutboxPublisher(outboxRepo)

	// 3. RabbitMQ Connection
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		slog.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("Failed to close DB connection", "error", err)
		}
	}()

	publisher, err := rabbitmq.NewPublisherWithConnection(conn)
	if err != nil {
		slog.Error("failed to create publisher", "error", err)
		os.Exit(1)
	}
	if err := publisher.DeclareTopicExchange("customer.events", true, false, false, false); err != nil {
		slog.Error("failed to declare exchange", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			slog.Error("Failed to close publisher", "error", err)
		}
	}()

	outboxWorker := postgres.NewOutboxWorker(outboxRepo, publisher, slog.Default())

	// 4. Handlers & Listener
	// 4. Handlers & Listener
	handlers := transportRabbit.NewHandlers(repo, publisher, tm, eventPublisher)
	listener, err := transportRabbit.NewListener(conn)
	if err != nil {
		slog.Error("failed to create listener", "error", err)
		os.Exit(1)
	}

	// 4.1 Initialize RPC Handler for pricing queries
	rpcHandler := transportRabbit.NewRPCHandler(repo, publisher, logger)

	// 4.2 Create RPC consumer
	rpcConsumer, err := rabbitmq.NewConsumerWithConnection(conn, "customer.events", "customer_rpc_queue")
	if err != nil {
		slog.Error("Failed to create RPC consumer", "error", err)
		os.Exit(1)
	}

	// 4.3 Bind RPC handlers
	if err := rpcHandler.BindRPCHandlers(rpcConsumer); err != nil {
		slog.Error("Failed to bind RPC handlers", "error", err)
		os.Exit(1)
	}

	// 5. Start Service
	// Create a context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go outboxWorker.Start(ctx)
	defer outboxWorker.Stop()

	go func() {
		if err := listener.Start(ctx, handlers); err != nil {
			slog.Error("listener stopped", "error", err)
			stop()
		}
	}()

	// 6. Start Health Check Server
	healthHandler := transportHttp.NewHealthHandler(db, conn)
	metricsHandler := transportHttp.MetricsHandler()

	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/metrics", metricsHandler)

	port := getEnv("HTTP_PORT", "8081")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		slog.Info("starting health check server", "addr", ":"+port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("health check server failed", "error", err)
			stop()
		}
	}()

	// Wait for termination signal
	<-ctx.Done()

	slog.Info("Shutting down Customer Management service...")
	stop()

	// Shutdown HTTP Server
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Health check server forced to shutdown", "error", err)
	}

	// Give some time for other cleanup if needed, essentially waiting for listener to stop
	// The listener should stop when ctx is cancelled.
	// We could wait for it if we had a WaitGroup, but for now we follow the pattern.
	time.Sleep(1 * time.Second)
}

func runMigrations(dbURL string) {
	m, err := migrate.New(
		"file://internal/infrastructure/postgres/migrations",
		dbURL,
	)
	if err != nil {
		slog.Error("failed to create migration instance", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("Migrations completed successfully.")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
