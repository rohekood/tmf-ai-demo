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

func TestRabbitMQHandler_CreateProductOffering(t *testing.T) {
	repo := repository.NewCatalogRepo(sharedDB)
	catRepo := repository.NewCategoryRepo(sharedDB)
	specRepo := repository.NewProductSpecificationRepo(sharedDB)
	offeringRepo := repository.NewProductOfferingRepo(sharedDB)

	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createUC := catalog.NewCreateCatalog(repo, pub, &repository.NoOpTransactionManager{})
	listUC := catalog.NewListCatalogs(repo)
	createCatUC := category.NewCreateCategory(catRepo, pub, &repository.NoOpTransactionManager{})
	createSpecUC := specification.NewCreateProductSpecification(specRepo, pub, &repository.NoOpTransactionManager{})
	createOfferingUC := offering.NewCreateProductOffering(offeringRepo, specRepo, pub, &repository.NoOpTransactionManager{})

	h := handler.NewRabbitMQHandler(
		sharedPub,
		createUC, nil, nil, listUC, nil,
		createCatUC, nil, nil, nil, nil,
		createSpecUC, nil, nil, nil, nil,
		createOfferingUC, nil, nil, nil, nil,
	)
	bindTestHandler(t, h)

	specID := "spec-for-offering"
	spec := &domain.ProductSpecification{ID: specID, Name: "Spec For Offering", ProductNumber: "SFO-001"}
	err = specRepo.Create(context.Background(), spec)
	require.NoError(t, err)

	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	cmd := domain.ProductOfferingCreateEvent{
		Name:            "Async Offering",
		Description:     "Created via RabbitMQ",
		IsBundle:        false,
		IsSellable:      true,
		LifecycleStatus: "Active",
		ValidFor:        domain.TimePeriod{},
		ProductSpecID:   &specID,
		Prices: []domain.ProductOfferingPrice{
			{PriceType: "recurring", Price: domain.Money{Value: 120.0, Unit: "USD"}},
		},
	}
	body, _ := json.Marshal(cmd)

	err = ch.Publish("catalog_events", "cmd.catalog.offering.create", false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		list, err := offeringRepo.List(context.Background(), map[string]any{"name": "Async Offering"})
		return err == nil && len(list) == 1 && list[0].Name == "Async Offering"
	}, 10*time.Second, 100*time.Millisecond, "Offering should be created via RabbitMQ")
}

func TestRabbitMQHandler_Offering_AdvancedFeatures(t *testing.T) {
	offeringRepo := repository.NewProductOfferingRepo(sharedDB)
	specRepo := repository.NewProductSpecificationRepo(sharedDB)
	catRepo := repository.NewCategoryRepo(sharedDB)

	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createOfferingUC := offering.NewCreateProductOffering(offeringRepo, specRepo, pub, &repository.NoOpTransactionManager{})
	getOfferingUC := offering.NewGetProductOffering(offeringRepo, specRepo, catRepo)
	listOfferingUC := offering.NewListProductOfferings(offeringRepo)

	h := handler.NewRabbitMQHandler(
		sharedPub,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		createOfferingUC, nil, nil, getOfferingUC, listOfferingUC,
	)
	bindTestHandler(t, h)

	specID := "adv-spec-001"
	spec := &domain.ProductSpecification{ID: specID, Name: "Advanced Feature Spec", ProductNumber: "AFS-001"}
	require.NoError(t, specRepo.Create(context.Background(), spec))

	catID := "adv-cat-001"
	cat := &domain.Category{ID: catID, Name: "Advanced Category"}
	require.NoError(t, catRepo.Create(context.Background(), cat))

	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	q, err := ch.QueueDeclare("", false, false, true, false, nil)
	require.NoError(t, err)
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	offeringName := "Advanced Offering with Attachments"
	cmd := domain.ProductOfferingCreateEvent{
		Name:            offeringName,
		LifecycleStatus: "Active",
		ProductSpecID:   &specID,
		CategoryIDs:     []string{catID},
		Prices:          []domain.ProductOfferingPrice{{PriceType: "recurring", Price: domain.Money{Value: 99.99, Unit: "USD"}}},
		Attachments:     []domain.Attachment{{Name: "Manual", URL: "http://example.com/manual.pdf", Type: "Document"}},
	}
	body, _ := json.Marshal(cmd)
	err = ch.Publish("catalog_events", "cmd.catalog.offering.create", false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	var offeringID string
	assert.Eventually(t, func() bool {
		list, err := offeringRepo.List(context.Background(), map[string]any{"name": offeringName})
		if err != nil || len(list) != 1 {
			return false
		}
		offeringID = list[0].ID
		return len(list[0].Attachments) == 1 && list[0].Attachments[0].Name == "Manual"
	}, 10*time.Second, 100*time.Millisecond, "Offering with attachments should be created via RabbitMQ")

	filterPayload := map[string]any{"minPrice": 50.0, "maxPrice": 150.0, "category": catID}
	filterBody, _ := json.Marshal(filterPayload)
	corrId := "filter-req-adv-1"
	err = ch.Publish("catalog_events", "query.catalog.offering.list", false, false,
		amqp.Publishing{ContentType: "application/json", CorrelationId: corrId, ReplyTo: q.Name, Body: filterBody})
	require.NoError(t, err)

	select {
	case d := <-msgs:
		assert.Equal(t, corrId, d.CorrelationId)
		var results []*domain.ProductOffering
		err := json.Unmarshal(d.Body, &results)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, offeringName, results[0].Name)
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for filter response")
	}

	getPayload := map[string]any{"id": offeringID, "enrich": true}
	getBody, _ := json.Marshal(getPayload)
	corrIdGet := "get-req-adv-1"
	err = ch.Publish("catalog_events", "query.catalog.offering.get", false, false,
		amqp.Publishing{ContentType: "application/json", CorrelationId: corrIdGet, ReplyTo: q.Name, Body: getBody})
	require.NoError(t, err)

	select {
	case d := <-msgs:
		assert.Equal(t, corrIdGet, d.CorrelationId)
		var result domain.ProductOffering
		err := json.Unmarshal(d.Body, &result)
		assert.NoError(t, err)
		assert.Equal(t, offeringID, result.ID)
		assert.NotNil(t, result.ProductSpecification)
		assert.Equal(t, "Advanced Feature Spec", result.ProductSpecification.Name)
		assert.Len(t, result.Categories, 1)
		assert.Equal(t, "Advanced Category", result.Categories[0].Name)
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for get response")
	}
}
