package catalog

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type UpdateCatalogUseCase struct {
	repo      ports.CatalogRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewUpdateCatalogUseCase(repo ports.CatalogRepository, publisher ports.EventPublisher, tm repository.TransactionManager) ports.UpdateCatalogUseCase {
	return &UpdateCatalogUseCase{
		repo:      repo,
		publisher: publisher,
		tm:        tm, // Inject TM
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

	catalog.LastUpdate = time.Now().UTC()

	// 3. Validate
	if err := catalog.Validate(); err != nil {
		return nil, err
	}

	// 4. Persist & Publish in Transaction
	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Update(ctx, catalog); err != nil {
			return err
		}
		if err := uc.publisher.PublishCatalogUpdated(ctx, domain.CatalogUpdatedEvent{Catalog: catalog}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return catalog, err
	}

	return catalog, nil
}
