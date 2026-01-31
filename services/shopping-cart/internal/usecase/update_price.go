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

type updatePriceUseCase struct {
	repo ports.CartRepository
}

func NewUpdatePriceUseCase(repo ports.CartRepository) ports.UpdatePriceUseCase {
	return &updatePriceUseCase{repo: repo}
}

func (u *updatePriceUseCase) UpdatePrice(ctx context.Context, cmd ports.UpdateCartPriceCommand) error {
	// 1. Load Cart
	cart, err := u.repo.Get(ctx, cmd.CartID)
	if err != nil {
		return err
	}
	if cart == nil {
		return fmt.Errorf("cart not found: %s", cmd.CartID)
	}

	// 2. Optimistic Locking Check
	if cmd.ForVersion != cart.Version {
		slog.WarnContext(ctx, "Ignoring stale price update",
			"cartVersion", cart.Version, "cmdVersion", cmd.ForVersion)
		return nil // Discard stale update
	}

	// 3. Apply Prices
	priceMap := make(map[string]ports.ItemPriceDTO)
	for _, p := range cmd.Items {
		priceMap[p.ItemID] = p
	}

	for i := range cart.Items {
		if p, ok := priceMap[cart.Items[i].ID]; ok {
			// Update flat fields
			cart.Items[i].UnitAmount = p.UnitAmount
			cart.Items[i].Currency = p.Currency
		}
	}

	cart.TotalPriceAmount = cmd.Total
	cart.TotalPriceCurrency = cmd.Currency
	cart.Status = domain.CartStatusActive // Back to Active

	// 4. Prepare Event (Priced)
	eventPayload, _ := json.Marshal(cart)
	events := []domain.OutboxEvent{
		{
			ID:        uuid.New().String(),
			Topic:     rabbitmq.EvtCartSessionPriced,
			Payload:   eventPayload,
			Status:    "PENDING",
			CreatedAt: time.Now(),
		},
	}

	// 5. Save
	if err := u.repo.Save(ctx, cart, events); err != nil {
		return fmt.Errorf("failed to save price update: %w", err)
	}

	return nil
}
