package domain_test

import (
	"encoding/json"
	"testing"

	"tmf/services/shopping-cart/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestCartItem_MarshalJSON(t *testing.T) {
	item := domain.CartItem{
		ID:         "item-1",
		CartID:     "cart-1",
		OfferingID: "off-1",
		Quantity:   2,
		UnitAmount: 100.50,
		Currency:   "EUR",
	}

	b, err := json.Marshal(item)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(b, &result)
	assert.NoError(t, err)

	assert.Equal(t, "item-1", result["id"])
	assert.Equal(t, float64(2), result["quantity"])

	priceMap, ok := result["price"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 100.50, priceMap["amount"])
	assert.Equal(t, "EUR", priceMap["currency"])
}
