package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
		defer shutdown(context.Background())
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
	defer connMgr.Close()

	// 4. Initialize Repository and Publisher
	repo := infraPostgres.NewPartyRepository(db)
	publisher, err := infraRabbit.NewPublisher(connMgr.GetConnection())
	if err != nil {
		slog.Error("failed to create publisher", "error", err)
		os.Exit(1)
	}

	// 5. Initialize Handlers and Listener
	handlers := rabbitTransport.NewHandlers(repo, publisher)
	listener, err := rabbitTransport.NewListener(connMgr.GetConnection())
	if err != nil {
		slog.Error("failed to create listener", "error", err)
		os.Exit(1)
	}

	// 6. Start Health Check Server
	healthHandler := transportHttp.NewHealthHandler(db, connMgr)
	metricsHandler := transportHttp.MetricsHandler()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/health", healthHandler)
		mux.Handle("/metrics", metricsHandler)

		slog.Info("starting health check server", "addr", ":8080")
		if err := http.ListenAndServe(":8080", mux); err != nil && err != http.ErrServerClosed {
			slog.Error("health check server failed", "error", err)
		}
	}()

	// 7. Subscribe to Queues
	// Start listener in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := listener.Start(ctx, handlers); err != nil {
			slog.Error("listener stopped", "error", err)
			cancel()
		}
	}()

	// Wait for termination signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("party management service is running")

	<-stop
	slog.Info("shutting down gracefully...")

	// 8. Graceful Shutdown
	// Listener stops when context is cancelled or main exits
	connMgr.Close()

	slog.Info("service stopped")
}
