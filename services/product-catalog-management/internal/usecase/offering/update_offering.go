package offering

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/google/uuid"
)

type UpdateProductOfferingUseCase struct {
	repo      ports.ProductOfferingRepository
	specRepo  ports.ProductSpecificationRepository
	publisher ports.EventPublisher
	tm        repository.TransactionManager
}

func NewUpdateProductOfferingUseCase(
	repo ports.ProductOfferingRepository,
	specRepo ports.ProductSpecificationRepository,
	publisher ports.EventPublisher,
	tm repository.TransactionManager,
) ports.UpdateProductOfferingUseCase {
	return &UpdateProductOfferingUseCase{
		repo:      repo,
		specRepo:  specRepo,
		publisher: publisher,
		tm:        tm,
	}
}

func (uc *UpdateProductOfferingUseCase) Execute(ctx context.Context, input ports.UpdateProductOfferingInput) (*domain.ProductOffering, error) {
	offering, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		offering.Name = *input.Name
	}
	if input.Description != nil {
		offering.Description = *input.Description
	}

	if input.LifecycleStatus != nil {
		newStatus := *input.LifecycleStatus
		// Validate Transitions
		if err := validateLifecycleTransition(offering.LifecycleStatus, newStatus); err != nil {
			return nil, err
		}

		// Cross-Entity Validation on Activation
		if newStatus == domain.LifecycleStatusActive {
			var specID string
			if offering.ProductSpecificationID != nil {
				specID = *offering.ProductSpecificationID
			}

			if specID != "" {
				spec, err := uc.specRepo.Get(ctx, specID)
				if err != nil {
					return nil, err
				}
				if spec.LifecycleStatus == domain.SpecLifecycleStatusRetired {
					return nil, domain.ErrInvalidInput // Cannot activate if spec is retired
				}
			}
		}
		offering.LifecycleStatus = newStatus
	}

	if input.ValidFor != nil {
		offering.ValidFor = *input.ValidFor
	}
	if input.IsBundle != nil {
		offering.IsBundle = *input.IsBundle
	}
	if input.IsSellable != nil {
		offering.IsSellable = *input.IsSellable
	}
	if input.CategoryIDs != nil {
		offering.CategoryIDs = input.CategoryIDs
	}
	if input.Prices != nil {
		offering.ProductOfferingPrice = input.Prices
	}
	if input.Attachments != nil {
		offering.Attachments = input.Attachments
		for i := range offering.Attachments {
			if offering.Attachments[i].ID == "" {
				offering.Attachments[i].ID = uuid.New().String()
			}
		}
	}

	offering.LastUpdate = time.Now().UTC()

	if err := offering.Validate(); err != nil {
		return nil, err
	}

	if err := uc.tm.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Update(ctx, offering); err != nil {
			return err
		}

		if err := uc.publisher.PublishProductOfferingUpdated(ctx, domain.ProductOfferingUpdatedEvent{ProductOffering: offering}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return offering, nil
}

func validateLifecycleTransition(oldStatus, newStatus string) error {
	if oldStatus == newStatus {
		return nil
	}
	// Simplified State Machine
	switch oldStatus {
	case domain.LifecycleStatusDraft:
		if newStatus == domain.LifecycleStatusActive || newStatus == domain.LifecycleStatusRetired {
			return nil
		}
	case domain.LifecycleStatusActive:
		if newStatus == domain.LifecycleStatusSuspended || newStatus == domain.LifecycleStatusRetired {
			return nil
		}
	case domain.LifecycleStatusSuspended:
		if newStatus == domain.LifecycleStatusActive || newStatus == domain.LifecycleStatusRetired {
			return nil
		}
	case domain.LifecycleStatusRetired:
		// No exit from retired usually, maybe allowed for correction? Sticking to strict for now.
		return domain.ErrInvalidInput
	}
	// Allow transition if old status was unknown or empty (migration case)
	if oldStatus == "" {
		return nil
	}

	// Default: Allow for now if not strictly forbidden above?
	// Or stricter: return error. Let's return error to be safe.
	return domain.ErrInvalidInput
}
