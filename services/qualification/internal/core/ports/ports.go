package ports

import (
	"context"
	"tmf/services/qualification/internal/core/domain"
)

// Primary Port (Use Case)
type CheckEligibilityUseCase interface {
	Execute(ctx context.Context, cmd domain.CheckEligibilityCommand) error
}

// Secondary Ports (Driven Adapters)

type GISClient interface {
	CheckPolygon(ctx context.Context, addr domain.Address) (bool, error)
}

type InventoryClient interface {
	GetPortCapacity(ctx context.Context, addr domain.Address) (int, error)
}

type CatalogClient interface {
	GetOffersByCategory(ctx context.Context, category string) ([]domain.EligibleCategory, error)
}

type EventPublisher interface {
	PublishEligibilityChecked(ctx context.Context, result domain.EligibilityResult) error
}
