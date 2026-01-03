package specification

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/google/uuid"
)

type CreateProductSpecification struct {
	repo      ports.ProductSpecificationRepository
	publisher ports.EventPublisher
}

func NewCreateProductSpecification(repo ports.ProductSpecificationRepository, publisher ports.EventPublisher) *CreateProductSpecification {
	return &CreateProductSpecification{
		repo:      repo,
		publisher: publisher,
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
		LastUpdate:      time.Now(),
	}

	if spec.LifecycleStatus == "" {
		spec.LifecycleStatus = "Created"
	}

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, spec); err != nil {
		return nil, err
	}

	_ = uc.publisher.PublishProductSpecificationCreated(ctx, domain.ProductSpecificationCreatedEvent{ProductSpecification: spec})

	return spec, nil
}
