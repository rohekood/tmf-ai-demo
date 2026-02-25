package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"tmf/pkg/rabbitmq"
	"tmf/services/shopping-cart/internal/adapter/handler"
	"tmf/services/shopping-cart/internal/adapter/repository"
	"tmf/services/shopping-cart/internal/adapter/rpc"
	"tmf/services/shopping-cart/internal/adapter/worker"
	"tmf/services/shopping-cart/internal/usecase"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	// 1. Config
	rabbitURL := getEnv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	dbURL := getEnv("DB_URL")
	if dbURL == "" {
		return fmt.Errorf("DB_URL environment variable is required")
	}

	// 2. Database Connection
	db, err := gorm.Open(gormPostgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 3. Database Migrations
	m, err := migrate.New(
		"file://internal/infrastructure/postgres/migrations",
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	slog.Info("database migrations applied successfully")

	// 4. Infrastructure (RabbitMQ)
	pub, err := rabbitmq.NewPublisher(rabbitURL)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}
	defer func() { _ = pub.Close() }()

	if err := pub.DeclareTopicExchange("ex.domain.commerce", true, false, false, false); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	if err := pub.DeclareTopicExchange("ex.domain.market", true, false, false, false); err != nil {
		return fmt.Errorf("failed to declare market exchange: %w", err)
	}

	// 5. Layers
	repo := repository.NewCartRepository(db)

	// 5.1 Create RPC client for qualification service
	rpcClient, err := rabbitmq.NewRPCClient(rabbitURL)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}
	defer func() { _ = rpcClient.Close() }()

	qualRPCClient := rpc.NewQualificationClient(rpcClient)
	qualClient := rpc.NewQualificationClientAdapter(qualRPCClient)

	manageUC := usecase.NewManageItemsUseCase(repo, qualClient)
	syncUC := usecase.NewSyncCatalogUseCase(repo)
	priceUC := usecase.NewUpdatePriceUseCase(repo)

	// Inject repo and pub for RPC
	h := handler.NewCartHandler(manageUC, priceUC, syncUC, repo, pub)

	// 6. Workers
	outboxWorker := worker.NewOutboxWorker(db, pub, "ex.domain.commerce")

	// 7. Listener (Commands)
	consumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.commerce", "q.cart.commands")
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer func() { _ = consumer.Close() }()

	if err := consumer.Subscribe(rabbitmq.CmdCartItemAdd, h.HandleAddItem); err != nil {
		return fmt.Errorf("failed to subscribe cmds: %w", err)
	}

	// 8. Listener (Catalog Replication)
	catalogConsumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.catalog", "q.cart.catalog.sync")
	if err != nil {
		return fmt.Errorf("failed to create catalog consumer: %w", err)
	}
	defer func() { _ = catalogConsumer.Close() }()

	if err := catalogConsumer.Subscribe("evt.catalog.offering.#", h.HandleCatalogEvent); err != nil {
		return fmt.Errorf("failed to subscribe catalog events: %w", err)
	}

	// 9. Listener (RPC Queries)
	rpcConsumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.commerce", "q.cart.rpc.v2")
	if err != nil {
		return fmt.Errorf("failed to create rpc consumer: %w", err)
	}
	defer func() { _ = rpcConsumer.Close() }()

	if err := rpcConsumer.Subscribe("query.cart.session.get", h.HandleGetCart); err != nil {
		return fmt.Errorf("failed to subscribe rpc: %w", err)
	}

	// 10. Start
	go outboxWorker.Start(ctx)

	slog.Info("Shopping Cart Service Started")
	<-ctx.Done()
	slog.Info("Shutting down...")
	return nil
}
