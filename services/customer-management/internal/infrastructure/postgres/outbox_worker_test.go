package postgres

import (
	"context"
	"encoding/json"
	"errors"
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
	mockPub.On("Publish", mock.Anything, "customer.events", "test.pub.error", mock.Anything).Return(errors.New("publish error"))
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

	// Test Publish error
	mockPub.On("Publish", mock.Anything, "customer.events", "test.pub.error", mock.Anything).Return(errors.New("publish error"))
	eventPubErr := &domain.OutboxEvent{
		ID:         uuid.New().String(),
		RoutingKey: "test.pub.error",
		Payload:    []byte(`{"foo":"bar"}`),
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}
	require.NoError(t, repo.Save(ctx, eventPubErr))

	// Test Invalid Headers JSON
	// We can't insert invalid JSON into JSON column, but if headers is a bytea/text, we could.
	// Actually headers is JSON column? Let's assume it is.
	// We will skip testing invalid JSON unmarshal for Postgres since postgres enforces it.

	worker.processEvents(ctx)

	var updatedPubErr domain.OutboxEvent
	sharedDB.First(&updatedPubErr, "id = ?", eventPubErr.ID)
	assert.Equal(t, "PENDING", updatedPubErr.Status) // Publish failed
}

func TestOutboxWorker_StartStop(t *testing.T) {
	if sharedDB == nil {
		t.Skip("Shared DB not initialized")
	}
	repo := NewOutboxRepository(sharedDB)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockPub := new(MockPublisher)

	// mockPub might be called if there are pending events from other tests, so we allow it
	mockPub.On("Publish", mock.Anything, "customer.events", mock.Anything, mock.Anything).Return(nil)

	worker := NewOutboxWorker(repo, mockPub, logger)

	ctx, cancel := context.WithCancel(context.Background())

	// Test stopping via context
	go worker.Start(ctx)
	time.Sleep(100 * time.Millisecond) // Let it run a bit
	cancel()                           // This should stop the worker
	time.Sleep(100 * time.Millisecond) // Wait for it to stop

	// Test stopping via Stop() method
	worker2 := NewOutboxWorker(repo, mockPub, logger)
	ctx2 := context.Background()

	go worker2.Start(ctx2)
	time.Sleep(100 * time.Millisecond)
	worker2.Stop()
	time.Sleep(100 * time.Millisecond)
}
