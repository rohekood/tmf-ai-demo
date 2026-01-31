package usecase

import (
	"context"
	"time"

	"tmf/services/shopping-cart/internal/core/domain"
	"tmf/services/shopping-cart/internal/core/ports"
)

type syncCatalogUseCase struct {
	repo ports.CartRepository
}

func NewSyncCatalogUseCase(repo ports.CartRepository) ports.SyncCatalogUseCase {
	return &syncCatalogUseCase{repo: repo}
}

func (u *syncCatalogUseCase) SyncOffering(ctx context.Context, offeringID string, price float64, currency string) error {
	p := &domain.ProductPrice{
		ID:         offeringID,
		UnitAmount: price,
		Currency:   currency,
		UpdatedAt:  time.Now(),
	}
	return u.repo.UpsertPrice(ctx, p)
}
