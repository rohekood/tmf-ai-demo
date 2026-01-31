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

func TestRabbitMQHandler_Reply(t *testing.T) {
	// Setup Dependencies
	repo := repository.NewCatalogRepo(sharedDB)
	// Create Publisher
	// Create Publisher
	// Create Publisher
	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createUC := catalog.NewCreateCatalog(repo, pub, &repository.NoOpTransactionManager{})
	listUC := catalog.NewListCatalogs(repo)

	// Init Handler
	h, err := handler.NewRabbitMQHandler(
		rabbitConn,
		createUC,
		nil, nil,
		listUC,
		nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Handler
	go func() { _ = h.Start(ctx) }()
	time.Sleep(1 * time.Second)

	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	// Create a temporary reply queue
	q, err := ch.QueueDeclare(
		"",    // name (empty = auto-generated)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	require.NoError(t, err)

	// Consume from reply queue
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

	// Publish Command with ReplyTo
	cmd := domain.CatalogCreateEvent{
		Name:        "RPC Catalog",
		Description: "Testing RPC Reply",
	}
	body, _ := json.Marshal(cmd)
	corrId := "123456789"

	err = ch.Publish(
		"catalog_events",             // exchange
		"cmd.catalog.catalog.create", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			ReplyTo:       q.Name,
			CorrelationId: corrId,
		},
	)
	require.NoError(t, err)

	// Wait for response
	select {
	case d := <-msgs:
		assert.Equal(t, corrId, d.CorrelationId)
		var response map[string]interface{}
		err := json.Unmarshal(d.Body, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "id")
		assert.Equal(t, "RPC Catalog", response["name"])
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for RPC reply")
	}
}
