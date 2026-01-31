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
	// Setup Dependencies
	repo := repository.NewCatalogRepo(sharedDB)
	catRepo := repository.NewCategoryRepo(sharedDB)
	specRepo := repository.NewProductSpecificationRepo(sharedDB)
	offeringRepo := repository.NewProductOfferingRepo(sharedDB)

	// Create Publisher
	// Create Publisher
	// Create Publisher
	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createUC := catalog.NewCreateCatalog(repo, pub, &repository.NoOpTransactionManager{})
	listUC := catalog.NewListCatalogs(repo)
	createCatUC := category.NewCreateCategory(catRepo, pub, &repository.NoOpTransactionManager{})
	createSpecUC := specification.NewCreateProductSpecification(specRepo, pub, &repository.NoOpTransactionManager{})
	createOfferingUC := offering.NewCreateProductOffering(offeringRepo, specRepo, pub, &repository.NoOpTransactionManager{})

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

	// Start Handler
	go func() {
		_ = h.Start(ctx)
	}()

	time.Sleep(1 * time.Second)

	// Create a Specification first referenced by Offering
	specID := "spec-for-offering"
	spec := &domain.ProductSpecification{
		ID:            specID,
		Name:          "Spec For Offering",
		ProductNumber: "SFO-001",
	}
	err = specRepo.Create(context.Background(), spec)
	require.NoError(t, err)

	// Publish Command
	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() {
		if err := ch.Close(); err != nil {
			t.Logf("Error closing channel: %v", err)
		}
	}()

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

	err = ch.Publish(
		"catalog_events",              // exchange
		"cmd.catalog.offering.create", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	require.NoError(t, err)

	// Wait for processing using Eventually
	assert.Eventually(t, func() bool {
		list, err := offeringRepo.List(context.Background(), map[string]interface{}{"name": "Async Offering"})
		return err == nil && len(list) == 1 && list[0].Name == "Async Offering"
	}, 10*time.Second, 100*time.Millisecond, "Offering should be created via RabbitMQ")
}

func TestRabbitMQHandler_Offering_AdvancedFeatures(t *testing.T) {
	// Setup Dependencies
	offeringRepo := repository.NewProductOfferingRepo(sharedDB)
	specRepo := repository.NewProductSpecificationRepo(sharedDB)
	catRepo := repository.NewCategoryRepo(sharedDB)
	// Create Publisher
	// Create Publisher
	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createOfferingUC := offering.NewCreateProductOffering(offeringRepo, specRepo, pub, &repository.NoOpTransactionManager{})
	getOfferingUC := offering.NewGetProductOffering(offeringRepo, specRepo, catRepo)
	listOfferingUC := offering.NewListProductOfferings(offeringRepo)

	// Init Handler with relevant use cases
	h, err := handler.NewRabbitMQHandler(
		rabbitConn,
		nil, nil, nil, nil, nil, // catalog
		nil, nil, nil, nil, nil, // category
		nil, nil, nil, nil, nil, // spec
		createOfferingUC,
		nil,
		nil,
		getOfferingUC,
		listOfferingUC,
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Handler
	go func() {
		_ = h.Start(ctx)
	}()

	time.Sleep(1 * time.Second)

	// 1. Setup Data: Create Specification and Category
	specID := "adv-spec-001"
	spec := &domain.ProductSpecification{
		ID:            specID,
		Name:          "Advanced Feature Spec",
		ProductNumber: "AFS-001",
	}
	require.NoError(t, specRepo.Create(context.Background(), spec))

	catID := "adv-cat-001"
	cat := &domain.Category{
		ID:   catID,
		Name: "Advanced Category",
	}
	require.NoError(t, catRepo.Create(context.Background(), cat))

	// 2. Test Attachments & Create Offering
	offeringName := "Advanced Offering with Attachments"
	cmd := domain.ProductOfferingCreateEvent{
		Name:            offeringName,
		LifecycleStatus: "Active",
		ProductSpecID:   &specID,
		CategoryIDs:     []string{catID},
		Prices: []domain.ProductOfferingPrice{
			{PriceType: "recurring", Price: domain.Money{Value: 99.99, Unit: "USD"}},
		},
		Attachments: []domain.Attachment{
			{Name: "Manual", URL: "http://example.com/manual.pdf", Type: "Document"},
		},
	}
	body, _ := json.Marshal(cmd)

	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	// Create Queue for RPC responses
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	require.NoError(t, err)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	require.NoError(t, err)

	err = ch.Publish(
		"catalog_events",              // exchange
		"cmd.catalog.offering.create", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	require.NoError(t, err)

	// Verify Create & Attachments via Repo using Eventually
	var offeringID string
	assert.Eventually(t, func() bool {
		list, err := offeringRepo.List(context.Background(), map[string]interface{}{"name": offeringName})
		if err != nil || len(list) != 1 {
			return false
		}
		offeringID = list[0].ID
		return len(list[0].Attachments) == 1 && list[0].Attachments[0].Name == "Manual"
	}, 10*time.Second, 100*time.Millisecond, "Offering with attachments should be created via RabbitMQ")

	// 3. Test Advanced Filtering
	// Send RPC Query for Filtering
	filterPayload := map[string]interface{}{
		"minPrice": 50.0,
		"maxPrice": 150.0,
		"category": catID,
	}
	filterBody, _ := json.Marshal(filterPayload)

	corrId := "filter-req-adv-1"
	err = ch.Publish(
		"catalog_events",              // exchange
		"query.catalog.offering.list", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          filterBody,
		},
	)
	require.NoError(t, err)

	// Wait for response
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

	// 4. Test Enriched Retrieval
	getPayload := map[string]interface{}{
		"id":     offeringID,
		"enrich": true,
	}
	getBody, _ := json.Marshal(getPayload)
	corrIdGet := "get-req-adv-1"

	err = ch.Publish(
		"catalog_events",             // exchange
		"query.catalog.offering.get", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrIdGet,
			ReplyTo:       q.Name,
			Body:          getBody,
		},
	)
	require.NoError(t, err)

	select {
	case d := <-msgs:
		assert.Equal(t, corrIdGet, d.CorrelationId)
		var result domain.ProductOffering
		err := json.Unmarshal(d.Body, &result)
		assert.NoError(t, err)
		assert.Equal(t, offeringID, result.ID)
		// Verify Enrichment
		assert.NotNil(t, result.ProductSpecification)
		assert.Equal(t, "Advanced Feature Spec", result.ProductSpecification.Name)
		assert.Len(t, result.Categories, 1)
		assert.Equal(t, "Advanced Category", result.Categories[0].Name)

	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for get response")
	}
}
