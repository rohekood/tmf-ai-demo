package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tmf/services/customer-management/internal/infrastructure/postgres"
	infraRabbit "tmf/services/customer-management/internal/infrastructure/rabbitmq"
	"tmf/services/customer-management/internal/infrastructure/telemetry"
	"tmf/services/customer-management/internal/transport/rabbitmq"

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
		defer shutdown(context.Background())
	}

	// Configuration (using defaults or env vars)
	dbURL := getEnv("DB_URL", "postgres://postgres:password@localhost:5432/tmf_customer_db?sslmode=disable")
	rabbitURL := getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/")

	// 1. Database Migrations
	runMigrations(dbURL)

	// 2. Database Connection
	db, err := gorm.Open(gormPostgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	repo := postgres.NewCustomerRepository(db)

	// 3. RabbitMQ Connection
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		slog.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	publisher, err := infraRabbit.NewPublisher(conn)
	if err != nil {
		slog.Error("failed to create publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// 4. Handlers & Listener
	handlers := rabbitmq.NewHandlers(repo, publisher)
	listener, err := rabbitmq.NewListener(conn)
	if err != nil {
		slog.Error("failed to create listener", "error", err)
		os.Exit(1)
	}

	// 5. Start Service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := listener.Start(ctx, handlers); err != nil {
			slog.Error("listener stopped", "error", err)
			cancel()
		}
	}()

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down Customer Management service...")
	cancel()
	time.Sleep(1 * time.Second) // Give some time for cleanup
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
