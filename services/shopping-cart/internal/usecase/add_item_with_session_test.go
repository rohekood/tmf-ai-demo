package usecase_test

import (
	"context"
	"testing"

	"tmf/services/shopping-cart/internal/core/domain"
	"tmf/services/shopping-cart/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddItem_WithQualificationSession(t *testing.T) {
	ctx := context.Background()

	t.Run("Should use session price when qualificationSessionId provided", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCartRepository)
		mockQualClient := new(MockQualificationClient)
		mockSession := new(MockQualificationSession)
		uc := usecase.NewManageItemsUseCase(mockRepo, mockQualClient)

		cartID := "cart-123"
		offeringID := "offering-abc"
		sessionID := "session-xyz"
		qty := 2

		// Session returns discounted price (VIP: 20% off from 100 EUR = 80 EUR)
		mockQualClient.On("GetSession", ctx, sessionID).Return(mockSession, nil)
		mockSession.On("GetOfferingPrice", offeringID).Return(80.0, "EUR", true, true)

		// Expect Get to return nil (new cart)
		mockRepo.On("Get", ctx, cartID).Return(nil, nil)

		// Expect Save with session price
		mockRepo.On("Save", ctx, mock.MatchedBy(func(c *domain.Cart) bool {
			return c.ID == cartID &&
				len(c.Items) == 1 &&
				c.Items[0].UnitAmount == 80.0 && // Session price, not base price
				c.Items[0].Currency == "EUR" &&
				c.TotalPriceAmount == 160.0 // 80 * 2
		}), mock.Anything).Return(nil)

		// Act
		err := uc.AddItem(ctx, cartID, offeringID, sessionID, qty)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockQualClient.AssertExpectations(t)
		mockSession.AssertExpectations(t)
	})

	t.Run("Should fail when offering not found in session", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCartRepository)
		mockQualClient := new(MockQualificationClient)
		mockSession := new(MockQualificationSession)
		uc := usecase.NewManageItemsUseCase(mockRepo, mockQualClient)

		cartID := "cart-123"
		offeringID := "offering-not-in-session"
		sessionID := "session-xyz"
		qty := 1

		mockQualClient.On("GetSession", ctx, sessionID).Return(mockSession, nil)
		mockSession.On("GetOfferingPrice", offeringID).Return(0.0, "", false, false)

		mockRepo.On("Get", ctx, cartID).Return(nil, nil)

		// Act
		err := uc.AddItem(ctx, cartID, offeringID, sessionID, qty)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in qualification session")
		mockQualClient.AssertExpectations(t)
		mockSession.AssertExpectations(t)
	})

	t.Run("Should fail when offering not eligible in session", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCartRepository)
		mockQualClient := new(MockQualificationClient)
		mockSession := new(MockQualificationSession)
		uc := usecase.NewManageItemsUseCase(mockRepo, mockQualClient)

		cartID := "cart-123"
		offeringID := "offering-not-eligible"
		sessionID := "session-xyz"
		qty := 1

		mockQualClient.On("GetSession", ctx, sessionID).Return(mockSession, nil)
		mockSession.On("GetOfferingPrice", offeringID).Return(100.0, "EUR", false, true)

		mockRepo.On("Get", ctx, cartID).Return(nil, nil)

		// Act
		err := uc.AddItem(ctx, cartID, offeringID, sessionID, qty)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not eligible")
		mockQualClient.AssertExpectations(t)
		mockSession.AssertExpectations(t)
	})

	t.Run("Should fallback to internal pricing when no sessionId", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCartRepository)
		mockQualClient := new(MockQualificationClient)
		uc := usecase.NewManageItemsUseCase(mockRepo, mockQualClient)

		cartID := "cart-123"
		offeringID := "offering-abc"
		qty := 1

		mockRepo.On("Get", ctx, cartID).Return(nil, nil)
		mockRepo.On("GetPrice", ctx, offeringID).Return(&domain.ProductPrice{
			UnitAmount: 100.0,
			Currency:   "EUR",
		}, nil)

		mockRepo.On("Save", ctx, mock.MatchedBy(func(c *domain.Cart) bool {
			return c.Items[0].UnitAmount == 100.0 // Internal price
		}), mock.Anything).Return(nil)

		// Act
		err := uc.AddItem(ctx, cartID, offeringID, "", qty) // Empty sessionId

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		// QualClient should NOT be called
		mockQualClient.AssertNotCalled(t, "GetSession")
	})
}
