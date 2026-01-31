package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"tmf/pkg/rabbitmq"
	"tmf/services/shopping-cart/internal/core/domain"
	"tmf/services/shopping-cart/internal/core/ports"

	"github.com/google/uuid"
)

type manageItemsUseCase struct {
	repo ports.CartRepository
}

func NewManageItemsUseCase(repo ports.CartRepository) ports.ManageItemsUseCase {
	return &manageItemsUseCase{repo: repo}
}

func (u *manageItemsUseCase) AddItem(ctx context.Context, cartID, offeringID string, qty int) error {
	// 1. Load Cart
	cart, err := u.repo.Get(ctx, cartID)
	if err != nil {
		return err
	}
	if cart == nil {
		cart = &domain.Cart{
			ID:          cartID,
			Status:      domain.CartStatusActive,
			Version:     1,
			ValidForEnd: time.Now().Add(24 * time.Hour),
			Items:       []domain.CartItem{},
		}
	}

	// 2. Fetch Price (Internal Lookup)
	price, err := u.repo.GetPrice(ctx, offeringID)
	var unitAmount float64
	var currency = "EUR"

	if err != nil {
		slog.WarnContext(ctx, "Failed to look up price", "error", err)
	} else if price != nil {
		unitAmount = price.UnitAmount
		currency = price.Currency
	} else {
		slog.WarnContext(ctx, "Price not found for offering", "offeringId", offeringID)
	}

	// 3. Add Item Logic
	newItem := domain.CartItem{
		ID:         uuid.New().String(),
		CartID:     cartID,
		OfferingID: offeringID,
		Quantity:   qty,
		UnitAmount: unitAmount,
		Currency:   currency,
	}
	cart.Items = append(cart.Items, newItem)

	// 4. Recalculate Totals (Inline Pricing)
	var total float64
	for _, item := range cart.Items {
		total += item.UnitAmount * float64(item.Quantity)
	}
	cart.TotalPriceAmount = total
	cart.TotalPriceCurrency = currency // Simplify: assume single currency

	cart.Version++
	// Status remains Active because pricing is done!

	// 5. Prepare Outbox Event (Cart Updated & Priced)
	eventPayload, _ := json.Marshal(cart)
	events := []domain.OutboxEvent{
		{
			ID:        uuid.New().String(),
			Topic:     rabbitmq.EvtCartSessionUpdated, // Subscribers (if any) get the full priced cart
			Payload:   eventPayload,
			Status:    "PENDING",
			CreatedAt: time.Now(),
		},
	}

	// 6. Atomic Save
	if err := u.repo.Save(ctx, cart, events); err != nil {
		return fmt.Errorf("failed to save cart: %w", err)
	}

	return nil
}
