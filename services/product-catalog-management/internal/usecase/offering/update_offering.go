package offering

import (
	"context"
	"time"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/google/uuid"
)

type UpdateProductOfferingUseCase struct {
	repo      ports.ProductOfferingRepository
	publisher ports.EventPublisher
}

func NewUpdateProductOfferingUseCase(repo ports.ProductOfferingRepository, publisher ports.EventPublisher) ports.UpdateProductOfferingUseCase {
	return &UpdateProductOfferingUseCase{
		repo:      repo,
		publisher: publisher,
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
		offering.LifecycleStatus = *input.LifecycleStatus
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

	offering.LastUpdate = time.Now()

	if err := offering.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, offering); err != nil {
		return nil, err
	}

	if err := uc.publisher.PublishProductOfferingUpdated(ctx, domain.ProductOfferingUpdatedEvent{ProductOffering: offering}); err != nil {
		return offering, err
	}

	return offering, nil
}
