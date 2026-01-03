package ports

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"
)

type EventPublisher interface {
	PublishCatalogCreated(ctx context.Context, event domain.CatalogCreatedEvent) error
	PublishCategoryCreated(ctx context.Context, event domain.CategoryCreatedEvent) error
	PublishProductSpecificationCreated(ctx context.Context, event domain.ProductSpecificationCreatedEvent) error
	PublishProductOfferingCreated(ctx context.Context, event domain.ProductOfferingCreatedEvent) error

	PublishCatalogUpdated(ctx context.Context, event domain.CatalogUpdatedEvent) error
	PublishCatalogDeleted(ctx context.Context, event domain.CatalogDeletedEvent) error

	PublishCategoryUpdated(ctx context.Context, event domain.CategoryUpdatedEvent) error
	PublishCategoryDeleted(ctx context.Context, event domain.CategoryDeletedEvent) error

	PublishProductSpecificationUpdated(ctx context.Context, event domain.ProductSpecificationUpdatedEvent) error
	PublishProductSpecificationDeleted(ctx context.Context, event domain.ProductSpecificationDeletedEvent) error

	PublishProductOfferingUpdated(ctx context.Context, event domain.ProductOfferingUpdatedEvent) error
	PublishProductOfferingDeleted(ctx context.Context, event domain.ProductOfferingDeletedEvent) error
}
