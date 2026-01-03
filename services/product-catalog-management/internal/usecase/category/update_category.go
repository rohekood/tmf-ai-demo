package category

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type UpdateCategoryUseCase struct {
	repo      ports.CategoryRepository
	publisher ports.EventPublisher
}

func NewUpdateCategoryUseCase(repo ports.CategoryRepository, publisher ports.EventPublisher) ports.UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{
		repo:      repo,
		publisher: publisher,
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

	category.LastUpdate = time.Now()

	if err := category.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, category); err != nil {
		return nil, err
	}

	if err := uc.publisher.PublishCategoryUpdated(ctx, domain.CategoryUpdatedEvent{Category: category}); err != nil {
		return category, err
	}

	return category, nil
}
