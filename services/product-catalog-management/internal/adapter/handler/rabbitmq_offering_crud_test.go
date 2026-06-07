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
	"tmf/services/product-catalog-management/internal/usecase/offering"
)

func TestRabbitMQHandler_Offering_CRUD(t *testing.T) {
	repo := repository.NewProductOfferingRepo(sharedDB)
	specRepo := repository.NewProductSpecificationRepo(sharedDB)

	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	err = sharedPub.DeclareTopicExchange("catalog_events", true, false, false, false)
	require.NoError(t, err)

	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createUC := offering.NewCreateProductOffering(repo, specRepo, pub, &repository.NoOpTransactionManager{})
	updateUC := offering.NewUpdateProductOfferingUseCase(repo, specRepo, pub, &repository.NoOpTransactionManager{})
	deleteUC := offering.NewDeleteProductOfferingUseCase(repo, pub, &repository.NoOpTransactionManager{})
	getUC := offering.NewGetProductOffering(repo, specRepo, repository.NewCategoryRepo(sharedDB))
	listUC := offering.NewListProductOfferings(repo)

	h := handler.NewRabbitMQHandler(
		sharedPub,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		createUC,
		updateUC,
		deleteUC,
		getUC,
		listUC,
	)
	bindTestHandler(t, h)

	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	off := &domain.ProductOffering{Name: "Seed Offering", LifecycleStatus: "Active", IsSellable: true}
	err = repo.Create(context.Background(), off)
	require.NoError(t, err)

	updateMsg := map[string]any{"id": off.ID, "name": "Updated Offering"}
	body, _ := json.Marshal(updateMsg)
	err = ch.Publish("catalog_events", "cmd.catalog.offering.update", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		c, _ := repo.Get(context.Background(), off.ID)
		return c != nil && c.Name == "Updated Offering"
	}, 10*time.Second, 100*time.Millisecond)

	replyQueue, _ := ch.QueueDeclare("", false, true, true, false, nil)
	msgs, _ := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)

	getMsg := map[string]any{"id": off.ID}
	body, _ = json.Marshal(getMsg)
	err = ch.Publish("catalog_events", "query.catalog.offering.get", false, false, amqp.Publishing{ContentType: "application/json", Body: body, ReplyTo: replyQueue.Name, CorrelationId: "get-off"})
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.Equal(t, "get-off", msg.CorrelationId)
		var resp domain.ProductOffering
		_ = json.Unmarshal(msg.Body, &resp)
		assert.Equal(t, "Updated Offering", resp.Name)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for GET reply")
	}

	listMsg := map[string]any{"name": "Updated Offering"}
	body, _ = json.Marshal(listMsg)
	err = ch.Publish("catalog_events", "query.catalog.offering.list", false, false, amqp.Publishing{ContentType: "application/json", Body: body, ReplyTo: replyQueue.Name, CorrelationId: "list-off"})
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.Equal(t, "list-off", msg.CorrelationId)
		var resp []domain.ProductOffering
		_ = json.Unmarshal(msg.Body, &resp)
		assert.GreaterOrEqual(t, len(resp), 1)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for LIST reply")
	}

	deleteMsg := map[string]any{"id": off.ID}
	body, _ = json.Marshal(deleteMsg)
	err = ch.Publish("catalog_events", "cmd.catalog.offering.delete", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		c, err := repo.Get(context.Background(), off.ID)
		return c == nil && err != nil
	}, 10*time.Second, 100*time.Millisecond)
}

func TestRabbitMQHandler_Offering_Errors(t *testing.T) {
	repo := repository.NewProductOfferingRepo(sharedDB)
	specRepo := repository.NewProductSpecificationRepo(sharedDB)
	sharedPub, _ := rabbitmq.NewPublisherWithConnection(rabbitConn)
	_ = sharedPub.DeclareTopicExchange("catalog_events", true, false, false, false)
	pub, _ := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")

	createUC := offering.NewCreateProductOffering(repo, specRepo, pub, &repository.NoOpTransactionManager{})
	updateUC := offering.NewUpdateProductOfferingUseCase(repo, specRepo, pub, &repository.NoOpTransactionManager{})
	deleteUC := offering.NewDeleteProductOfferingUseCase(repo, pub, &repository.NoOpTransactionManager{})
	getUC := offering.NewGetProductOffering(repo, specRepo, repository.NewCategoryRepo(sharedDB))
	listUC := offering.NewListProductOfferings(repo)

	h := handler.NewRabbitMQHandler(
		sharedPub,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		createUC, updateUC, deleteUC, getUC, listUC,
	)
	bindTestHandler(t, h)

	ch, _ := rabbitConn.Channel()
	defer func() { _ = ch.Close() }()

	_ = ch.Publish("catalog_events", "cmd.catalog.offering.create", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "cmd.catalog.offering.update", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "cmd.catalog.offering.delete", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "query.catalog.offering.get", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "query.catalog.offering.list", false, false, amqp.Publishing{Body: []byte("{invalid")})

	_ = ch.Publish("catalog_events", "cmd.catalog.offering.create", false, false, amqp.Publishing{Body: []byte("{}")})
	_ = ch.Publish("catalog_events", "cmd.catalog.offering.update", false, false, amqp.Publishing{Body: []byte(`{"id":"non-existent"}`)})
	_ = ch.Publish("catalog_events", "cmd.catalog.offering.delete", false, false, amqp.Publishing{Body: []byte(`{"id":"non-existent"}`)})
	_ = ch.Publish("catalog_events", "query.catalog.offering.get", false, false, amqp.Publishing{Body: []byte(`{"id":"non-existent"}`)})

	time.Sleep(2 * time.Second)
}
