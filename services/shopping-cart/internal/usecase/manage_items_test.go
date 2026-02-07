package usecase_test

import (
	"context"
	"testing"

	"tmf/services/shopping-cart/internal/core/domain"
	"tmf/services/shopping-cart/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCartRepository is a mock implementation of ports.CartRepository
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

// MockQualificationClient for testing
type MockQualificationClient struct {
	mock.Mock
}

// MockQualificationSession for testing
type MockQualificationSession struct {
	mock.Mock
}

func (m *MockQualificationSession) GetOfferingPrice(offeringID string) (price float64, currency string, eligible bool, found bool) {
	args := m.Called(offeringID)
	return args.Get(0).(float64), args.String(1), args.Bool(2), args.Bool(3)
}

func (m *MockQualificationClient) GetSession(ctx context.Context, sessionID string) (usecase.QualificationSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(usecase.QualificationSession), args.Error(1)
}

func TestManageItemsUseCase_AddItem(t *testing.T) {
	ctx := context.Background()

	t.Run("Scenario 1: New Cart", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCartRepository)
		mockQualClient := new(MockQualificationClient)
		uc := usecase.NewManageItemsUseCase(mockRepo, mockQualClient)
		cartID := "cart-123"
		offeringID := "offering-abc"
		qty := 2

		// Expect Get to return nil (cart not found)
		mockRepo.On("Get", ctx, cartID).Return(nil, nil)

		// Expect GetPrice
		mockRepo.On("GetPrice", ctx, offeringID).Return(&domain.ProductPrice{UnitAmount: 10.0, Currency: "EUR"}, nil)

		// Expect Save to be called with a new cart
		mockRepo.On("Save", ctx, mock.MatchedBy(func(c *domain.Cart) bool {
			return c.ID == cartID &&
				c.Version == 2 && // 1 (init) + 1 (increment)
				c.Status == domain.CartStatusActive &&
				len(c.Items) == 1 &&
				c.Items[0].OfferingID == offeringID &&
				c.Items[0].Quantity == qty
		}), mock.MatchedBy(func(events []domain.OutboxEvent) bool {
			return len(events) == 1 && events[0].Status == "PENDING"
		})).Return(nil)

		// Act
		err := uc.AddItem(ctx, cartID, offeringID, "", qty)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Scenario 2: Existing Cart", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCartRepository)
		mockQualClient := new(MockQualificationClient)
		uc := usecase.NewManageItemsUseCase(mockRepo, mockQualClient)
		cartID := "cart-456"
		offeringID := "offering-xyz"
		qty := 1
		existingItem := domain.CartItem{ID: "item-1", OfferingID: "offering-old", Quantity: 1}

		existingCart := &domain.Cart{
			ID:      cartID,
			Status:  domain.CartStatusActive,
			Version: 5,
			Items:   []domain.CartItem{existingItem},
		}

		// Expect Get to return existing cart
		mockRepo.On("Get", ctx, cartID).Return(existingCart, nil)

		// Expect GetPrice
		mockRepo.On("GetPrice", ctx, offeringID).Return(&domain.ProductPrice{UnitAmount: 10.0, Currency: "EUR"}, nil)

		// Expect Save to be called with updated cart
		mockRepo.On("Save", ctx, mock.MatchedBy(func(c *domain.Cart) bool {
			return c.ID == cartID &&
				c.Version == 6 && // 5 + 1
				c.Status == domain.CartStatusActive &&
				len(c.Items) == 2 &&
				c.Items[1].OfferingID == offeringID
		}), mock.Anything).Return(nil)

		// Act
		err := uc.AddItem(ctx, cartID, offeringID, "", qty)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
