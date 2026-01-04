package category

import (
	"context"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type DeleteCategoryUseCase struct {
	repo      ports.CategoryRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewDeleteCategoryUseCase(repo ports.CategoryRepository, publisher ports.EventPublisher, tm repository.TransactionManager) ports.DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{
		repo:      repo,
		publisher: publisher,
		tm:        tm,
	}
}

func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, input ports.DeleteCategoryInput) error {
	existing, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return err
	}

	// Perform Delete & Publish in Transaction
	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Delete(ctx, existing.ID); err != nil {
			return err
		}
		if err := uc.publisher.PublishCategoryDeleted(ctx, domain.CategoryDeletedEvent{ID: existing.ID}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
