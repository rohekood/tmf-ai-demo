package ports

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
)

type CatalogRepository interface {
	Create(ctx context.Context, catalog *domain.Catalog) error
	Get(ctx context.Context, id string) (*domain.Catalog, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*domain.Catalog, error)
	Update(ctx context.Context, catalog *domain.Catalog) error
	Delete(ctx context.Context, id string) error
}

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	Get(ctx context.Context, id string) (*domain.Category, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id string) error
}

type ProductSpecificationRepository interface {
	Create(ctx context.Context, spec *domain.ProductSpecification) error
	Get(ctx context.Context, id string) (*domain.ProductSpecification, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*domain.ProductSpecification, error)
	Update(ctx context.Context, spec *domain.ProductSpecification) error
	Delete(ctx context.Context, id string) error
}

type ProductOfferingRepository interface {
	Create(ctx context.Context, offering *domain.ProductOffering) error
	Get(ctx context.Context, id string) (*domain.ProductOffering, error)
	List(ctx context.Context, filters map[string]interface{}) ([]*domain.ProductOffering, error)
	Update(ctx context.Context, offering *domain.ProductOffering) error
	Delete(ctx context.Context, id string) error
}
