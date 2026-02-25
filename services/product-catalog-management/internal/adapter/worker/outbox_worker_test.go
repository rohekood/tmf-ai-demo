package worker_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tmf/services/product-catalog-management/internal/adapter/publisher"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/adapter/worker"
)

type MockRabbitMQPublisher struct {
	mock.Mock
}

func (m *MockRabbitMQPublisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	args := m.Called(ctx, exchange, routingKey, body)
	return args.Error(0)
}

func (m *MockRabbitMQPublisher) PublishToQueue(ctx context.Context, queueName string, correlationID string, body interface{}) error {
	args := m.Called(ctx, queueName, correlationID, body)
	return args.Error(0)
}

func (m *MockRabbitMQPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	args := m.Called(name, durable, autoDelete, internal, noWait)
	return args.Error(0)
}

func (m *MockRabbitMQPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestOutboxWorker_ProcessEvent(t *testing.T) {
	mockPub := new(MockRabbitMQPublisher)
	rabbitPub, _ := publisher.NewRabbitMQPublisher(mockPub, "test_exchange")
	
	w := worker.NewOutboxWorker(sharedDB, rabbitPub)
	
	ctx := context.Background()

	// Add an event to DB
	headers := map[string]string{"user": "test-user"}
	headersBytes, _ := json.Marshal(headers)
	event := &repository.OutboxEventModel{
		RoutingKey: "test.key",
		Payload:    json.RawMessage(`{"hello":"world"}`),
		Headers:    headersBytes,
		Status:     repository.StatusPending,
	}
	sharedDB.Create(event)

	// Expect Publish to be called
	mockPub.On("Publish", mock.Anything, "test_exchange", "test.key", mock.Anything).Return(nil).Once()

	// Process batch (which will fetch the event and process it)
	// We need to call processBatch, but it's private.
	// We can use reflection or test it by calling Start with a timeout, or we can just make it public. Let's start context and cancel it.

	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	
	go w.Start(ctxTimeout)
	time.Sleep(1500 * time.Millisecond)

	// Check DB status
	var updatedEvent repository.OutboxEventModel
	sharedDB.First(&updatedEvent, "id = ?", event.ID)

	assert.Equal(t, repository.StatusPublished, updatedEvent.Status)
	assert.NotNil(t, updatedEvent.ProcessedAt)
	mockPub.AssertExpectations(t)
}

func TestOutboxWorker_ProcessEvent_Errors(t *testing.T) {
	mockPub := new(MockRabbitMQPublisher)
	rabbitPub, _ := publisher.NewRabbitMQPublisher(mockPub, "test_exchange")
	
	w := worker.NewOutboxWorker(sharedDB, rabbitPub)
	
	ctx := context.Background()

	// 1. Invalid Headers
	eventInvalidHeaders := &repository.OutboxEventModel{
		RoutingKey: "test.invalid.headers",
		Payload:    json.RawMessage(`{}`),
		Headers:    []byte(`"invalid string json"`),
		Status:     repository.StatusPending,
	}
	sharedDB.Create(eventInvalidHeaders)
	mockPub.On("Publish", mock.Anything, "test_exchange", "test.invalid.headers", mock.Anything).Return(nil).Once()

	// 2. Publish fails
	eventPublishFail := &repository.OutboxEventModel{
		RoutingKey: "test.publish.fail",
		Payload:    json.RawMessage(`{}`),
		Status:     repository.StatusPending,
	}
	sharedDB.Create(eventPublishFail)
	mockPub.On("Publish", mock.Anything, "test_exchange", "test.publish.fail", mock.Anything).Return(assert.AnError).Once()

	// Process batch via Start
	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	
	go w.Start(ctxTimeout)
	time.Sleep(1500 * time.Millisecond)

	var invalidHdrEvent repository.OutboxEventModel
	sharedDB.First(&invalidHdrEvent, "id = ?", eventInvalidHeaders.ID)
	assert.Equal(t, repository.StatusPublished, invalidHdrEvent.Status)

	var failEvent repository.OutboxEventModel
	sharedDB.First(&failEvent, "id = ?", eventPublishFail.ID)
	assert.Equal(t, repository.StatusFailed, failEvent.Status)

	mockPub.AssertExpectations(t)
}
