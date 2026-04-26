package handler_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/qualification/internal/adapter/handler"
	"tmf/services/qualification/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.QualificationSession) (string, error) {
	args := m.Called(ctx, session)
	return args.String(0), args.Error(1)
}

func (m *MockSessionRepository) Get(ctx context.Context, sessionID string) (*domain.QualificationSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.QualificationSession), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *domain.QualificationSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) FindExpired(ctx context.Context) ([]*domain.QualificationSession, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.QualificationSession), args.Error(1)
}

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

type MockConsumer struct {
	mock.Mock
}

func (m *MockConsumer) Subscribe(topic string, handler rabbitmq.ConsumerHandler) error {
	args := m.Called(topic, handler)
	return args.Error(0)
}

func (m *MockConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRPCHandler(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	t.Run("HandleGetSession - Success", func(t *testing.T) {
		repo := new(MockSessionRepository)
		pub := new(MockPublisher)
		h := handler.NewRPCHandler(repo, pub, logger)

		ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "replyQ")
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr1")

		session := &domain.QualificationSession{
			ID:        "sess-1",
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		}
		repo.On("Get", ctx, "sess-1").Return(session, nil)
		pub.On("PublishToQueue", ctx, "replyQ", "corr1", session).Return(nil)

		payload := []byte(`{"sessionId":"sess-1"}`)
		err := h.HandleGetSession(ctx, payload)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		pub.AssertExpectations(t)
	})

	t.Run("HandleGetSession - Expired", func(t *testing.T) {
		repo := new(MockSessionRepository)
		pub := new(MockPublisher)
		h := handler.NewRPCHandler(repo, pub, logger)

		ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "replyQ")
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr1")

		session := &domain.QualificationSession{
			ID:        "sess-1",
			ExpiresAt: time.Now().UTC().Add(-1 * time.Hour),
		}
		repo.On("Get", ctx, "sess-1").Return(session, nil)
		pub.On("PublishToQueue", ctx, "replyQ", "corr1", mock.Anything).Return(nil)

		payload := []byte(`{"sessionId":"sess-1"}`)
		err := h.HandleGetSession(ctx, payload)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		pub.AssertExpectations(t)
	})

	t.Run("HandleGetSession - Not Found", func(t *testing.T) {
		repo := new(MockSessionRepository)
		pub := new(MockPublisher)
		h := handler.NewRPCHandler(repo, pub, logger)

		ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "replyQ")
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr1")

		repo.On("Get", ctx, "sess-1").Return(nil, errors.New("not found"))
		pub.On("PublishToQueue", ctx, "replyQ", "corr1", mock.Anything).Return(nil)

		payload := []byte(`{"sessionId":"sess-1"}`)
		err := h.HandleGetSession(ctx, payload)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		pub.AssertExpectations(t)
	})

	t.Run("HandleGetSession - Invalid JSON", func(t *testing.T) {
		h := handler.NewRPCHandler(nil, nil, logger)
		err := h.HandleGetSession(context.Background(), []byte(`{`))
		assert.Error(t, err)
	})

	t.Run("HandleGetSession - Missing Context", func(t *testing.T) {
		repo := new(MockSessionRepository)
		h := handler.NewRPCHandler(repo, nil, logger)

		session := &domain.QualificationSession{
			ID:        "sess-1",
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		}
		repo.On("Get", mock.Anything, "sess-1").Return(session, nil)

		payload := []byte(`{"sessionId":"sess-1"}`)
		err := h.HandleGetSession(context.Background(), payload)
		assert.Error(t, err)
	})

	t.Run("replyError - Missing Context", func(t *testing.T) {
		repo := new(MockSessionRepository)
		h := handler.NewRPCHandler(repo, nil, logger)

		repo.On("Get", mock.Anything, "sess-1").Return(nil, errors.New("not found"))

		payload := []byte(`{"sessionId":"sess-1"}`)
		err := h.HandleGetSession(context.Background(), payload)
		assert.Error(t, err)
	})

	t.Run("BindRPCHandlers", func(t *testing.T) {
		h := handler.NewRPCHandler(nil, nil, logger)
		cons := new(MockConsumer)
		cons.On("Subscribe", "query.qual.session.get", mock.Anything).Return(nil)

		err := h.BindRPCHandlers(cons)
		assert.NoError(t, err)
	})

	t.Run("BindRPCHandlers Error", func(t *testing.T) {
		h := handler.NewRPCHandler(nil, nil, logger)
		cons := new(MockConsumer)
		cons.On("Subscribe", "query.qual.session.get", mock.Anything).Return(errors.New("sub error"))

		err := h.BindRPCHandlers(cons)
		assert.Error(t, err)
	})
}
