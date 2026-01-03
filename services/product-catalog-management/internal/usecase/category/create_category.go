package category

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/google/uuid"
)

type CreateCategory struct {
	repo      ports.CategoryRepository
	publisher ports.EventPublisher
}

func NewCreateCategory(repo ports.CategoryRepository, publisher ports.EventPublisher) *CreateCategory {
	return &CreateCategory{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *CreateCategory) Execute(ctx context.Context, input ports.CreateCategoryInput) (*domain.Category, error) {
	category := &domain.Category{
		ID:              uuid.New().String(),
		Name:            input.Name,
		Description:     input.Description,
		ParentID:        input.ParentID,
		IsRoot:          input.IsRoot,
		CatalogID:       input.CatalogID,
		ValidFor:        input.ValidFor,
		LastUpdate:      time.Now(),
		LifecycleStatus: "Active",
	}

	if err := category.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, category); err != nil {
		return nil, err
	}

	_ = uc.publisher.PublishCategoryCreated(ctx, domain.CategoryCreatedEvent{Category: category})

	return category, nil
}
