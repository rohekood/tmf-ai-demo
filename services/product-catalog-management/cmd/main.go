package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tmf/pkg/rabbitmq"
	"tmf/services/product-catalog-management/internal/adapter/handler"
	"tmf/services/product-catalog-management/internal/adapter/publisher"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/adapter/worker"
	"tmf/services/product-catalog-management/internal/usecase/catalog"
	"tmf/services/product-catalog-management/internal/usecase/category"
	"tmf/services/product-catalog-management/internal/usecase/offering"
	"tmf/services/product-catalog-management/internal/usecase/specification"
)

func main() {
	// Init logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	// 1. Config
	dbDSN := os.Getenv("POSTGRES_URL")
	if dbDSN == "" {
		dbDSN = "host=localhost user=postgres password=postgres dbname=tmf port=5432 sslmode=disable"
	}
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	// 2. Database
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Database Migrations
	logger.Info("Running database migrations...")
	m, err := migrate.New(
		"file://internal/adapter/repository/migrations",
		dbDSN,
	)
	if err != nil {
		logger.Error("Failed to initialize migrations", "error", err)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Database migrations applied successfully")

	// 3. RabbitMQ
	var conn *amqp.Connection
	for i := range 10 {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		logger.Warn("Failed to connect to RabbitMQ, retrying in 2s...", "error", err, "attempt", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ after retries", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error("Error closing RabbitMQ connection", "error", err)
		}
	}()

	// 4. Init Adapters (Repositories)
	catalogRepo := repository.NewCatalogRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	specRepo := repository.NewProductSpecificationRepo(db)
	offeringRepo := repository.NewProductOfferingRepo(db)

	// 5. Init Publisher & Transaction Manager
	// Create shared Publisher from pkg/rabbitmq
	sharedPublisher, err := rabbitmq.NewPublisherWithConnection(conn)
	if err != nil {
		logger.Error("Failed to initialize shared rabbitmq publisher", "error", err)
		os.Exit(1)
	}
	// Wrap it in adapter
	rabbitPublisher, err := publisher.NewRabbitMQPublisher(sharedPublisher, "catalog_events")
	if err != nil {
		logger.Error("Failed to initialize RabbitMQ publisher adapter", "error", err)
		os.Exit(1)
	}

	tm := repository.NewTransactionManager(db)
	outboxPublisher := publisher.NewOutboxPublisher(db)
	outboxWorker := worker.NewOutboxWorker(db, rabbitPublisher)

	// 6. Init UseCases (Injecting tm and outboxPublisher)
	createCatalogUC := catalog.NewCreateCatalog(catalogRepo, outboxPublisher, tm)
	updateCatalogUC := catalog.NewUpdateCatalogUseCase(catalogRepo, outboxPublisher, tm)
	deleteCatalogUC := catalog.NewDeleteCatalogUseCase(catalogRepo, outboxPublisher, tm)
	listCatalogsUC := catalog.NewListCatalogs(catalogRepo)
	getCatalogUC := catalog.NewGetCatalog(catalogRepo)

	createCategoryUC := category.NewCreateCategory(categoryRepo, outboxPublisher, tm)
	updateCategoryUC := category.NewUpdateCategoryUseCase(categoryRepo, outboxPublisher, tm)
	deleteCategoryUC := category.NewDeleteCategoryUseCase(categoryRepo, outboxPublisher, tm)
	getCategoryUC := category.NewGetCategory(categoryRepo)
	listCategoriesUC := category.NewListCategories(categoryRepo)

	createProductSpecificationUC := specification.NewCreateProductSpecification(specRepo, outboxPublisher, tm)
	updateProductSpecificationUC := specification.NewUpdateProductSpecificationUseCase(specRepo, outboxPublisher, tm)
	deleteProductSpecificationUC := specification.NewDeleteProductSpecificationUseCase(specRepo, outboxPublisher, tm)
	getProductSpecificationUC := specification.NewGetProductSpecification(specRepo)
	listProductSpecificationsUC := specification.NewListProductSpecifications(specRepo)

	// Inject TM into CreateProductOffering
	createProductOfferingUC := offering.NewCreateProductOffering(offeringRepo, specRepo, outboxPublisher, tm)
	updateProductOfferingUC := offering.NewUpdateProductOfferingUseCase(offeringRepo, specRepo, outboxPublisher, tm)
	deleteProductOfferingUC := offering.NewDeleteProductOfferingUseCase(offeringRepo, outboxPublisher, tm)
	getProductOfferingUC := offering.NewGetProductOffering(offeringRepo, specRepo, categoryRepo)
	listProductOfferingsUC := offering.NewListProductOfferings(offeringRepo)

	// 7. Init Handlers
	rabbitHandler, err := handler.NewRabbitMQHandler(
		conn,
		createCatalogUC,
		updateCatalogUC,
		deleteCatalogUC,
		listCatalogsUC,
		getCatalogUC,
		createCategoryUC,
		updateCategoryUC,
		deleteCategoryUC,
		getCategoryUC,
		listCategoriesUC,
		createProductSpecificationUC,
		updateProductSpecificationUC,
		deleteProductSpecificationUC,
		getProductSpecificationUC,
		listProductSpecificationsUC,
		createProductOfferingUC,
		updateProductOfferingUC,
		deleteProductOfferingUC,
		getProductOfferingUC,
		listProductOfferingsUC,
	)
	if err != nil {
		logger.Error("Failed to initialize RabbitMQ handler", "error", err)
		os.Exit(1)
	}

	// 7.1 Init RPC Handler for pricing queries
	rpcHandler := handler.NewCatalogRPCHandler(offeringRepo, sharedPublisher, logger)

	// 7.2 Create RPC consumer
	rpcConsumer, err := rabbitmq.NewConsumerWithConnection(conn, "catalog_events", "catalog_rpc_queue")
	if err != nil {
		logger.Error("Failed to create RPC consumer", "error", err)
		os.Exit(1)
	}

	// 7.3 Bind RPC handlers
	if err := rpcHandler.BindRPCHandlers(rpcConsumer); err != nil {
		logger.Error("Failed to bind RPC handlers", "error", err)
		os.Exit(1)
	}

	// 8. Start
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := rabbitHandler.Start(ctx); err != nil {
			logger.Error("RabbitMQ handler error", "error", err)
			cancel() // Shutdown on error
		}
	}()

	go outboxWorker.Start(ctx)

	logger.Info("Product Catalog Management Service Started")

	// 8. Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down...")
	// Cleanup happens via defers (conn.Close, etc.) and ctx cancellation
}
