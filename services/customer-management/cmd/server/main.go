package main

import (
	"context"
	"errors"
	"fmt"
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

var osExit = os.Exit

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create a context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, getEnv,
		func(dsn string) (*gorm.DB, error) { return gorm.Open(gormPostgres.Open(dsn), &gorm.Config{}) },
		amqp.Dial,
		runMigrations,
		rabbitmq.NewPublisherWithConnection,
		transportRabbit.NewListener,
		func(conn *amqp.Connection) (rabbitmq.Consumer, error) {
			return rabbitmq.NewConsumerWithConnection(conn, "customer.events", "customer_rpc_queue")
		},
		logger); err != nil {
		slog.Error("service exited with error", "error", err)
		osExit(1)
	}
}

func run(ctx context.Context,
	getEnvFn func(string, string) string,
	dbDialFn func(string) (*gorm.DB, error),
	rabbitDialFn func(string) (*amqp.Connection, error),
	runMigrationsFn func(string) error,
	newPublisherFn func(*amqp.Connection) (rabbitmq.Publisher, error),
	newListenerFn func(*amqp.Connection) (*transportRabbit.Listener, error),
	newRpcConsumerFn func(*amqp.Connection) (rabbitmq.Consumer, error),
	logger *slog.Logger) error {
	slog.Info("starting customer management service")

	// Initialize OpenTelemetry
	shutdown, err := telemetry.InitTracer("customer-management")
	if err != nil {
		slog.Error("failed to initialize tracer", "error", err)
	} else {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				slog.Error("Failed to shutdown trace provider", "error", err)
			}
		}()
	}

	// Configuration (using defaults or env vars)
	dbURL := getEnvFn("POSTGRES_URL", "postgres://postgres:password@localhost:5432/tmf_customer_db?sslmode=disable")
	rabbitURL := getEnvFn("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	// 1. Database Migrations
	if err := runMigrationsFn(dbURL); err != nil {
		return err
	}

	// 2. Database Connection
	db, err := dbDialFn(dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	repo := postgres.NewCustomerRepository(db)
	tm := postgres.NewTransactionManager(db)
	outboxRepo := postgres.NewOutboxRepository(db)
	eventPublisher := postgres.NewOutboxPublisher(outboxRepo)

	// 3. RabbitMQ Connection
	conn, err := rabbitDialFn(rabbitURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer func() {
		if conn != nil {
			defer func() { recover() }() // Catch amqp091-go panics on empty structs
			if err := conn.Close(); err != nil {
				slog.Error("Failed to close RabbitMQ connection", "error", err)
			}
		}
	}()

	publisher, err := newPublisherFn(conn)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			slog.Error("Failed to close publisher", "error", err)
		}
	}()

	outboxWorker := postgres.NewOutboxWorker(outboxRepo, publisher, slog.Default())

	// 4. Handlers & Listener
	handlers := transportRabbit.NewHandlers(repo, publisher, tm, eventPublisher)
	listener, err := newListenerFn(conn)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// 4.1 Initialize RPC Handler for pricing queries
	rpcHandler := transportRabbit.NewRPCHandler(repo, publisher, logger)

	// 4.2 Create RPC consumer
	rpcConsumer, err := newRpcConsumerFn(conn)
	if err != nil {
		return fmt.Errorf("failed to create RPC consumer: %w", err)
	}

	// 4.3 Bind RPC handlers
	if err := rpcHandler.BindRPCHandlers(rpcConsumer); err != nil {
		return fmt.Errorf("failed to bind RPC handlers: %w", err)
	}

	// 5. Start Service
	// Create another cancellable context for internal workers tied to the run lifecycle
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	go outboxWorker.Start(workerCtx)
	defer outboxWorker.Stop()

	go func() {
		if conn != nil && listener != nil {
			if err := listener.Start(workerCtx, handlers); err != nil {
				slog.Error("listener stopped", "error", err)
				workerCancel()
			}
		}
	}()

	// 6. Start Health Check Server
	healthHandler := transportHttp.NewHealthHandler(db, conn)
	metricsHandler := transportHttp.MetricsHandler()

	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/metrics", metricsHandler)

	port := getEnvFn("HTTP_PORT", "8081")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		slog.Info("starting health check server", "addr", ":"+port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("health check server failed", "error", err)
			if workerCancel != nil {
				workerCancel()
			}
		}
	}()

	// Wait for termination signal
	<-ctx.Done()

	slog.Info("Shutting down Customer Management service...")
	if workerCancel != nil {
		workerCancel() // Cancel background listeners
	}

	// Shutdown HTTP Server
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Health check server forced to shutdown", "error", err)
	}

	time.Sleep(1 * time.Second)
	return nil
}

func runMigrations(dbURL string) error {
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
	slog.Info("Migrations completed successfully.")
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
