package catalog

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type UpdateCatalogUseCase struct {
	repo      ports.CatalogRepository
	publisher ports.EventPublisher
}

func NewUpdateCatalogUseCase(repo ports.CatalogRepository, publisher ports.EventPublisher) ports.UpdateCatalogUseCase {
	return &UpdateCatalogUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *UpdateCatalogUseCase) Execute(ctx context.Context, input ports.UpdateCatalogInput) (*domain.Catalog, error) {
	// 1. Get existing catalog
	catalog, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// 2. Update fields if provided
	if input.Name != nil {
		catalog.Name = *input.Name
	}
	if input.Description != nil {
		catalog.Description = *input.Description
	}
	if input.ValidFor != nil {
		catalog.ValidFor = *input.ValidFor
	}
	if input.LifecycleStatus != nil {
		catalog.LifecycleStatus = *input.LifecycleStatus
	}

	catalog.LastUpdate = time.Now()

	// 3. Validate
	if err := catalog.Validate(); err != nil {
		return nil, err
	}

	// 4. Persist
	if err := uc.repo.Update(ctx, catalog); err != nil {
		return nil, err
	}

	// 5. Publish Event
	if err := uc.publisher.PublishCatalogUpdated(ctx, domain.CatalogUpdatedEvent{Catalog: catalog}); err != nil {
		return catalog, err
	}

	return catalog, nil
}
