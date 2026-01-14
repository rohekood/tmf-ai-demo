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
	specRepo  ports.ProductSpecificationRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewCreateProductOffering(
	repo ports.ProductOfferingRepository,
	specRepo ports.ProductSpecificationRepository,
	publisher ports.EventPublisher,
	tm repository.TransactionManager,
) *CreateProductOffering {
	return &CreateProductOffering{
		repo:      repo,
		specRepo:  specRepo,
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
		offering.LifecycleStatus = domain.LifecycleStatusDraft
	}

	// Validate Offering (Price, Enums)
	if err := offering.Validate(); err != nil {
		return nil, err
	}

	// Cross-Entity Validation: Specification
	if offering.ProductSpecificationID != nil {
		spec, err := uc.specRepo.Get(ctx, *offering.ProductSpecificationID)
		if err != nil {
			return nil, err // Spec must exist
		}
		if spec.LifecycleStatus == domain.SpecLifecycleStatusRetired {
			return nil, domain.ErrInvalidInput // Cannot create offering for retired spec
		}
	} else if !offering.IsBundle {
		// Non-bundle offerings usually require a spec, but we'll leave optional if business allows
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
