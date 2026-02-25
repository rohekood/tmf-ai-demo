package usecase_test

import (
	"context"
	"errors"
	"testing"

	"tmf/services/shopping-cart/internal/core/domain"
	"tmf/services/shopping-cart/internal/core/ports"
	"tmf/services/shopping-cart/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdatePriceUseCase_UpdatePrice(t *testing.T) {
	ctx := context.Background()

	t.Run("Should update price successfully", func(t *testing.T) {
		mockRepo := new(MockCartRepository)
		uc := usecase.NewUpdatePriceUseCase(mockRepo)

		cmd := ports.UpdateCartPriceCommand{
			CartID:     "cart-1",
			ForVersion: 1,
			Total:      100.0,
			Currency:   "USD",
			Items: []ports.ItemPriceDTO{
				{ItemID: "item-1", UnitAmount: 50.0, Currency: "USD"},
				{ItemID: "item-2", UnitAmount: 50.0, Currency: "USD"},
			},
		}

		cart := &domain.Cart{
			ID:      "cart-1",
			Version: 1,
			Items: []domain.CartItem{
				{ID: "item-1", UnitAmount: 0.0, Currency: ""},
			},
		}

		mockRepo.On("Get", ctx, "cart-1").Return(cart, nil)
		mockRepo.On("Save", ctx, cart, mock.AnythingOfType("[]domain.OutboxEvent")).Return(nil)

		err := uc.UpdatePrice(ctx, cmd)

		assert.NoError(t, err)
		assert.Equal(t, 50.0, cart.Items[0].UnitAmount)
		assert.Equal(t, "USD", cart.Items[0].Currency)
		assert.Equal(t, 100.0, cart.TotalPriceAmount)
		assert.Equal(t, "USD", cart.TotalPriceCurrency)
		assert.Equal(t, domain.CartStatusActive, cart.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Should discard stale update", func(t *testing.T) {
		mockRepo := new(MockCartRepository)
		uc := usecase.NewUpdatePriceUseCase(mockRepo)

		cmd := ports.UpdateCartPriceCommand{
			CartID:     "cart-1",
			ForVersion: 1,
		}

		cart := &domain.Cart{
			ID:      "cart-1",
			Version: 2, // Different version
		}

		mockRepo.On("Get", ctx, "cart-1").Return(cart, nil)

		err := uc.UpdatePrice(ctx, cmd)

		assert.NoError(t, err) // Should silently discard
		mockRepo.AssertExpectations(t)
	})

	t.Run("Should fail if cart not found", func(t *testing.T) {
		mockRepo := new(MockCartRepository)
		uc := usecase.NewUpdatePriceUseCase(mockRepo)

		cmd := ports.UpdateCartPriceCommand{CartID: "cart-1"}

		mockRepo.On("Get", ctx, "cart-1").Return(nil, nil)

		err := uc.UpdatePrice(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cart not found")
	})

	t.Run("Should propagate repo get error", func(t *testing.T) {
		mockRepo := new(MockCartRepository)
		uc := usecase.NewUpdatePriceUseCase(mockRepo)

		cmd := ports.UpdateCartPriceCommand{CartID: "cart-1"}

		mockRepo.On("Get", ctx, "cart-1").Return(nil, errors.New("db error"))

		err := uc.UpdatePrice(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("Should propagate repo save error", func(t *testing.T) {
		mockRepo := new(MockCartRepository)
		uc := usecase.NewUpdatePriceUseCase(mockRepo)

		cmd := ports.UpdateCartPriceCommand{
			CartID:     "cart-1",
			ForVersion: 1,
		}

		cart := &domain.Cart{ID: "cart-1", Version: 1}

		mockRepo.On("Get", ctx, "cart-1").Return(cart, nil)
		mockRepo.On("Save", ctx, cart, mock.AnythingOfType("[]domain.OutboxEvent")).Return(errors.New("db error"))

		err := uc.UpdatePrice(ctx, cmd)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save price update")
	})
}
