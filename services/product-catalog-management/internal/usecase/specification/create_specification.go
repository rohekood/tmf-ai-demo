package specification

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/google/uuid"
)

type CreateProductSpecification struct {
	repo      ports.ProductSpecificationRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewCreateProductSpecification(repo ports.ProductSpecificationRepository, publisher ports.EventPublisher, tm repository.TransactionManager) *CreateProductSpecification {
	return &CreateProductSpecification{
		repo:      repo,
		publisher: publisher,
		tm:        tm,
	}
}

func (uc *CreateProductSpecification) Execute(ctx context.Context, input ports.CreateProductSpecificationInput) (*domain.ProductSpecification, error) {
	spec := &domain.ProductSpecification{
		ID:              uuid.New().String(),
		Name:            input.Name,
		ProductNumber:   input.ProductNumber,
		Description:     input.Description,
		IsBundle:        input.IsBundle,
		LifecycleStatus: input.LifecycleStatus,
		ValidFor:        input.ValidFor,
		Characteristics: input.Characteristics,
		LastUpdate:      time.Now().UTC(),
	}

	if spec.LifecycleStatus == "" {
		spec.LifecycleStatus = "Created"
	}

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	// Persist & Publish in Transaction
	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Create(ctx, spec); err != nil {
			return err
		}
		if err := uc.publisher.PublishProductSpecificationCreated(ctx, domain.ProductSpecificationCreatedEvent{ProductSpecification: spec}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return spec, nil
}
