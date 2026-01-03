package catalog

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type GetCatalog struct {
	repo ports.CatalogRepository
}

func NewGetCatalog(repo ports.CatalogRepository) *GetCatalog {
	return &GetCatalog{repo: repo}
}

func (uc *GetCatalog) Execute(ctx context.Context, input ports.GetCatalogInput) (*domain.Catalog, error) {
	return uc.repo.Get(ctx, input.ID)
}
