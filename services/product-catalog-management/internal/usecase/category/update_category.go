package category

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type UpdateCategoryUseCase struct {
	repo      ports.CategoryRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewUpdateCategoryUseCase(repo ports.CategoryRepository, publisher ports.EventPublisher, tm repository.TransactionManager) ports.UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{
		repo:      repo,
		publisher: publisher,
		tm:        tm,
	}
}

func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, input ports.UpdateCategoryInput) (*domain.Category, error) {
	category, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		category.Name = *input.Name
	}
	if input.Description != nil {
		category.Description = *input.Description
	}
	if input.ParentID != nil {
		category.ParentID = input.ParentID
	}
	if input.IsRoot != nil {
		category.IsRoot = *input.IsRoot
	}
	if input.CatalogID != nil {
		category.CatalogID = input.CatalogID
	}
	if input.ValidFor != nil {
		category.ValidFor = *input.ValidFor
	}
	if input.LifecycleStatus != nil {
		category.LifecycleStatus = *input.LifecycleStatus
	}

	category.LastUpdate = time.Now().UTC()

	if err := category.Validate(); err != nil {
		return nil, err
	}

	// Persist & Publish in Transaction
	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Update(ctx, category); err != nil {
			return err
		}
		if err := uc.publisher.PublishCategoryUpdated(ctx, domain.CategoryUpdatedEvent{Category: category}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return category, err
	}

	return category, nil
}
