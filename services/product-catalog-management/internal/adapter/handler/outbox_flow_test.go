package handler_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tmf/services/product-catalog-management/internal/adapter/publisher"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/adapter/worker"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
	"tmf/services/product-catalog-management/internal/usecase/offering"
)

func TestOutboxFlow_CreateOffering(t *testing.T) {
	// 1. Setup Dependencies with Real Components
	// Ensure DB schema is up to date for this test
	err := sharedDB.AutoMigrate(&repository.OutboxEventModel{}, &repository.ProductOfferingModel{}) // ensure offering model too
	require.NoError(t, err)

	sharedDB.Exec("TRUNCATE TABLE product_offerings CASCADE")
	sharedDB.Exec("DELETE FROM outbox_event_models")

	offeringRepo := repository.NewProductOfferingRepo(sharedDB)
	tm := repository.NewTransactionManager(sharedDB)
	outboxPub := publisher.NewOutboxPublisher(sharedDB)

	rabbitPub, err := publisher.NewRabbitMQPublisher(rabbitConn)
	require.NoError(t, err)

	outboxWorker := worker.NewOutboxWorker(sharedDB, rabbitPub)

	createUC := offering.NewCreateProductOffering(offeringRepo, outboxPub, tm)

	// 2. Prepare RabbitMQ Consumer to verify delivery
	ch, err := rabbitConn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// Declare exchange (in case publisher hasn't yet, though it should)
	err = ch.ExchangeDeclare("catalog_events", "topic", true, false, false, false, nil)
	require.NoError(t, err)

	// Declare a queue to bind
	q, err := ch.QueueDeclare(
		"",    // name (random)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	require.NoError(t, err)

	queryRoutingKey := "evt.catalog.offering.created"
	err = ch.QueueBind(
		q.Name,           // queue name
		queryRoutingKey,  // routing key
		"catalog_events", // exchange
		false,
		nil,
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

	// 3. Execute Use Case
	ctx := context.Background()
	specID := "spec-123"
	input := ports.CreateProductOfferingInput{
		Name:            "Outbox Offering",
		Description:     "Testing Transactional Outbox",
		IsBundle:        false,
		IsSellable:      true,
		LifecycleStatus: "Active",
		ProductSpecID:   &specID,
		Prices:          []domain.ProductOfferingPrice{},
	}

	res, err := createUC.Execute(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, res)

	// 4. Verify ATOMICITY (DB State before Worker runs)
	// Offering should exist
	repoOffering, err := offeringRepo.Get(ctx, res.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Outbox Offering", repoOffering.Name)

	// Outbox Event should exist and be PENDING
	var outboxEvent repository.OutboxEventModel
	err = sharedDB.Where("status = ?", repository.StatusPending).First(&outboxEvent).Error
	assert.NoError(t, err, "Should have PENDING outbox event")
	assert.Equal(t, "evt.catalog.offering.created", outboxEvent.RoutingKey)

	// 5. Start Worker and Verify Delivery
	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go outboxWorker.Start(workerCtx)

	// Wait for message
	select {
	case msg := <-msgs:
		t.Logf("Received message: %s", msg.Body)
		var event domain.ProductOfferingCreatedEvent
		err := json.Unmarshal(msg.Body, &event)
		assert.NoError(t, err)
		assert.Equal(t, res.ID, event.ProductOffering.ID)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message from RabbitMQ")
	}

	// 6. Verify Outbox Status Updated
	time.Sleep(100 * time.Millisecond) // Give time for DB update
	err = sharedDB.Where("id = ?", outboxEvent.ID).First(&outboxEvent).Error
	assert.NoError(t, err)
	assert.Equal(t, repository.StatusPublished, outboxEvent.Status)
}
