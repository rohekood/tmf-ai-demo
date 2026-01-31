package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"tmf/pkg/rabbitmq"
	"tmf/services/shopping-cart/internal/adapter/handler"
	"tmf/services/shopping-cart/internal/adapter/repository"
	"tmf/services/shopping-cart/internal/adapter/worker"
	"tmf/services/shopping-cart/internal/usecase"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 1. Config
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		slog.Error("DB_URL environment variable is required")
		os.Exit(1)
	}

	// 2. Database Connection
	db, err := gorm.Open(gormPostgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// 3. Database Migrations
	m, err := migrate.New(
		"file://internal/infrastructure/postgres/migrations",
		dbURL,
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

	// 4. Infrastructure (RabbitMQ)
	pub, err := rabbitmq.NewPublisher(rabbitURL)
	if err != nil {
		slog.Error("Failed to create publisher", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := pub.Close(); err != nil {
			slog.Error("Failed to close publisher", "error", err)
		}
	}()
	if err := pub.DeclareTopicExchange("ex.domain.commerce", true, false, false, false); err != nil {
		slog.Error("Failed to declare exchange", "error", err)
		os.Exit(1)
	}

	// 5. Layers
	repo := repository.NewCartRepository(db)

	manageUC := usecase.NewManageItemsUseCase(repo)
	syncUC := usecase.NewSyncCatalogUseCase(repo)
	priceUC := usecase.NewUpdatePriceUseCase(repo)

	// Inject repo and pub for RPC
	h := handler.NewCartHandler(manageUC, priceUC, syncUC, repo, pub)

	// 6. Workers
	outboxWorker := worker.NewOutboxWorker(db, pub, "ex.domain.commerce")

	// 7. Listener (Commands)
	consumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.commerce", "q.cart.commands")
	if err != nil {
		slog.Error("Failed to create consumer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			slog.Error("Failed to close consumer", "error", err)
		}
	}()

	// Bindings
	if err := consumer.Subscribe(rabbitmq.CmdCartItemAdd, h.HandleAddItem); err != nil {
		slog.Error("Failed to subscribe", "topic", rabbitmq.CmdCartItemAdd, "error", err)
		os.Exit(1)
	}

	// 8. Listener (Catalog Replication)
	catalogConsumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.catalog", "q.cart.catalog.sync")
	if err != nil {
		slog.Error("Failed to create catalog consumer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := catalogConsumer.Close(); err != nil {
			slog.Error("Failed to close catalogConsumer", "error", err)
		}
	}()

	if err := catalogConsumer.Subscribe("evt.catalog.offering.#", h.HandleCatalogEvent); err != nil {
		slog.Error("Failed to subscribe catalog events", "error", err)
		os.Exit(1)
	}

	// 9. Listener (RPC Queries)
	// We use the same Exchange but different queue for RPC queries
	rpcConsumer, err := rabbitmq.NewConsumer(rabbitURL, "ex.domain.commerce", "q.cart.rpc.v2")
	if err != nil {
		slog.Error("Failed to create rpc consumer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := rpcConsumer.Close(); err != nil {
			slog.Error("Failed to close rpcConsumer", "error", err)
		}
	}()

	// Bind Query Topic (Assuming defined in pkg or string literal)
	// "query.cart.session.get"
	if err := rpcConsumer.Subscribe("query.cart.session.get", h.HandleGetCart); err != nil {
		slog.Error("Failed to subscribe rpc", "error", err)
		os.Exit(1)
	}

	// 10. Start
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go outboxWorker.Start(ctx)

	slog.Info("Shopping Cart Service Started")
	<-ctx.Done()
	slog.Info("Shutting down...")
}
