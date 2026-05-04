package ports

import (
	"context"
	"tmf/services/shopping-cart/internal/core/domain"
)

// Primary Port (Use Cases)
type ManageItemsUseCase interface {
	AddItem(ctx context.Context, cartID, offeringID, qualificationSessionID string, qty int) error
	RemoveItem(ctx context.Context, cartID, itemID string) error
}

type SyncCatalogUseCase interface {
	SyncOffering(ctx context.Context, offeringID string, price float64, currency string) error
}

type UpdatePriceUseCase interface {
	UpdatePrice(ctx context.Context, cmd UpdateCartPriceCommand) error
}

type UpdateCartPriceCommand struct {
	CartID     string
	ForVersion int
	Items      []ItemPriceDTO
	Total      float64
	Currency   string
}

type ItemPriceDTO struct {
	ItemID     string
	UnitAmount float64
	Currency   string
}

// Secondary Port (Repository)
type CartRepository interface {
	Get(ctx context.Context, id string) (*domain.Cart, error)
	Save(ctx context.Context, cart *domain.Cart, events []domain.OutboxEvent) error

	// Pricing Support
	GetPrice(ctx context.Context, offeringID string) (*domain.ProductPrice, error)
	UpsertPrice(ctx context.Context, price *domain.ProductPrice) error
}
