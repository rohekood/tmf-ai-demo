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

func TestRabbitMQHandler_CreateProductSpecification(t *testing.T) {
	repo := repository.NewCatalogRepo(sharedDB)
	catRepo := repository.NewCategoryRepo(sharedDB)
	specRepo := repository.NewProductSpecificationRepo(sharedDB)

	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createUC := catalog.NewCreateCatalog(repo, pub, &repository.NoOpTransactionManager{})
	listUC := catalog.NewListCatalogs(repo)
	createCatUC := category.NewCreateCategory(catRepo, pub, &repository.NoOpTransactionManager{})
	createSpecUC := specification.NewCreateProductSpecification(specRepo, pub, &repository.NoOpTransactionManager{})
	createOfferingUC := offering.NewCreateProductOffering(repository.NewProductOfferingRepo(sharedDB), specRepo, pub, &repository.NoOpTransactionManager{})

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

	cmd := domain.ProductSpecificationCreateEvent{
		Name:            "Async Spec",
		ProductNumber:   "ASYNC-001",
		Description:     "Created via RabbitMQ",
		IsBundle:        false,
		LifecycleStatus: "Active",
		ValidFor:        domain.TimePeriod{},
	}
	body, _ := json.Marshal(cmd)

	err = ch.Publish("catalog_events", "cmd.catalog.specification.create", false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		list, err := specRepo.List(context.Background(), map[string]any{"name": "Async Spec"})
		return err == nil && len(list) == 1 && list[0].Name == "Async Spec" && list[0].ProductNumber == "ASYNC-001"
	}, 10*time.Second, 100*time.Millisecond, "Specification should be created via RabbitMQ")
}
