package catalog

import (
	"context"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type DeleteCatalogUseCase struct {
	repo      ports.CatalogRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewDeleteCatalogUseCase(repo ports.CatalogRepository, publisher ports.EventPublisher, tm repository.TransactionManager) ports.DeleteCatalogUseCase {
	return &DeleteCatalogUseCase{
		repo:      repo,
		publisher: publisher,
		tm:        tm,
	}
}

func (uc *DeleteCatalogUseCase) Execute(ctx context.Context, input ports.DeleteCatalogInput) error {
	// 1. Check if exists
	existing, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return err
	}

	// 2. Perform Delete & Publish in Transaction
	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Delete(ctx, input.ID); err != nil {
			return err
		}
		if err := uc.publisher.PublishCatalogDeleted(ctx, domain.CatalogDeletedEvent{ID: existing.ID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
