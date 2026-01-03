package handler_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tmf/services/product-catalog-management/internal/adapter/handler"
	"tmf/services/product-catalog-management/internal/adapter/publisher"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/usecase/catalog"
	"tmf/services/product-catalog-management/internal/usecase/category"
	"tmf/services/product-catalog-management/internal/usecase/offering"
	"tmf/services/product-catalog-management/internal/usecase/specification"
)

func TestRabbitMQHandler_CreateCatalog(t *testing.T) {
	// Setup Dependencies
	repo := repository.NewCatalogRepo(sharedDB)
	catRepo := repository.NewCategoryRepo(sharedDB)
	// Create Publisher
	pub, err := publisher.NewRabbitMQPublisher(rabbitConn)
	require.NoError(t, err)

	createUC := catalog.NewCreateCatalog(repo, pub)
	listUC := catalog.NewListCatalogs(repo)
	createCatUC := category.NewCreateCategory(catRepo, pub)
	createSpecUC := specification.NewCreateProductSpecification(repository.NewProductSpecificationRepo(sharedDB), pub)
	createOfferingUC := offering.NewCreateProductOffering(repository.NewProductOfferingRepo(sharedDB), pub)

	// Init Handler
	h, err := handler.NewRabbitMQHandler(
		rabbitConn,
		createUC,
		nil, // updateCatalogUC
		nil, // deleteCatalogUC
		listUC,
		nil, // getCatalogUC
		createCatUC,
		nil, // updateCategoryUC
		nil, // deleteCategoryUC
		nil, // getCategoryUC
		nil, // listCategoriesUC
		createSpecUC,
		nil, // updateProductSpecificationUC
		nil, // deleteProductSpecificationUC
		nil, // getProductSpecificationUC
		nil, // listProductSpecificationsUC
		createOfferingUC,
		nil, // updateProductOfferingUC
		nil, // deleteProductOfferingUC
		nil, // getProductOfferingUC
		nil, // listProductOfferingsUC
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Handler in background
	go func() {
		_ = h.Start(ctx)
	}()

	// Wait for handler to initialize queues
	time.Sleep(1 * time.Second)

	// Publish Command
	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() {
		if err := ch.Close(); err != nil {
			t.Logf("Error closing channel: %v", err)
		}
	}()

	cmd := domain.CatalogCreateEvent{
		Name:        "Async Catalog",
		Description: "Created via RabbitMQ",
		ValidFor:    domain.TimePeriod{},
	}
	body, _ := json.Marshal(cmd)

	err = ch.Publish(
		"catalog_events",             // exchange
		"cmd.catalog.catalog.create", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Verify in DB
	list, err := repo.List(context.Background(), map[string]interface{}{"name": "Async Catalog"})
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Async Catalog", list[0].Name)
	assert.Equal(t, "Created via RabbitMQ", list[0].Description)
}
