package offering

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/google/uuid"
)

type CreateProductOffering struct {
	repo      ports.ProductOfferingRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewCreateProductOffering(repo ports.ProductOfferingRepository, publisher ports.EventPublisher, tm repository.TransactionManager) *CreateProductOffering {
	return &CreateProductOffering{
		repo:      repo,
		publisher: publisher,
		tm:        tm,
	}
}

func (uc *CreateProductOffering) Execute(ctx context.Context, input ports.CreateProductOfferingInput) (*domain.ProductOffering, error) {
	offering := &domain.ProductOffering{
		ID:                     uuid.New().String(),
		Name:                   input.Name,
		Description:            input.Description,
		IsBundle:               input.IsBundle,
		IsSellable:             input.IsSellable,
		LifecycleStatus:        input.LifecycleStatus,
		ValidFor:               input.ValidFor,
		LastUpdate:             time.Now(),
		ProductSpecificationID: input.ProductSpecID,
		CategoryIDs:            input.CategoryIDs,
		ProductOfferingPrice:   input.Prices,
		Attachments:            input.Attachments,
	}

	for i := range offering.Attachments {
		if offering.Attachments[i].ID == "" {
			offering.Attachments[i].ID = uuid.New().String()
		}
	}

	if offering.LifecycleStatus == "" {
		offering.LifecycleStatus = "Created"
	}

	// Basic validation could be extended here or in domain
	if offering.Name == "" {
		return nil, domain.ErrInvalidInput
	}

	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Create(ctx, offering); err != nil {
			return err
		}

		if err := uc.publisher.PublishProductOfferingCreated(ctx, domain.ProductOfferingCreatedEvent{ProductOffering: offering}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return offering, nil
}
