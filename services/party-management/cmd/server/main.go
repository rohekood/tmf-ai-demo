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
	"tmf/services/party-management/internal/config"
	infraPostgres "tmf/services/party-management/internal/infrastructure/postgres"
	infraRabbit "tmf/services/party-management/internal/infrastructure/rabbitmq"
	"tmf/services/party-management/internal/infrastructure/telemetry"
	transportHttp "tmf/services/party-management/internal/transport/http"
	rabbitTransport "tmf/services/party-management/internal/transport/rabbitmq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("starting party management service")

	// Initialize OpenTelemetry
	shutdown, err := telemetry.InitTracer("party-management")
	if err != nil {
		slog.Error("failed to initialize tracer", "error", err)
	} else {
		defer func() { _ = shutdown(context.Background()) }()
	}

	cfg := config.Load()

	// 1. Database Connection
	db, err := gorm.Open(gormPostgres.Open(cfg.PostgresURL), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	// 2. Database Migrations
	m, err := migrate.New(
		"file://internal/infrastructure/postgres/migrations",
		cfg.PostgresURL,
	)
	if err != nil {
		slog.Error("failed to initialize migrations", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations applied successfully")

	// 3. RabbitMQ Connection Management
	connMgr := infraRabbit.NewConnectionManager(cfg.RabbitMQURL)
	if err := connMgr.Connect(); err != nil {
		slog.Error("failed to connect to rabbitmq", "error", err)
		os.Exit(1)
	}
	defer func() { _ = connMgr.Close() }()

	// 4. Initialize Repository and Publisher
	repo := infraPostgres.NewPartyRepository(db)
	publisher, err := rabbitmq.NewPublisherWithConnection(connMgr.GetConnection())
	if err != nil {
		slog.Error("failed to create publisher", "error", err)
		os.Exit(1)
	}
	if err := publisher.DeclareTopicExchange("tmf.events", true, false, false, false); err != nil {
		slog.Error("failed to declare exchange (tmf.events)", "error", err)
		os.Exit(1)
	}

	tm := infraPostgres.NewTransactionManager(db)
	outboxRepo := infraPostgres.NewOutboxRepository(db)
	outboxPublisher := infraPostgres.NewOutboxPublisher(outboxRepo)
	outboxWorker := infraPostgres.NewOutboxWorker(outboxRepo, publisher)

	// 5. Initialize Handlers and Listener
	handlers := rabbitTransport.NewHandlers(repo, outboxPublisher, publisher, tm)
	listener, err := rabbitTransport.NewListener(connMgr.GetConnection())
	if err != nil {
		slog.Error("failed to create listener", "error", err)
		os.Exit(1)
	}

	// 7. Subscribe to Queues
	// Start listener in a goroutine
	// Create a context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start Outbox Worker
	go outboxWorker.Start(ctx)

	go func() {
		if err := listener.Start(ctx, handlers); err != nil {
			slog.Error("listener stopped", "error", err)
			stop()
		}
	}()

	// 8. Start Health Check Server
	healthHandler := transportHttp.NewHealthHandler(db, connMgr)
	metricsHandler := transportHttp.MetricsHandler()

	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/metrics", metricsHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	go func() {
		slog.Info("starting health check server", "addr", ":"+cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("health check server failed", "error", err)
			stop()
		}
	}()

	slog.Info("party management service is running")

	// Wait for termination signal
	<-ctx.Done()
	slog.Info("shutting down gracefully...")
	stop()

	// 9. Graceful Shutdown
	// Shutdown HTTP Server
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Health check server forced to shutdown", "error", err)
	}

	// Wait for listener to cleanup if needed
	time.Sleep(1 * time.Second)

	slog.Info("service stopped")
}
