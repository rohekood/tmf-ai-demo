package offering

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type DeleteProductOfferingUseCase struct {
	repo      ports.ProductOfferingRepository
	publisher ports.EventPublisher
}

func NewDeleteProductOfferingUseCase(repo ports.ProductOfferingRepository, publisher ports.EventPublisher) ports.DeleteProductOfferingUseCase {
	return &DeleteProductOfferingUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *DeleteProductOfferingUseCase) Execute(ctx context.Context, input ports.DeleteProductOfferingInput) error {
	existing, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, existing.ID); err != nil {
		return err
	}

	if err := uc.publisher.PublishProductOfferingDeleted(ctx, domain.ProductOfferingDeletedEvent{ID: existing.ID}); err != nil {
		return err
	}

	return nil
}
