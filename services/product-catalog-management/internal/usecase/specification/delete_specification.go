package specification

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type DeleteProductSpecificationUseCase struct {
	repo      ports.ProductSpecificationRepository
	publisher ports.EventPublisher
}

func NewDeleteProductSpecificationUseCase(repo ports.ProductSpecificationRepository, publisher ports.EventPublisher) ports.DeleteProductSpecificationUseCase {
	return &DeleteProductSpecificationUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *DeleteProductSpecificationUseCase) Execute(ctx context.Context, input ports.DeleteProductSpecificationInput) error {
	existing, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, existing.ID); err != nil {
		return err
	}

	if err := uc.publisher.PublishProductSpecificationDeleted(ctx, domain.ProductSpecificationDeletedEvent{ID: existing.ID}); err != nil {
		return err
	}

	return nil
}
