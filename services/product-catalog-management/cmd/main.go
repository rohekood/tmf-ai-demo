package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tmf/services/product-catalog-management/internal/adapter/handler"
	"tmf/services/product-catalog-management/internal/adapter/publisher"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/usecase/catalog"
	"tmf/services/product-catalog-management/internal/usecase/category"
	"tmf/services/product-catalog-management/internal/usecase/offering"
	"tmf/services/product-catalog-management/internal/usecase/specification"
)

func main() {
	// 1. Config
	dbDSN := os.Getenv("DB_DSN")
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
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate
	log.Println("Running AutoMigrate...")
	err = db.AutoMigrate(
		&repository.CatalogModel{},
		&repository.CategoryModel{},
		&repository.ProductSpecificationModel{},
		&repository.ProductOfferingModel{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 3. RabbitMQ
	var conn *amqp.Connection
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ, retrying in 2s... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after retries: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}()

	// 4. Init Adapters (Repositories)
	catalogRepo := repository.NewCatalogRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	specRepo := repository.NewProductSpecificationRepo(db)
	offeringRepo := repository.NewProductOfferingRepo(db)

	// 5. Init Publisher
	rabbitPublisher, err := publisher.NewRabbitMQPublisher(conn)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ publisher: %v", err)
	}

	// 6. Init UseCases
	createCatalogUC := catalog.NewCreateCatalog(catalogRepo, rabbitPublisher)
	updateCatalogUC := catalog.NewUpdateCatalogUseCase(catalogRepo, rabbitPublisher)
	deleteCatalogUC := catalog.NewDeleteCatalogUseCase(catalogRepo, rabbitPublisher)
	listCatalogsUC := catalog.NewListCatalogs(catalogRepo)
	getCatalogUC := catalog.NewGetCatalog(catalogRepo)

	createCategoryUC := category.NewCreateCategory(categoryRepo, rabbitPublisher)
	updateCategoryUC := category.NewUpdateCategoryUseCase(categoryRepo, rabbitPublisher)
	deleteCategoryUC := category.NewDeleteCategoryUseCase(categoryRepo, rabbitPublisher)
	getCategoryUC := category.NewGetCategory(categoryRepo)
	listCategoriesUC := category.NewListCategories(categoryRepo)

	createProductSpecificationUC := specification.NewCreateProductSpecification(specRepo, rabbitPublisher)
	updateProductSpecificationUC := specification.NewUpdateProductSpecificationUseCase(specRepo, rabbitPublisher)
	deleteProductSpecificationUC := specification.NewDeleteProductSpecificationUseCase(specRepo, rabbitPublisher)
	getProductSpecificationUC := specification.NewGetProductSpecification(specRepo)
	listProductSpecificationsUC := specification.NewListProductSpecifications(specRepo)

	createProductOfferingUC := offering.NewCreateProductOffering(offeringRepo, rabbitPublisher)
	updateProductOfferingUC := offering.NewUpdateProductOfferingUseCase(offeringRepo, rabbitPublisher)
	deleteProductOfferingUC := offering.NewDeleteProductOfferingUseCase(offeringRepo, rabbitPublisher)
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
		log.Fatalf("Failed to initialize RabbitMQ handler: %v", err)
	}

	// 8. Start
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := rabbitHandler.Start(ctx); err != nil {
			log.Printf("RabbitMQ handler error: %v", err)
			cancel() // Shutdown on error
		}
	}()

	log.Println("Product Catalog Management Service Started")

	// 8. Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down...")
	// Cleanup happens via defers (conn.Close, etc.) and ctx cancellation
}
