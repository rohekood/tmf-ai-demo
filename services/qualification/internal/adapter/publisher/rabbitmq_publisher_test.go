package publisher_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"tmf/services/qualification/internal/adapter/publisher"
	"tmf/services/qualification/internal/core/domain"
)

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, exchange, routingKey string, payload any) error {
	args := m.Called(ctx, exchange, routingKey, payload)
	return args.Error(0)
}

func (m *MockPublisher) PublishToQueue(ctx context.Context, queueName string, correlationID string, body any) error {
	args := m.Called(ctx, queueName, correlationID, body)
	return args.Error(0)
}

func (m *MockPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	args := m.Called(name, durable, autoDelete, internal, noWait)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestEventPublisher(t *testing.T) {
	mockPub := new(MockPublisher)
	pub := publisher.NewEventPublisher(mockPub, "ex.test")

	result := domain.EligibilityResult{SessionID: "sess-1"}

	mockPub.On("Publish", mock.Anything, "ex.test", "evt.qual.checked", result).Return(nil)

	err := pub.PublishEligibilityChecked(context.Background(), result)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}
