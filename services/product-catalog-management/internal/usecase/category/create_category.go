package category

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/google/uuid"
)

type CreateCategory struct {
	repo      ports.CategoryRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewCreateCategory(repo ports.CategoryRepository, publisher ports.EventPublisher, tm repository.TransactionManager) *CreateCategory {
	return &CreateCategory{
		repo:      repo,
		publisher: publisher,
		tm:        tm,
	}
}

func (uc *CreateCategory) Execute(ctx context.Context, input ports.CreateCategoryInput) (*domain.Category, error) {
	lifecycleStatus := input.LifecycleStatus
	if lifecycleStatus == "" {
		lifecycleStatus = "Draft"
	}

	category := &domain.Category{
		ID:              uuid.New().String(),
		Name:            input.Name,
		Description:     input.Description,
		ParentID:        input.ParentID,
		IsRoot:          input.IsRoot,
		CatalogID:       input.CatalogID,
		ValidFor:        input.ValidFor,
		LastUpdate:      time.Now().UTC(),
		LifecycleStatus: lifecycleStatus,
	}

	if err := category.Validate(); err != nil {
		return nil, err
	}

	// Persist & Publish in Transaction
	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Create(ctx, category); err != nil {
			return err
		}
		if err := uc.publisher.PublishCategoryCreated(ctx, domain.CategoryCreatedEvent{Category: category}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return category, nil
}
