package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"tmf/pkg/rabbitmq"
	"tmf/services/pocv/internal/core/domain"
)

type mockConsumer struct {
	subscribe func(topic string, handler rabbitmq.ConsumerHandler) error
}

func (m *mockConsumer) Subscribe(topic string, h rabbitmq.ConsumerHandler) error {
	if m.subscribe != nil {
		return m.subscribe(topic, h)
	}
	return nil
}

func (m *mockConsumer) Close() error { return nil }

func TestRPCHandlerHandleGetSaga(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Success", func(t *testing.T) {
		saga := &domain.SagaInstance{ID: "s1", CartID: "c1", Status: domain.SagaStatusPending}

		uc := &mockSagaUseCase{
			getSaga: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
				if id == "s1" {
					return saga, nil
				}
				return nil, errors.New("not found")
			},
		}

		var publishedBody any
		pub := &mockPublisher{
			publishToQueue: func(ctx context.Context, queueName, correlationID string, body any) error {
				publishedBody = body
				if queueName != "reply_queue" {
					t.Errorf("Expected reply_queue, got %s", queueName)
				}
				if correlationID != "cor-1" {
					t.Errorf("Expected cor-1, got %s", correlationID)
				}
				return nil
			},
		}

		h := NewRPCHandler(uc, pub, logger)
		ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "reply_queue")
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "cor-1")

		err := h.HandleGetSaga(ctx, []byte(`{"sagaId":"s1"}`))
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if publishedBody != saga {
			t.Errorf("Expected saga instance to be published, got %v", publishedBody)
		}
	})

	t.Run("NotFound returns error response", func(t *testing.T) {
		uc := &mockSagaUseCase{
			getSaga: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
				return nil, errors.New("record not found")
			},
		}

		var publishedBody any
		pub := &mockPublisher{
			publishToQueue: func(ctx context.Context, queueName, correlationID string, body any) error {
				publishedBody = body
				return nil
			},
		}

		h := NewRPCHandler(uc, pub, logger)
		ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "reply_queue")
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "cor-1")

		err := h.HandleGetSaga(ctx, []byte(`{"sagaId":"unknown"}`))
		if err != nil {
			t.Fatalf("Expected no error (error sent as reply), got %v", err)
		}

		errResp, ok := publishedBody.(map[string]string)
		if !ok {
			t.Fatalf("Expected map[string]string error response, got %T", publishedBody)
		}
		if errResp["error"] != "saga not found" {
			t.Errorf("Expected 'saga not found', got %s", errResp["error"])
		}
	})

	t.Run("InvalidJSON returns error", func(t *testing.T) {
		h := NewRPCHandler(&mockSagaUseCase{}, &mockPublisher{}, logger)
		err := h.HandleGetSaga(context.Background(), []byte(`{invalid`))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("MissingReplyTo returns error", func(t *testing.T) {
		saga := &domain.SagaInstance{ID: "s1"}
		uc := &mockSagaUseCase{
			getSaga: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
				return saga, nil
			},
		}

		h := NewRPCHandler(uc, &mockPublisher{}, logger)
		err := h.HandleGetSaga(context.Background(), []byte(`{"sagaId":"s1"}`))
		if err == nil {
			t.Error("Expected error when replyTo is missing")
		}
	})

	t.Run("MissingReplyToOnNotFound returns error", func(t *testing.T) {
		uc := &mockSagaUseCase{
			getSaga: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
				return nil, errors.New("not found")
			},
		}

		h := NewRPCHandler(uc, &mockPublisher{}, logger)
		err := h.HandleGetSaga(context.Background(), []byte(`{"sagaId":"s1"}`))
		if err == nil {
			t.Error("Expected error when replyTo is missing and saga not found")
		}
	})
}

func TestRPCHandlerBindRPCHandlers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Success", func(t *testing.T) {
		h := NewRPCHandler(&mockSagaUseCase{}, &mockPublisher{}, logger)
		cons := &mockConsumer{}
		err := h.BindRPCHandlers(cons)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	t.Run("SubscribeError", func(t *testing.T) {
		h := NewRPCHandler(&mockSagaUseCase{}, &mockPublisher{}, logger)
		cons := &mockConsumer{
			subscribe: func(topic string, handler rabbitmq.ConsumerHandler) error {
				return errors.New("subscribe failed")
			},
		}
		err := h.BindRPCHandlers(cons)
		if err == nil {
			t.Error("Expected error when subscribe fails")
		}
	})

	t.Run("BindsCorrectRoutingKey", func(t *testing.T) {
		h := NewRPCHandler(&mockSagaUseCase{}, &mockPublisher{}, logger)
		var subscribedKey string
		cons := &mockConsumer{
			subscribe: func(topic string, handler rabbitmq.ConsumerHandler) error {
				subscribedKey = topic
				return nil
			},
		}
		if err := h.BindRPCHandlers(cons); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if subscribedKey != "query.pocv.saga.get" {
			t.Errorf("Expected routing key 'query.pocv.saga.get', got '%s'", subscribedKey)
		}
	})
}
