package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"tmf/pkg/rabbitmq"
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

func (m *MockManageItemsUseCase) RemoveItem(ctx context.Context, cartID, itemID string) error {
	args := m.Called(ctx, cartID, itemID)
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

		payload := map[string]any{
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

		payload := map[string]any{
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

		payload := map[string]any{
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

func TestCartHandler_HandleUpdatePrice(t *testing.T) {
	ctx := context.Background()

	t.Run("Should update price successfully", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		payload := map[string]any{
			"cartId": "cart-1",
			"items": []map[string]any{
				{
					"itemId": "item-1",
					"price":  99.99,
				},
			},
		}
		payloadBytes, _ := json.Marshal(payload)

		mockPriceUC.On("UpdatePrice", ctx, mock.AnythingOfType("ports.UpdateCartPriceCommand")).Return(nil)

		err := h.HandleUpdatePrice(ctx, payloadBytes)

		assert.NoError(t, err)
		mockPriceUC.AssertExpectations(t)
	})

	t.Run("Should fail on invalid JSON", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		err := h.HandleUpdatePrice(ctx, []byte("invalid"))

		assert.Error(t, err)
	})
}

func TestCartHandler_HandleCatalogEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("Should handle catalog event successfully", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		payload := map[string]any{
			"id":    "offering-1",
			"price": map[string]any{"amount": 10.0, "currency": "USD"},
		}
		payloadBytes, _ := json.Marshal(payload)

		mockSyncUC.On("SyncOffering", ctx, "offering-1", 10.0, "USD").Return(nil)

		err := h.HandleCatalogEvent(ctx, payloadBytes)

		assert.NoError(t, err)
		mockSyncUC.AssertExpectations(t)
	})

	t.Run("Should fail on invalid JSON", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		err := h.HandleCatalogEvent(ctx, []byte("invalid"))

		assert.Error(t, err)
	})
}

func TestCartHandler_HandleGetCart(t *testing.T) {
	ctx := context.Background()

	t.Run("Should get cart successfully", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		payload := map[string]any{
			"cartId": "cart-1",
		}
		payloadBytes, _ := json.Marshal(payload)

		expectedCart := &domain.Cart{ID: "cart-1"}
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyReplyTo, "reply-queue")
		ctx = context.WithValue(ctx, rabbitmq.ContextKeyAMQPCorrelationID, "corr-1")
		mockRepo.On("Get", ctx, "cart-1").Return(expectedCart, nil)
		mockPub.On("PublishToQueue", ctx, "reply-queue", "corr-1", expectedCart).Return(nil)

		err := h.HandleGetCart(ctx, payloadBytes)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("Should fail on invalid JSON", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		err := h.HandleGetCart(ctx, []byte("invalid"))

		assert.Error(t, err)
	})

	t.Run("Should fail if cart not found", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		payload := map[string]any{
			"cartId": "cart-1",
		}
		payloadBytes, _ := json.Marshal(payload)

		mockRepo.On("Get", ctx, "cart-1").Return(nil, errors.New("not found"))

		err := h.HandleGetCart(ctx, payloadBytes)

		assert.Error(t, err)
	})
}

func TestCartHandler_HandleRemoveItem(t *testing.T) {
	ctx := context.Background()

	t.Run("Should remove item successfully", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		payload := map[string]any{
			"cartId": "cart-123",
			"itemId": "item-abc",
		}
		payloadBytes, _ := json.Marshal(payload)

		mockManageUC.On("RemoveItem", ctx, "cart-123", "item-abc").Return(nil)

		err := h.HandleRemoveItem(ctx, payloadBytes)

		assert.NoError(t, err)
		mockManageUC.AssertExpectations(t)
	})

	t.Run("Should fail when payload is invalid JSON", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		err := h.HandleRemoveItem(ctx, []byte("invalid json"))

		assert.Error(t, err)
	})

	t.Run("Should propagate use case errors", func(t *testing.T) {
		mockManageUC := new(MockManageItemsUseCase)
		mockPriceUC := new(MockUpdatePriceUseCase)
		mockSyncUC := new(MockSyncCatalogUseCase)
		mockRepo := new(MockCartRepository)
		mockPub := new(MockPublisher)

		h := handler.NewCartHandler(mockManageUC, mockPriceUC, mockSyncUC, mockRepo, mockPub)

		payload := map[string]any{
			"cartId": "cart-123",
			"itemId": "item-not-found",
		}
		payloadBytes, _ := json.Marshal(payload)

		mockManageUC.On("RemoveItem", ctx, "cart-123", "item-not-found").
			Return(errors.New("item item-not-found not found in cart cart-123"))

		err := h.HandleRemoveItem(ctx, payloadBytes)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockManageUC.AssertExpectations(t)
	})
}
