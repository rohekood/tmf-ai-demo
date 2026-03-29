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
	"tmf/services/product-catalog-management/internal/usecase/category"
)

func TestRabbitMQHandler_Category_CRUD(t *testing.T) {
	repo := repository.NewCategoryRepo(sharedDB)

	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	err = sharedPub.DeclareTopicExchange("catalog_events", true, false, false, false)
	require.NoError(t, err)

	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createUC := category.NewCreateCategory(repo, pub, &repository.NoOpTransactionManager{})
	updateUC := category.NewUpdateCategoryUseCase(repo, pub, &repository.NoOpTransactionManager{})
	deleteUC := category.NewDeleteCategoryUseCase(repo, pub, &repository.NoOpTransactionManager{})
	getUC := category.NewGetCategory(repo)
	listUC := category.NewListCategories(repo)

	h, err := handler.NewRabbitMQHandler(
		rabbitConn,
		nil, nil, nil, nil, nil,
		createUC,
		updateUC,
		deleteUC,
		getUC,
		listUC,
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

	cat := &domain.Category{Name: "Seed Category", IsRoot: true}
	err = repo.Create(ctx, cat)
	require.NoError(t, err)

	updateMsg := map[string]interface{}{
		"id":   cat.ID,
		"name": "Updated Category",
	}
	body, _ := json.Marshal(updateMsg)
	err = ch.Publish("catalog_events", "cmd.catalog.category.update", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		c, _ := repo.Get(context.Background(), cat.ID)
		return c != nil && c.Name == "Updated Category"
	}, 10*time.Second, 100*time.Millisecond)

	replyQueue, _ := ch.QueueDeclare("", false, true, true, false, nil)
	msgs, _ := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)

	getMsg := map[string]interface{}{"id": cat.ID}
	body, _ = json.Marshal(getMsg)
	err = ch.Publish("catalog_events", "query.catalog.category.get", false, false, amqp.Publishing{ContentType: "application/json", Body: body, ReplyTo: replyQueue.Name, CorrelationId: "get-cat"})
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.Equal(t, "get-cat", msg.CorrelationId)
		var resp domain.Category
		_ = json.Unmarshal(msg.Body, &resp)
		assert.Equal(t, "Updated Category", resp.Name)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for GET reply")
	}

	listMsg := map[string]interface{}{"name": "Updated Category"}
	body, _ = json.Marshal(listMsg)
	err = ch.Publish("catalog_events", "query.catalog.category.list", false, false, amqp.Publishing{ContentType: "application/json", Body: body, ReplyTo: replyQueue.Name, CorrelationId: "list-cat"})
	require.NoError(t, err)

	select {
	case msg := <-msgs:
		assert.Equal(t, "list-cat", msg.CorrelationId)
		var resp []domain.Category
		_ = json.Unmarshal(msg.Body, &resp)
		assert.GreaterOrEqual(t, len(resp), 1)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for LIST reply")
	}

	deleteMsg := map[string]interface{}{"id": cat.ID}
	body, _ = json.Marshal(deleteMsg)
	err = ch.Publish("catalog_events", "cmd.catalog.category.delete", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		c, err := repo.Get(context.Background(), cat.ID)
		return c == nil && err != nil
	}, 10*time.Second, 100*time.Millisecond)
}

func TestRabbitMQHandler_Category_Errors(t *testing.T) {
	repo := repository.NewCategoryRepo(sharedDB)
	sharedPub, _ := rabbitmq.NewPublisherWithConnection(rabbitConn)
	_ = sharedPub.DeclareTopicExchange("catalog_events", true, false, false, false)
	pub, _ := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")

	createUC := category.NewCreateCategory(repo, pub, &repository.NoOpTransactionManager{})
	updateUC := category.NewUpdateCategoryUseCase(repo, pub, &repository.NoOpTransactionManager{})
	deleteUC := category.NewDeleteCategoryUseCase(repo, pub, &repository.NoOpTransactionManager{})
	getUC := category.NewGetCategory(repo)
	listUC := category.NewListCategories(repo)

	h, _ := handler.NewRabbitMQHandler(
		rabbitConn,
		nil, nil, nil, nil, nil,
		createUC, updateUC, deleteUC, getUC, listUC,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = h.Start(ctx) }()
	time.Sleep(1 * time.Second)

	ch, _ := rabbitConn.Channel()
	defer func() { _ = ch.Close() }()

	_ = ch.Publish("catalog_events", "cmd.catalog.category.create", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "cmd.catalog.category.update", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "cmd.catalog.category.delete", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "query.catalog.category.get", false, false, amqp.Publishing{Body: []byte("{invalid")})
	_ = ch.Publish("catalog_events", "query.catalog.category.list", false, false, amqp.Publishing{Body: []byte("{invalid")})

	_ = ch.Publish("catalog_events", "cmd.catalog.category.create", false, false, amqp.Publishing{Body: []byte("{}")})
	_ = ch.Publish("catalog_events", "cmd.catalog.category.update", false, false, amqp.Publishing{Body: []byte(`{"id":"non-existent"}`)})
	_ = ch.Publish("catalog_events", "cmd.catalog.category.delete", false, false, amqp.Publishing{Body: []byte(`{"id":"non-existent"}`)})
	_ = ch.Publish("catalog_events", "query.catalog.category.get", false, false, amqp.Publishing{Body: []byte(`{"id":"non-existent"}`)})

	time.Sleep(2 * time.Second)
}
