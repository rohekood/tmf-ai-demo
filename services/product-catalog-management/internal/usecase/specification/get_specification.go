package specification

import (
	"context"

	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type GetProductSpecificationUseCase struct {
	repo ports.ProductSpecificationRepository
}

func NewGetProductSpecification(repo ports.ProductSpecificationRepository) ports.GetProductSpecificationUseCase {
	return &GetProductSpecificationUseCase{
		repo: repo,
	}
}

func (uc *GetProductSpecificationUseCase) Execute(ctx context.Context, input ports.GetProductSpecificationInput) (*domain.ProductSpecification, error) {
	return uc.repo.Get(ctx, input.ID)
}
