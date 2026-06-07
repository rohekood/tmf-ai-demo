package handler_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tmf/pkg/rabbitmq"
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
	repo := repository.NewCatalogRepo(sharedDB)
	catRepo := repository.NewCategoryRepo(sharedDB)

	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	err = sharedPub.DeclareTopicExchange("catalog_events", true, false, false, false)
	require.NoError(t, err)

	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createUC := catalog.NewCreateCatalog(repo, pub, &repository.NoOpTransactionManager{})
	listUC := catalog.NewListCatalogs(repo)
	createCatUC := category.NewCreateCategory(catRepo, pub, &repository.NoOpTransactionManager{})
	createSpecUC := specification.NewCreateProductSpecification(repository.NewProductSpecificationRepo(sharedDB), pub, &repository.NoOpTransactionManager{})
	createOfferingUC := offering.NewCreateProductOffering(repository.NewProductOfferingRepo(sharedDB), repository.NewProductSpecificationRepo(sharedDB), pub, &repository.NoOpTransactionManager{})

	h := handler.NewRabbitMQHandler(
		sharedPub,
		createUC, nil, nil, listUC, nil,
		createCatUC, nil, nil, nil, nil,
		createSpecUC, nil, nil, nil, nil,
		createOfferingUC, nil, nil, nil, nil,
	)
	bindTestHandler(t, h)

	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	cmd := domain.CatalogCreateEvent{
		Name:        "Async Catalog",
		Description: "Created via RabbitMQ",
		ValidFor:    domain.TimePeriod{},
	}
	body, _ := json.Marshal(cmd)

	err = ch.Publish("catalog_events", "cmd.catalog.catalog.create", false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		list, err := repo.List(context.Background(), map[string]any{"name": "Async Catalog"})
		return err == nil && len(list) == 1 && list[0].Name == "Async Catalog" && list[0].Description == "Created via RabbitMQ"
	}, 10*time.Second, 100*time.Millisecond, "Catalog should be created via RabbitMQ")
}
