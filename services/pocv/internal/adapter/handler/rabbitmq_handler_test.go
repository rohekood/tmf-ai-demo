package handler

import (
	"context"
	"errors"
	"testing"

	"tmf/pkg/rabbitmq"
	"tmf/services/pocv/internal/core/domain"
)

type mockSagaUseCase struct {
	startSaga               func(ctx context.Context, cartID string) error
	handleInventoryReserved func(ctx context.Context, sagaID string) error
	handleInventoryFailed   func(ctx context.Context, sagaID string) error
	handlePaymentAuthorized func(ctx context.Context, sagaID string) error
	handlePaymentDeclined   func(ctx context.Context, sagaID string) error
	handleOrderCreated      func(ctx context.Context, sagaID string) error
	getSaga                 func(ctx context.Context, id string) (*domain.SagaInstance, error)
}

func (m *mockSagaUseCase) StartSaga(ctx context.Context, cartID string) error {
	if m.startSaga != nil {
		return m.startSaga(ctx, cartID)
	}
	return nil
}

func (m *mockSagaUseCase) HandleInventoryReserved(ctx context.Context, sagaID string) error {
	if m.handleInventoryReserved != nil {
		return m.handleInventoryReserved(ctx, sagaID)
	}
	return nil
}

func (m *mockSagaUseCase) HandleInventoryFailed(ctx context.Context, sagaID string) error {
	if m.handleInventoryFailed != nil {
		return m.handleInventoryFailed(ctx, sagaID)
	}
	return nil
}

func (m *mockSagaUseCase) HandlePaymentAuthorized(ctx context.Context, sagaID string) error {
	if m.handlePaymentAuthorized != nil {
		return m.handlePaymentAuthorized(ctx, sagaID)
	}
	return nil
}

func (m *mockSagaUseCase) HandlePaymentDeclined(ctx context.Context, sagaID string) error {
	if m.handlePaymentDeclined != nil {
		return m.handlePaymentDeclined(ctx, sagaID)
	}
	return nil
}

func (m *mockSagaUseCase) HandleOrderCreated(ctx context.Context, sagaID string) error {
	if m.handleOrderCreated != nil {
		return m.handleOrderCreated(ctx, sagaID)
	}
	return nil
}

func (m *mockSagaUseCase) GetSaga(ctx context.Context, id string) (*domain.SagaInstance, error) {
	if m.getSaga != nil {
		return m.getSaga(ctx, id)
	}
	return nil, nil
}

func TestHandleSagaEvent(t *testing.T) {
	uc := &mockSagaUseCase{}
	h := NewRabbitMQHandler(uc, &mockPublisher{})

	tests := []struct {
		name       string
		routingKey string
		payload    []byte
	}{
		{"OrderCheckoutSubmit", rabbitmq.CmdOrderCheckoutSubmit, []byte(`{"cartId": "c1"}`)},
		{"InventoryReserved", rabbitmq.EvtInventoryResourceReserved, []byte(`{"orderId": "o1"}`)},
		{"InventoryFailed", rabbitmq.EvtInventoryResourceFailed, []byte(`{"orderId": "o1"}`)},
		{"PaymentAuthorized", rabbitmq.EvtPaymentTransactionAuthorized, []byte(`{"sagaId": "s1"}`)},
		{"PaymentDeclined", rabbitmq.EvtPaymentTransactionDeclined, []byte(`{"sagaId": "s1"}`)},
		{"OrderCreated", rabbitmq.EvtOrderManagementCreated, []byte(`{"orderId": "o1"}`)},
		{"Unknown", "unknown.key", []byte(`{}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyRoutingKey, tt.routingKey)
			err := h.HandleSagaEvent(ctx, tt.payload)
			if err != nil {
				t.Errorf("HandleSagaEvent() unexpected error: %v", err)
			}
		})
	}
}

func TestHandleSubmitOrder(t *testing.T) {
	tests := []struct {
		name        string
		payload     []byte
		mockErr     error
		expectError bool
	}{
		{"Success", []byte(`{"cartId": "c1"}`), nil, false},
		{"UnmarshalError", []byte(`invalid`), nil, false}, // Returns nil on unmarshal err
		{"UseCaseError", []byte(`{"cartId": "c1"}`), errors.New("err"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSagaUseCase{
				startSaga: func(ctx context.Context, cartID string) error {
					return tt.mockErr
				},
			}
			h := NewRabbitMQHandler(uc, &mockPublisher{})
			err := h.HandleSubmitOrder(context.Background(), tt.payload)
			if (err != nil) != tt.expectError {
				t.Errorf("HandleSubmitOrder() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandleInventoryReserved(t *testing.T) {
	tests := []struct {
		name        string
		payload     []byte
		mockErr     error
		expectError bool
	}{
		{"Success", []byte(`{"orderId": "o1"}`), nil, false},
		{"UnmarshalError", []byte(`invalid`), nil, false},
		{"UseCaseError", []byte(`{"orderId": "o1"}`), errors.New("err"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSagaUseCase{
				handleInventoryReserved: func(ctx context.Context, sagaID string) error {
					return tt.mockErr
				},
			}
			h := NewRabbitMQHandler(uc, &mockPublisher{})
			err := h.HandleInventoryReserved(context.Background(), tt.payload)
			if (err != nil) != tt.expectError {
				t.Errorf("HandleInventoryReserved() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandleInventoryFailed(t *testing.T) {
	tests := []struct {
		name        string
		payload     []byte
		mockErr     error
		expectError bool
	}{
		{"Success", []byte(`{"orderId": "o1"}`), nil, false},
		{"UnmarshalError", []byte(`invalid`), nil, false},
		{"UseCaseError", []byte(`{"orderId": "o1"}`), errors.New("err"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSagaUseCase{
				handleInventoryFailed: func(ctx context.Context, sagaID string) error {
					return tt.mockErr
				},
			}
			h := NewRabbitMQHandler(uc, &mockPublisher{})
			err := h.HandleInventoryFailed(context.Background(), tt.payload)
			if (err != nil) != tt.expectError {
				t.Errorf("HandleInventoryFailed() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandlePaymentAuthorized(t *testing.T) {
	tests := []struct {
		name        string
		payload     []byte
		mockErr     error
		expectError bool
	}{
		{"Success", []byte(`{"sagaId": "s1"}`), nil, false},
		{"UnmarshalError", []byte(`invalid`), nil, false},
		{"UseCaseError", []byte(`{"sagaId": "s1"}`), errors.New("err"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSagaUseCase{
				handlePaymentAuthorized: func(ctx context.Context, sagaID string) error {
					return tt.mockErr
				},
			}
			h := NewRabbitMQHandler(uc, &mockPublisher{})
			err := h.HandlePaymentAuthorized(context.Background(), tt.payload)
			if (err != nil) != tt.expectError {
				t.Errorf("HandlePaymentAuthorized() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandlePaymentDeclined(t *testing.T) {
	tests := []struct {
		name        string
		payload     []byte
		mockErr     error
		expectError bool
	}{
		{"Success", []byte(`{"sagaId": "s1"}`), nil, false},
		{"UnmarshalError", []byte(`invalid`), nil, false},
		{"UseCaseError", []byte(`{"sagaId": "s1"}`), errors.New("err"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSagaUseCase{
				handlePaymentDeclined: func(ctx context.Context, sagaID string) error {
					return tt.mockErr
				},
			}
			h := NewRabbitMQHandler(uc, &mockPublisher{})
			err := h.HandlePaymentDeclined(context.Background(), tt.payload)
			if (err != nil) != tt.expectError {
				t.Errorf("HandlePaymentDeclined() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHandleOrderCreated(t *testing.T) {
	tests := []struct {
		name        string
		payload     []byte
		mockErr     error
		expectError bool
	}{
		{"Success", []byte(`{"orderId": "o1"}`), nil, false},
		{"UnmarshalError", []byte(`invalid`), nil, false},
		{"UseCaseError", []byte(`{"orderId": "o1"}`), errors.New("err"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSagaUseCase{
				handleOrderCreated: func(ctx context.Context, sagaID string) error {
					return tt.mockErr
				},
			}
			h := NewRabbitMQHandler(uc, &mockPublisher{})
			err := h.HandleOrderCreated(context.Background(), tt.payload)
			if (err != nil) != tt.expectError {
				t.Errorf("HandleOrderCreated() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

type mockPublisher struct {
	publishToQueue func(ctx context.Context, queueName, correlationID string, body interface{}) error
}

func (m *mockPublisher) PublishToQueue(ctx context.Context, queueName, correlationID string, body interface{}) error {
	if m.publishToQueue != nil {
		return m.publishToQueue(ctx, queueName, correlationID, body)
	}
	return nil
}

func (m *mockPublisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error { return nil }
func (m *mockPublisher) PublishToTopic(ctx context.Context, routingKey string, body interface{}) error { return nil }
func (m *mockPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error { return nil }
func (m *mockPublisher) Close() error { return nil }

func TestGetSagaQueryReturnsSagaInstanceToCaller(t *testing.T) {
	mockSaga := &domain.SagaInstance{ID: "s1", Status: "PENDING"}

	uc := &mockSagaUseCase{
		getSaga: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
			if id == "s1" {
				return mockSaga, nil
			}
			return nil, errors.New("not found")
		},
	}

	pub := &mockPublisher{
		publishToQueue: func(ctx context.Context, queueName, correlationID string, body interface{}) error {
			if queueName != "reply_queue" {
				t.Errorf("Expected reply_queue, got %s", queueName)
			}
			return nil
		},
	}

	h := NewRabbitMQHandler(uc, pub)

	ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "reply_queue")
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "cor-123")

	err := h.HandleGetSaga(ctx, []byte(`{"id":"s1"}`))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetSagaQueryFailsWhenSagaNotFound(t *testing.T) {
	uc := &mockSagaUseCase{
		getSaga: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
			return nil, errors.New("not found")
		},
	}
	h := NewRabbitMQHandler(uc, &mockPublisher{})
	
	err := h.HandleGetSaga(context.Background(), []byte(`{"id":"s1"}`))
	if err == nil {
		t.Errorf("Expected error when saga not found")
	}
}

func TestGetSagaQueryFailsWhenJsonInvalid(t *testing.T) {
	h := NewRabbitMQHandler(&mockSagaUseCase{}, &mockPublisher{})
	err := h.HandleGetSaga(context.Background(), []byte(`invalid json`))
	if err == nil {
		t.Errorf("Expected error when JSON is invalid")
	}
}

func TestGetSagaQueryFailsWhenReplyToMissing(t *testing.T) {
	uc := &mockSagaUseCase{
		getSaga: func(ctx context.Context, id string) (*domain.SagaInstance, error) {
			return &domain.SagaInstance{ID: "s1"}, nil
		},
	}
	h := NewRabbitMQHandler(uc, &mockPublisher{})
	
	err := h.HandleGetSaga(context.Background(), []byte(`{"id":"s1"}`))
	if err != nil {
		t.Errorf("Expected no error when ReplyTo is missing (it should just log and return nil), got %v", err)
	}
}
