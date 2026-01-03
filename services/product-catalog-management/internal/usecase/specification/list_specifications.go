package specification

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type ListProductSpecifications struct {
	repo ports.ProductSpecificationRepository
}

func NewListProductSpecifications(repo ports.ProductSpecificationRepository) *ListProductSpecifications {
	return &ListProductSpecifications{repo: repo}
}

func (uc *ListProductSpecifications) Execute(ctx context.Context, input ports.ListProductSpecificationsInput) ([]*domain.ProductSpecification, error) {
	return uc.repo.List(ctx, input.Filters)
}
