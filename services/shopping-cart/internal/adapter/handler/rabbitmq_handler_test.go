package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"tmf/services/shopping-cart/internal/adapter/handler"
	"tmf/services/shopping-cart/internal/core/domain"
	"tmf/services/shopping-cart/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockManageItemsUseCase mocks the ManageItemsUseCase
type MockManageItemsUseCase struct {
	mock.Mock
}

func (m *MockManageItemsUseCase) AddItem(ctx context.Context, cartID, offeringID, qualificationSessionID string, qty int) error {
	args := m.Called(ctx, cartID, offeringID, qualificationSessionID, qty)
	return args.Error(0)
}

// MockUpdatePriceUseCase mocks the UpdatePriceUseCase
type MockUpdatePriceUseCase struct {
	mock.Mock
}

func (m *MockUpdatePriceUseCase) UpdatePrice(ctx context.Context, cmd ports.UpdateCartPriceCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

// MockSyncCatalogUseCase mocks the SyncCatalogUseCase
type MockSyncCatalogUseCase struct {
	mock.Mock
}

func (m *MockSyncCatalogUseCase) SyncOffering(ctx context.Context, offeringID string, price float64, currency string) error {
	args := m.Called(ctx, offeringID, price, currency)
	return args.Error(0)
}

// MockCartRepository mocks the CartRepository
type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) Get(ctx context.Context, id string) (*domain.Cart, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Cart), args.Error(1)
}

func (m *MockCartRepository) Save(ctx context.Context, cart *domain.Cart, events []domain.OutboxEvent) error {
	args := m.Called(ctx, cart, events)
	return args.Error(0)
}

func (m *MockCartRepository) GetPrice(ctx context.Context, offeringID string) (*domain.ProductPrice, error) {
	args := m.Called(ctx, offeringID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProductPrice), args.Error(1)
}

func (m *MockCartRepository) UpsertPrice(ctx context.Context, price *domain.ProductPrice) error {
	args := m.Called(ctx, price)
	return args.Error(0)
}

// MockPublisher mocks the Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, exchange, routingKey string, payload interface{}) error {
	args := m.Called(ctx, exchange, routingKey, payload)
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

func TestCartHandler_HandleAddItem(t *testing.T) {
	ctx := context.Background()

	t.Run("Should handle add item with qualification session", func(t *testing.T) {
		// Arrange
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		payload := map[string]interface{}{
			"cartId":                 "cart-123",
			"offeringId":             "offering-abc",
			"quantity":               2,
			"qualificationSessionId": "session-xyz",
		}
		payloadBytes, _ := json.Marshal(payload)

		mockManageUC.On("AddItem", ctx, "cart-123", "offering-abc", "session-xyz", 2).
			Return(nil)

		// Act
		err := h.HandleAddItem(ctx, payloadBytes)

		// Assert
		assert.NoError(t, err)
		mockManageUC.AssertExpectations(t)
	})

	t.Run("Should handle add item without qualification session", func(t *testing.T) {
		// Arrange
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		payload := map[string]interface{}{
			"cartId":     "cart-123",
			"offeringId": "offering-abc",
			"quantity":   1,
		}
		payloadBytes, _ := json.Marshal(payload)

		mockManageUC.On("AddItem", ctx, "cart-123", "offering-abc", "", 1).
			Return(nil)

		// Act
		err := h.HandleAddItem(ctx, payloadBytes)

		// Assert
		assert.NoError(t, err)
		mockManageUC.AssertExpectations(t)
	})

	t.Run("Should fail when payload is invalid JSON", func(t *testing.T) {
		// Arrange
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		// Act
		err := h.HandleAddItem(ctx, []byte("invalid json"))

		// Assert
		assert.Error(t, err)
	})

	t.Run("Should propagate use case errors", func(t *testing.T) {
		// Arrange
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		payload := map[string]interface{}{
			"cartId":                 "cart-123",
			"offeringId":             "offering-abc",
			"quantity":               1,
			"qualificationSessionId": "session-expired",
		}
		payloadBytes, _ := json.Marshal(payload)

		mockManageUC.On("AddItem", ctx, "cart-123", "offering-abc", "session-expired", 1).
			Return(errors.New("session has expired"))

		// Act
		err := h.HandleAddItem(ctx, payloadBytes)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
		mockManageUC.AssertExpectations(t)
	})
}
