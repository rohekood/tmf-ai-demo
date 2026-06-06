package ports

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
)

type CreateCatalogUseCase interface {
	Execute(ctx context.Context, input CreateCatalogInput) (*domain.Catalog, error)
}

type CreateCatalogInput struct {
	Name            string
	Description     string
	ValidFor        domain.TimePeriod
	LifecycleStatus string
}

type ListCatalogsUseCase interface {
	Execute(ctx context.Context, input ListCatalogsInput) ([]*domain.Catalog, error)
}

type ListCatalogsInput struct {
	Filters map[string]any
}

type GetCatalogUseCase interface {
	Execute(ctx context.Context, input GetCatalogInput) (*domain.Catalog, error)
}

type GetCatalogInput struct {
	ID string
}

type CreateCategoryUseCase interface {
	Execute(ctx context.Context, input CreateCategoryInput) (*domain.Category, error)
}

type CreateCategoryInput struct {
	Name            string
	Description     string
	ParentID        *string
	IsRoot          bool
	CatalogID       *string
	ValidFor        domain.TimePeriod
	LifecycleStatus string
}

type ListCategoriesUseCase interface {
	Execute(ctx context.Context, input ListCategoriesInput) ([]*domain.Category, error)
}

type ListCategoriesInput struct {
	Filters map[string]any
}

type GetCategoryUseCase interface {
	Execute(ctx context.Context, input GetCategoryInput) (*domain.Category, error)
}

type GetCategoryInput struct {
	ID string
}

type CreateProductSpecificationUseCase interface {
	Execute(ctx context.Context, input CreateProductSpecificationInput) (*domain.ProductSpecification, error)
}

type CreateProductSpecificationInput struct {
	Name            string
	ProductNumber   string
	Description     string
	IsBundle        bool
	LifecycleStatus string
	ValidFor        domain.TimePeriod
	Characteristics map[string]domain.ProductSpecCharacteristic
}

type ListProductSpecificationsUseCase interface {
	Execute(ctx context.Context, input ListProductSpecificationsInput) ([]*domain.ProductSpecification, error)
}

type ListProductSpecificationsInput struct {
	Filters map[string]any
}

type GetProductSpecificationUseCase interface {
	Execute(ctx context.Context, input GetProductSpecificationInput) (*domain.ProductSpecification, error)
}

type GetProductSpecificationInput struct {
	ID string
}

type CreateProductOfferingUseCase interface {
	Execute(ctx context.Context, input CreateProductOfferingInput) (*domain.ProductOffering, error)
}

type CreateProductOfferingInput struct {
	Name            string
	Description     string
	IsBundle        bool
	IsSellable      bool
	LifecycleStatus string
	ValidFor        domain.TimePeriod
	ProductSpecID   *string
	CategoryIDs     []string
	Prices          []domain.ProductOfferingPrice
	Attachments     []domain.Attachment
}

type ListProductOfferingsUseCase interface {
	Execute(ctx context.Context, input ListProductOfferingsInput) ([]*domain.ProductOffering, error)
}

type ListProductOfferingsInput struct {
	Filters ProductOfferingFilters
}

type ProductOfferingFilters struct {
	Name     *string
	Category *string // Filter by Category ID
	MinPrice *float64
	MaxPrice *float64
}

type GetProductOfferingUseCase interface {
	Execute(ctx context.Context, input GetProductOfferingInput) (*domain.ProductOffering, error)
}

type GetProductOfferingInput struct {
	ID     string
	Enrich bool
}

// Catalog
type UpdateCatalogUseCase interface {
	Execute(ctx context.Context, input UpdateCatalogInput) (*domain.Catalog, error)
}

type UpdateCatalogInput struct {
	ID              string
	Name            *string
	Description     *string
	ValidFor        *domain.TimePeriod
	LifecycleStatus *string
}

type DeleteCatalogUseCase interface {
	Execute(ctx context.Context, input DeleteCatalogInput) error
}

type DeleteCatalogInput struct {
	ID string
}

// Category
type UpdateCategoryUseCase interface {
	Execute(ctx context.Context, input UpdateCategoryInput) (*domain.Category, error)
}

type UpdateCategoryInput struct {
	ID              string
	Name            *string
	Description     *string
	ParentID        *string
	IsRoot          *bool
	CatalogID       *string
	ValidFor        *domain.TimePeriod
	LifecycleStatus *string
}

type DeleteCategoryUseCase interface {
	Execute(ctx context.Context, input DeleteCategoryInput) error
}

type DeleteCategoryInput struct {
	ID string
}

// Product Specification
type UpdateProductSpecificationUseCase interface {
	Execute(ctx context.Context, input UpdateProductSpecificationInput) (*domain.ProductSpecification, error)
}

type UpdateProductSpecificationInput struct {
	ID              string
	Name            *string
	Description     *string
	ProductNumber   *string
	IsBundle        *bool
	LifecycleStatus *string
	ValidFor        *domain.TimePeriod
	Characteristics map[string]domain.ProductSpecCharacteristic
}

type DeleteProductSpecificationUseCase interface {
	Execute(ctx context.Context, input DeleteProductSpecificationInput) error
}

type DeleteProductSpecificationInput struct {
	ID string
}

// Product Offering
type UpdateProductOfferingUseCase interface {
	Execute(ctx context.Context, input UpdateProductOfferingInput) (*domain.ProductOffering, error)
}

type UpdateProductOfferingInput struct {
	ID              string
	Name            *string
	Description     *string
	LifecycleStatus *string
	ValidFor        *domain.TimePeriod
	IsBundle        *bool
	IsSellable      *bool
	CategoryIDs     []string
	Prices          []domain.ProductOfferingPrice
	Attachments     []domain.Attachment
}

type DeleteProductOfferingUseCase interface {
	Execute(ctx context.Context, input DeleteProductOfferingInput) error
}

type DeleteProductOfferingInput struct {
	ID string
}

// ... other use case interfaces can be added here or in separate files
// For now, defining the main pattern
