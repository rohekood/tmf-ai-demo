package postgres

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"
	"tmf/services/customer-management/internal/domain"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	args := m.Called(ctx, exchange, routingKey, body)
	return args.Error(0)
}
func (m *MockPublisher) PublishToQueue(ctx context.Context, queue, correlationID string, body interface{}) error {
	return nil
}
func (m *MockPublisher) Close() error {
	return nil
}
func (m *MockPublisher) GetChannel() *amqp.Channel {
	return nil
}
func (m *MockPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	return nil
}

func TestOutboxWorker_ProcessEvents(t *testing.T) {
	if sharedDB == nil {
		t.Fatal("Shared DB not initialized")
	}
	sharedDB.Exec("TRUNCATE TABLE outbox_events")

	repo := NewOutboxRepository(sharedDB)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockPub := new(MockPublisher)
	mockPub.On("Publish", mock.Anything, "customer.events", mock.Anything, mock.Anything).Return(nil)

	worker := NewOutboxWorker(repo, mockPub, logger)

	ctx := context.Background()

	// Seed an event
	payload := map[string]string{"foo": "bar"}
	payloadBytes, _ := json.Marshal(payload)
	event := &domain.OutboxEvent{
		ID:         uuid.New().String(),
		RoutingKey: "test.worker.event",
		Payload:    payloadBytes,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}
	require.NoError(t, repo.Save(ctx, event))

	// Run processEvents once manually (to avoid timing issues with Start loop)
	worker.processEvents(ctx)

	// Verify event is marked processed
	var updated domain.OutboxEvent
	err := sharedDB.First(&updated, "id = ?", event.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "PUBLISHED", updated.Status)
	assert.NotNil(t, updated.ProcessedAt)

	// Verify header handling (create another event with headers)
	headers := map[string]string{"user": "u1", "Authorization": "token"}
	headerBytes, _ := json.Marshal(headers)
	eventHeaders := &domain.OutboxEvent{
		ID:         uuid.New().String(),
		RoutingKey: "test.worker.headers",
		Payload:    payloadBytes,
		Headers:    headerBytes,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}
	require.NoError(t, repo.Save(ctx, eventHeaders))

	worker.processEvents(ctx)

	var updatedHeaders domain.OutboxEvent
	err = sharedDB.First(&updatedHeaders, "id = ?", eventHeaders.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "PUBLISHED", updatedHeaders.Status)
}
