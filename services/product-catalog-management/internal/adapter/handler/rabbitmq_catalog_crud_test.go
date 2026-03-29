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
)

func TestRabbitMQHandler_Catalog_CRUD(t *testing.T) {
	repo := repository.NewCatalogRepo(sharedDB)

	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	err = sharedPub.DeclareTopicExchange("catalog_events", true, false, false, false)
	require.NoError(t, err)

	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createUC := catalog.NewCreateCatalog(repo, pub, &repository.NoOpTransactionManager{})
	updateUC := catalog.NewUpdateCatalogUseCase(repo, pub, &repository.NoOpTransactionManager{})
	deleteUC := catalog.NewDeleteCatalogUseCase(repo, pub, &repository.NoOpTransactionManager{})
	getUC := catalog.NewGetCatalog(repo)
	listUC := catalog.NewListCatalogs(repo)

	h, err := handler.NewRabbitMQHandler(
		rabbitConn,
		createUC,
		updateUC,
		deleteUC,
		listUC,
		getUC,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = h.Start(ctx)
	}()

	time.Sleep(1 * time.Second)

	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	cat := &domain.Catalog{Name: "Seed Catalog"}
	err = repo.Create(ctx, cat)
	require.NoError(t, err)

	updateMsg := map[string]interface{}{
		"id":          cat.ID,
		"name":        "Updated Catalog",
		"description": "Updated via RabbitMQ",
	}
	body, _ := json.Marshal(updateMsg)
	err = ch.Publish("catalog_events", "cmd.catalog.catalog.update", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		list, _ := repo.List(context.Background(), map[string]interface{}{"name": "Updated Catalog"})
		return len(list) == 1
	}, 10*time.Second, 100*time.Millisecond)

	replyQueue, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	getMsg := map[string]interface{}{"id": cat.ID}
	body, _ = json.Marshal(getMsg)
	err = ch.Publish("catalog_events", "query.catalog.catalog.get", false, false, amqp.Publishing{
		ContentType:   "application/json",
		Body:          body,
		ReplyTo:       replyQueue.Name,
		CorrelationId: "get-123",
	})
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.Equal(t, "get-123", msg.CorrelationId)
		var resp domain.Catalog
		_ = json.Unmarshal(msg.Body, &resp)
		assert.Equal(t, "Updated Catalog", resp.Name)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for GET reply")
	}

	listMsg := map[string]interface{}{"name": "Updated Catalog"}
	body, _ = json.Marshal(listMsg)
	err = ch.Publish("catalog_events", "query.catalog.catalog.list", false, false, amqp.Publishing{
		ContentType:   "application/json",
		Body:          body,
		ReplyTo:       replyQueue.Name,
		CorrelationId: "list-123",
	})
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.Equal(t, "list-123", msg.CorrelationId)
		var resp []domain.Catalog
		_ = json.Unmarshal(msg.Body, &resp)
		assert.GreaterOrEqual(t, len(resp), 1)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for LIST reply")
	}

	deleteMsg := map[string]interface{}{"id": cat.ID}
	body, _ = json.Marshal(deleteMsg)
	err = ch.Publish("catalog_events", "cmd.catalog.catalog.delete", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		list, _ := repo.List(context.Background(), map[string]interface{}{"id": cat.ID})
		return len(list) == 0
	}, 10*time.Second, 100*time.Millisecond)
}

func TestRabbitMQHandler_Catalog_Errors(t *testing.T) {
	repo := repository.NewCatalogRepo(sharedDB)
	sharedPub, _ := rabbitmq.NewPublisherWithConnection(rabbitConn)
	_ = sharedPub.DeclareTopicExchange("catalog_events", true, false, false, false)
	pub, _ := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")

	createUC := catalog.NewCreateCatalog(repo, pub, &repository.NoOpTransactionManager{})
	updateUC := catalog.NewUpdateCatalogUseCase(repo, pub, &repository.NoOpTransactionManager{})
	deleteUC := catalog.NewDeleteCatalogUseCase(repo, pub, &repository.NoOpTransactionManager{})
	getUC := catalog.NewGetCatalog(repo)
	listUC := catalog.NewListCatalogs(repo)

	h, _ := handler.NewRabbitMQHandler(
		rabbitConn,
		createUC, updateUC, deleteUC, listUC, getUC,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = h.Start(ctx) }()
	time.Sleep(1 * time.Second)

	ch, _ := rabbitConn.Channel()
	defer func() { _ = ch.Close() }()

	_ = ch.Publish("catalog_events", "cmd.catalog.catalog.create", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "cmd.catalog.catalog.update", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "cmd.catalog.catalog.delete", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "query.catalog.catalog.get", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "query.catalog.catalog.list", false, false, amqp.Publishing{Body: []byte("{invalid")})

	_ = ch.Publish("catalog_events", "cmd.catalog.catalog.create", false, false, amqp.Publishing{Body: []byte("{}")})
	_ = ch.Publish("catalog_events", "cmd.catalog.catalog.update", false, false, amqp.Publishing{Body: []byte(`{"id":"non-existent"}`)})
	_ = ch.Publish("catalog_events", "cmd.catalog.catalog.delete", false, false, amqp.Publishing{Body: []byte(`{"id":"non-existent"}`)})
	_ = ch.Publish("catalog_events", "query.catalog.catalog.get", false, false, amqp.Publishing{Body: []byte(`{"id":"non-existent"}`)})

	time.Sleep(2 * time.Second)
}
