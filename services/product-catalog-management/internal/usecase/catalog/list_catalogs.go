package catalog

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type ListCatalogs struct {
	repo ports.CatalogRepository
}

func NewListCatalogs(repo ports.CatalogRepository) *ListCatalogs {
	return &ListCatalogs{repo: repo}
}

func (uc *ListCatalogs) Execute(ctx context.Context, input ports.ListCatalogsInput) ([]*domain.Catalog, error) {
	return uc.repo.List(ctx, input.Filters)
}
