package specification

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type UpdateProductSpecificationUseCase struct {
	repo      ports.ProductSpecificationRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewUpdateProductSpecificationUseCase(repo ports.ProductSpecificationRepository, publisher ports.EventPublisher, tm repository.TransactionManager) ports.UpdateProductSpecificationUseCase {
	return &UpdateProductSpecificationUseCase{
		repo:      repo,
		publisher: publisher,
		tm:        tm,
	}
}

func (uc *UpdateProductSpecificationUseCase) Execute(ctx context.Context, input ports.UpdateProductSpecificationInput) (*domain.ProductSpecification, error) {
	spec, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		spec.Name = *input.Name
	}
	if input.Description != nil {
		spec.Description = *input.Description
	}
	if input.ProductNumber != nil {
		spec.ProductNumber = *input.ProductNumber
	}
	if input.IsBundle != nil {
		spec.IsBundle = *input.IsBundle
	}
	if input.LifecycleStatus != nil {
		spec.LifecycleStatus = *input.LifecycleStatus
	}
	if input.ValidFor != nil {
		spec.ValidFor = *input.ValidFor
	}
	if input.Characteristics != nil {
		spec.Characteristics = input.Characteristics
	}

	spec.LastUpdate = time.Now()

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	// Persist & Publish in Transaction
	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Update(ctx, spec); err != nil {
			return err
		}
		if err := uc.publisher.PublishProductSpecificationUpdated(ctx, domain.ProductSpecificationUpdatedEvent{ProductSpecification: spec}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return spec, err
	}

	return spec, nil
}
