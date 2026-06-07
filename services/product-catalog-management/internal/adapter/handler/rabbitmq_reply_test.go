package handler_test

import (
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
	repo := repository.NewCatalogRepo(sharedDB)

	sharedPub, err := rabbitmq.NewPublisherWithConnection(rabbitConn)
	require.NoError(t, err)
	pub, err := publisher.NewRabbitMQPublisher(sharedPub, "catalog_events")
	require.NoError(t, err)

	createUC := catalog.NewCreateCatalog(repo, pub, &repository.NoOpTransactionManager{})
	listUC := catalog.NewListCatalogs(repo)

	h := handler.NewRabbitMQHandler(
		sharedPub,
		createUC, nil, nil, listUC, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
	)
	bindTestHandler(t, h)

	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	require.NoError(t, err)

	cmd := domain.CatalogCreateEvent{Name: "RPC Catalog", Description: "Testing RPC Reply"}
	body, _ := json.Marshal(cmd)
	corrId := "123456789"

	err = ch.Publish("catalog_events", "cmd.catalog.catalog.create", false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			ReplyTo:       q.Name,
			CorrelationId: corrId,
		})
	require.NoError(t, err)

	select {
	case d := <-msgs:
		assert.Equal(t, corrId, d.CorrelationId)
		var response map[string]any
		err := json.Unmarshal(d.Body, &response)
		require.NoError(t, err)
		assert.Contains(t, response, "id")
		assert.Equal(t, "RPC Catalog", response["name"])
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for RPC reply")
	}
}
