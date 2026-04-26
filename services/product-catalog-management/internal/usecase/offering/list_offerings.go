package offering

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type ListProductOfferings struct {
	repo ports.ProductOfferingRepository
}

func NewListProductOfferings(repo ports.ProductOfferingRepository) *ListProductOfferings {
	return &ListProductOfferings{repo: repo}
}

func (uc *ListProductOfferings) Execute(ctx context.Context, input ports.ListProductOfferingsInput) ([]*domain.ProductOffering, error) {
	filters := make(map[string]any)
	if input.Filters.Name != nil {
		filters["name"] = *input.Filters.Name
	}
	if input.Filters.Category != nil {
		filters["category"] = *input.Filters.Category
	}
	if input.Filters.MinPrice != nil {
		filters["min_price"] = *input.Filters.MinPrice
	}
	if input.Filters.MaxPrice != nil {
		filters["max_price"] = *input.Filters.MaxPrice
	}
	return uc.repo.List(ctx, filters)
}
