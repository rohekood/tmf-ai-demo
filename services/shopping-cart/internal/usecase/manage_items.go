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
	repo       ports.CartRepository
	qualClient QualificationClient
}

// QualificationSession interface for session data
type QualificationSession interface {
	GetOfferingPrice(offeringID string) (price float64, currency string, eligible bool, found bool)
}

// QualificationClient interface for RPC calls
type QualificationClient interface {
	GetSession(ctx context.Context, sessionID string) (QualificationSession, error)
}

func NewManageItemsUseCase(repo ports.CartRepository, qualClient QualificationClient) ports.ManageItemsUseCase {
	return &manageItemsUseCase{
		repo:       repo,
		qualClient: qualClient,
	}
}

func (u *manageItemsUseCase) AddItem(ctx context.Context, cartID, offeringID, qualificationSessionID string, qty int) error {
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

	// 2. Determine Price Source
	var unitAmount float64
	var currency = "EUR"

	if qualificationSessionID != "" {
		// Use qualification session price
		slog.InfoContext(ctx, "Using qualification session for pricing", "sessionId", qualificationSessionID)
		session, err := u.qualClient.GetSession(ctx, qualificationSessionID)
		if err != nil {
			return fmt.Errorf("failed to get qualification session: %w", err)
		}

		price, curr, eligible, found := session.GetOfferingPrice(offeringID)
		if !found {
			return fmt.Errorf("offering %s not found in qualification session", offeringID)
		}
		if !eligible {
			return fmt.Errorf("offering %s is not eligible in this session", offeringID)
		}
		unitAmount = price
		currency = curr
	} else {
		// Fallback: Fetch Price from internal lookup
		price, err := u.repo.GetPrice(ctx, offeringID)
		if err != nil {
			slog.WarnContext(ctx, "Failed to look up price", "error", err)
		} else if price != nil {
			unitAmount = price.UnitAmount
			currency = price.Currency
		} else {
			slog.WarnContext(ctx, "Price not found for offering", "offeringId", offeringID)
		}
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
