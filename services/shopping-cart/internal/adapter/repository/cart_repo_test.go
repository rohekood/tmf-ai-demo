package repository_test

import (
	"context"
	"testing"

	"tmf/services/shopping-cart/internal/adapter/repository"
	"tmf/services/shopping-cart/internal/core/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLegacyCartRepository_AddItem(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewLegacyCartRepository(db)

	cartID := uuid.New().String()
	offeringID := uuid.New().String()
	customerID := uuid.New().String()

	db.Create(&domain.Cart{
		ID:         cartID,
		CustomerID: customerID,
		Status:     "Active",
	})

	item := domain.CartItem{
		ID:         uuid.New().String(),
		CartID:     cartID,
		OfferingID: offeringID,
		Quantity:   1,
		UnitAmount: 10.0,
		Currency:   "USD",
	}

	type ctxKey string
	var corrKey ctxKey = "X-Correlation-ID"
	var userKey ctxKey = "user"
	ctx := context.WithValue(context.Background(), corrKey, "corr-123")
	ctx = context.WithValue(ctx, userKey, "user-123")

	err := repo.AddItem(ctx, cartID, item)
	require.NoError(t, err)

	var cart domain.Cart
	err = db.First(&cart, "id = ?", cartID).Error
	assert.NoError(t, err)
	assert.Equal(t, cartID, cart.ID)

	var savedItem domain.CartItem
	err = db.First(&savedItem, "id = ?", item.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, offeringID, savedItem.OfferingID)
}
