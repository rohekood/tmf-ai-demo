package category

import (
	"context"

	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type GetCategoryUseCase struct {
	repo ports.CategoryRepository
}

func NewGetCategory(repo ports.CategoryRepository) ports.GetCategoryUseCase {
	return &GetCategoryUseCase{
		repo: repo,
	}
}

func (uc *GetCategoryUseCase) Execute(ctx context.Context, input ports.GetCategoryInput) (*domain.Category, error) {
	return uc.repo.Get(ctx, input.ID)
}
