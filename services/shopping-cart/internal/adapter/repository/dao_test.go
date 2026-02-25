package repository

import (
	"testing"
	"tmf/services/shopping-cart/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestToDomainCart_NilCustomerID(t *testing.T) {
	dao := &CartTable{
		ID: "cart-1",
		CustomerID: nil,
	}

	cart := toDomainCart(dao)
	assert.Equal(t, "cart-1", cart.ID)
	assert.Equal(t, "", cart.CustomerID)
}

func TestToDAOCart_NilItems(t *testing.T) {
	cart := &domain.Cart{
		ID: "cart-1",
	}

	dao := toDAOCart(cart)
	assert.Equal(t, "cart-1", dao.ID)
}
