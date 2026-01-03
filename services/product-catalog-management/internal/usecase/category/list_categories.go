package category

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type ListCategories struct {
	repo ports.CategoryRepository
}

func NewListCategories(repo ports.CategoryRepository) *ListCategories {
	return &ListCategories{repo: repo}
}

func (uc *ListCategories) Execute(ctx context.Context, input ports.ListCategoriesInput) ([]*domain.Category, error) {
	return uc.repo.List(ctx, input.Filters)
}
