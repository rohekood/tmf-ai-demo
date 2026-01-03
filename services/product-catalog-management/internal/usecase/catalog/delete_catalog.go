package catalog

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type DeleteCatalogUseCase struct {
	repo      ports.CatalogRepository
	publisher ports.EventPublisher
}

func NewDeleteCatalogUseCase(repo ports.CatalogRepository, publisher ports.EventPublisher) ports.DeleteCatalogUseCase {
	return &DeleteCatalogUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *DeleteCatalogUseCase) Execute(ctx context.Context, input ports.DeleteCatalogInput) error {
	// 1. Check if exists
	existing, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return err
	}

	// 2. Perform Delete
	if err := uc.repo.Delete(ctx, input.ID); err != nil {
		return err
	}

	// 3. Publish Event
	if err := uc.publisher.PublishCatalogDeleted(ctx, domain.CatalogDeletedEvent{ID: existing.ID}); err != nil {
		return err
	}

	return nil
}
