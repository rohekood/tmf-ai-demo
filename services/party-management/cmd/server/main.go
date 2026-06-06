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

	amqp "github.com/rabbitmq/amqp091-go"

	pkgrmq "tmf/pkg/rabbitmq"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("starting party management service")

	shutdown, err := telemetry.InitTracer("party-management")
	if err != nil {
		slog.Error("failed to initialize tracer", "error", err)
	} else {
		defer func() { _ = shutdown(context.Background()) }()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	envFn := os.Getenv
	dbDialFn := func(dsn string) (*gorm.DB, error) {
		return gorm.Open(gormPostgres.Open(dsn), &gorm.Config{})
	}
	rabbitConnFn := func(url string) rmqConnManager {
		return infraRabbit.NewConnectionManager(url)
	}

	runMigrationsFn := func(dsn string) error {
		m, err := migrate.New("file://internal/infrastructure/postgres/migrations", dsn)
		if err != nil {
			return err
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
		return nil
	}
	newPublisherFn := func(connMgr rmqConnManager) (rabbitmq.Publisher, error) {
		return rabbitmq.NewPublisherWithConnection(connMgr.GetConnection())
	}
	newListenerFn := func(connMgr rmqConnManager) (*rabbitTransport.Listener, error) {
		return rabbitTransport.NewListener(connMgr.GetConnection())
	}

	if err := run(ctx, envFn, dbDialFn, rabbitConnFn, runMigrationsFn, newPublisherFn, newListenerFn, logger); err != nil {
		slog.Error("service exited with error", "error", err)
		os.Exit(1)
	}
}

type rmqConnManager interface {
	Connect() error
	Close() error
	GetConnection() *amqp.Connection
}

func run(
	ctx context.Context,
	envFn func(string) string,
	dbDialFn func(string) (*gorm.DB, error),
	rabbitConnFn func(string) rmqConnManager,
	runMigrationsFn func(string) error,
	newPublisherFn func(rmqConnManager) (rabbitmq.Publisher, error),
	newListenerFn func(rmqConnManager) (*rabbitTransport.Listener, error),
	logger *slog.Logger,
) error {
	cfg := config.Load()

	db, err := dbDialFn(cfg.PostgresURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := runMigrationsFn(cfg.PostgresURL); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	logger.Info("database migrations applied successfully")

	connMgr := rabbitConnFn(cfg.RabbitMQURL)
	if err := connMgr.Connect(); err != nil {
		return fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}
	defer func() { _ = connMgr.Close() }()

	repo := infraPostgres.NewPartyRepository(db)
	publisher, err := newPublisherFn(connMgr)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}
	tm := infraPostgres.NewTransactionManager(db)
	outboxRepo := infraPostgres.NewOutboxRepository(db)
	outboxPublisher := infraPostgres.NewOutboxPublisher(outboxRepo)
	outboxWorker := infraPostgres.NewOutboxWorker(outboxRepo, publisher)

	handlers := rabbitTransport.NewHandlers(repo, outboxPublisher, publisher, tm)

	// Wire customer checker for pre-deletion validation.
	customerRPCClient, err := pkgrmq.NewRPCClient(cfg.RabbitMQURL)
	if err != nil {
		logger.Warn("failed to create customer RPC client, deletion pre-check disabled", "error", err)
	} else {
		handlers.WithCustomerChecker(infraRabbit.NewCustomerCheckerRPC(customerRPCClient))
	}

	listener, err := newListenerFn(connMgr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	go outboxWorker.Start(signalCtx)

	if listener != nil {
		go func() {
			if err := listener.Start(signalCtx, handlers); err != nil {
				logger.Error("listener stopped", "error", err)
				stop()
			}
		}()
	}

	var cMgr *infraRabbit.ConnectionManager
	if m, ok := connMgr.(*infraRabbit.ConnectionManager); ok {
		cMgr = m
	}
	healthHandler := transportHttp.NewHealthHandler(db, cMgr)
	metricsHandler := transportHttp.MetricsHandler()

	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/metrics", metricsHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	go func() {
		logger.Info("starting health check server", "addr", ":"+cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health check server failed", "error", err)
			stop()
		}
	}()

	logger.Info("party management service is running")

	<-signalCtx.Done()
	logger.Info("shutting down gracefully...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("health check server forced to shutdown", "error", err)
	}

	time.Sleep(1 * time.Second)
	logger.Info("service stopped")
	return nil
}
