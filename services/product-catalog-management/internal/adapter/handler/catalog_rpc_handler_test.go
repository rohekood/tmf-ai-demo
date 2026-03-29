package handler

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"tmf/pkg/rabbitmq"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOfferingRepo struct {
	mock.Mock
}

func (m *MockOfferingRepo) Create(ctx context.Context, offering *domain.ProductOffering) error {
	args := m.Called(ctx, offering)
	return args.Error(0)
}

func (m *MockOfferingRepo) Get(ctx context.Context, id string) (*domain.ProductOffering, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.ProductOffering), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOfferingRepo) List(ctx context.Context, filters map[string]interface{}) ([]*domain.ProductOffering, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.ProductOffering), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockOfferingRepo) Update(ctx context.Context, offering *domain.ProductOffering) error {
	args := m.Called(ctx, offering)
	return args.Error(0)
}

func (m *MockOfferingRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	args := m.Called(ctx, exchange, routingKey, body)
	return args.Error(0)
}

func (m *MockPublisher) PublishToQueue(ctx context.Context, queueName string, correlationID string, body interface{}) error {
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

func (m *MockConsumer) Subscribe(routingKey string, handler rabbitmq.ConsumerHandler) error {
	args := m.Called(routingKey, handler)
	return args.Error(0)
}
func (m *MockConsumer) SubscribeToQueue(queueName string, handler rabbitmq.ConsumerHandler) error {
	args := m.Called(queueName, handler)
	return args.Error(0)
}
func (m *MockConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestCatalogRPCHandler_HandleGetOffering(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "reply_queue")
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr-123")

	payload := []byte(`{"offeringId":"off1"}`)
	offering := &domain.ProductOffering{
		ID:   "off1",
		Name: "Test Offering",
		ProductOfferingPrice: []domain.ProductOfferingPrice{
			{Price: domain.Money{Value: 150.0, Unit: "USD"}},
		},
	}

	repo.On("Get", ctx, "off1").Return(offering, nil)
	pub.On("PublishToQueue", ctx, "reply_queue", "corr-123", map[string]interface{}{
		"id":        "off1",
		"name":      "Test Offering",
		"basePrice": 150.0,
		"currency":  "USD",
	}).Return(nil)

	err := handler.HandleGetOffering(ctx, payload)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	pub.AssertExpectations(t)
}

func TestCatalogRPCHandler_HandleGetOffering_NoPrices(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "reply_queue")
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr-123")

	payload := []byte(`{"offeringId":"off1"}`)
	offering := &domain.ProductOffering{
		ID:   "off1",
		Name: "Test Offering",
	}

	repo.On("Get", ctx, "off1").Return(offering, nil)
	pub.On("PublishToQueue", ctx, "reply_queue", "corr-123", map[string]interface{}{
		"id":        "off1",
		"name":      "Test Offering",
		"basePrice": 100.0,
		"currency":  "EUR",
	}).Return(nil)

	err := handler.HandleGetOffering(ctx, payload)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	pub.AssertExpectations(t)
}

func TestCatalogRPCHandler_HandleGetOffersByCategory(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "reply_queue")
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr-123")

	payload := []byte(`{"category":"cat1"}`)
	offerings := []*domain.ProductOffering{
		{ID: "off1", Name: "Offer 1"},
	}

	repo.On("List", ctx, map[string]interface{}{"category": "cat1"}).Return(offerings, nil)

	pub.On("PublishToQueue", ctx, "reply_queue", "corr-123", mock.Anything).Return(nil)

	err := handler.HandleGetOffersByCategory(ctx, payload)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	pub.AssertExpectations(t)
}

func TestCatalogRPCHandler_HandleDispatch(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyRoutingKey, "query.catalog.offering.get")
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyReplyTo, "reply_queue")
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr-123")

	payload := []byte(`{"offeringId":"off1"}`)
	repo.On("Get", ctx, "off1").Return(&domain.ProductOffering{ID: "off1", Name: "Test"}, nil)
	pub.On("PublishToQueue", ctx, "reply_queue", "corr-123", mock.Anything).Return(nil)

	err := handler.HandleDispatch(ctx, payload)
	assert.NoError(t, err)

	ctx2 := context.WithValue(context.Background(), rabbitmq.ContextKeyRoutingKey, "unknown.key")
	err = handler.HandleDispatch(ctx2, []byte(`{}`))
	assert.NoError(t, err)
}

func TestCatalogRPCHandler_BindRPCHandlers(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	mockCons := new(MockConsumer)
	mockCons.On("Subscribe", "query.catalog.offering.#", mock.AnythingOfType("rabbitmq.ConsumerHandler")).Return(nil)

	err := handler.BindRPCHandlers(mockCons)
	assert.NoError(t, err)
	mockCons.AssertExpectations(t)
}

func TestCatalogRPCHandler_HandleGetOffering_Error(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "reply_queue")
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr-123")

	payload := []byte(`{"offeringId":"off1"}`)

	repo.On("Get", ctx, "off1").Return((*domain.ProductOffering)(nil), errors.New("not found"))

	pub.On("PublishToQueue", ctx, "reply_queue", "corr-123", map[string]string{"error": "not found"}).Return(nil)

	err := handler.HandleGetOffering(ctx, payload)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	pub.AssertExpectations(t)
}

func TestCatalogRPCHandler_InvalidPayload(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.Background()

	err := handler.HandleGetOffering(ctx, []byte("{invalid"))
	assert.Error(t, err)

	err = handler.HandleGetOffersByCategory(ctx, []byte("{invalid"))
	assert.Error(t, err)
}

func TestCatalogRPCHandler_MissingReplyTo(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.Background()
	payload := []byte(`{"offeringId":"off1"}`)

	repo.On("Get", ctx, "off1").Return(&domain.ProductOffering{ID: "1"}, nil)

	err := handler.HandleGetOffering(ctx, payload)
	assert.Error(t, err)

	payload2 := []byte(`{"category":"cat1"}`)
	repo.On("List", ctx, mock.Anything).Return([]*domain.ProductOffering{{ID: "1"}}, nil)

	err = handler.HandleGetOffersByCategory(ctx, payload2)
	assert.Error(t, err)
}

func TestCatalogRPCHandler_HandleGetOffersByCategory_Error(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.WithValue(context.Background(), rabbitmq.ContextKeyReplyTo, "reply_queue")
	ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr-123")

	payload := []byte(`{"category":"cat1"}`)

	repo.On("List", ctx, mock.Anything).Return([]*domain.ProductOffering(nil), errors.New("list error"))

	err := handler.HandleGetOffersByCategory(ctx, payload)
	assert.Error(t, err)
}

func TestCatalogRPCHandler_ReplyError_MissingReplyTo(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.Background()
	payload := []byte(`{"offeringId":"off1"}`)

	repo.On("Get", ctx, "off1").Return((*domain.ProductOffering)(nil), errors.New("get err"))

	err := handler.HandleGetOffering(ctx, payload)
	assert.Error(t, err)
}

func TestCatalogRPCHandler_BindRPCHandlers_Error(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	mockCons := new(MockConsumer)
	mockCons.On("Subscribe", "query.catalog.offering.#", mock.Anything).Return(errors.New("sub err"))

	err := handler.BindRPCHandlers(mockCons)
	assert.Error(t, err)
}

func TestCatalogRPCHandler_HandleDispatch_NoRoutingKey(t *testing.T) {
	repo := new(MockOfferingRepo)
	pub := new(MockPublisher)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewCatalogRPCHandler(repo, pub, logger)

	ctx := context.Background()

	err := handler.HandleDispatch(ctx, []byte(`{}`))
	assert.Error(t, err)
}
