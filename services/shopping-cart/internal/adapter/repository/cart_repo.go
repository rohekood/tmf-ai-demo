package repository

import (
	"context"
	"tmf/services/shopping-cart/internal/core/domain"

	"gorm.io/gorm"
)

type CartRepository struct {
	db *gorm.DB
}

func NewLegacyCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

// Transactional Add Item + Outbox
func (r *CartRepository) AddItem(ctx context.Context, cartID string, item domain.CartItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get or Create Cart
		var cart domain.Cart
		if err := tx.FirstOrCreate(&cart, domain.Cart{ID: cartID, Status: "Active"}).Error; err != nil {
			return err
		}

		// 2. Add Item
		item.CartID = cartID
		if err := tx.Create(&item).Error; err != nil {
			return err
		}

		// 3. Create Event Payload
		// In real app, re-fetch full cart logic here
		event := domain.CartUpdatedEvent{
			CartID: cartID,
			Items:  []domain.CartItem{item}, // Simplified for demo
		}

		// 4. Save Outbox Event
		outbox, err := NewOutboxEvent("evt.cart.updated", event, extractHeaders(ctx))
		if err != nil {
			return err
		}
		if err := tx.Create(outbox).Error; err != nil {
			return err
		}

		return nil
	})
}

func extractHeaders(ctx context.Context) map[string]string {
	h := make(map[string]string)
	if v, ok := ctx.Value("X-Correlation-ID").(string); ok {
		h["X-Correlation-ID"] = v
	}
	if v, ok := ctx.Value("user").(string); ok {
		h["user"] = v
	}
	return h
}
