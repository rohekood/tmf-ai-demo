package publisher_test

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/adapter/publisher"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestOutboxPublisher_Events(t *testing.T) {
	pub := publisher.NewOutboxPublisher(sharedDB)
	ctx := context.WithValue(context.Background(), domain.UserContextKey, "testuser")

	tests := []struct {
		name       string
		action     func() error
		routingKey string
	}{
		{
			name: "PublishCatalogCreated",
			action: func() error {
				return pub.PublishCatalogCreated(ctx, domain.CatalogCreatedEvent{Catalog: &domain.Catalog{ID: "cat1"}})
			},
			routingKey: "evt.catalog.catalog.created",
		},
		{
			name: "PublishCategoryCreated",
			action: func() error {
				return pub.PublishCategoryCreated(ctx, domain.CategoryCreatedEvent{Category: &domain.Category{ID: "cat1"}})
			},
			routingKey: "evt.catalog.category.created",
		},
		{
			name: "PublishProductSpecificationCreated",
			action: func() error {
				return pub.PublishProductSpecificationCreated(ctx, domain.ProductSpecificationCreatedEvent{ProductSpecification: &domain.ProductSpecification{ID: "spec1"}})
			},
			routingKey: "evt.catalog.specification.created",
		},
		{
			name: "PublishProductOfferingCreated",
			action: func() error {
				return pub.PublishProductOfferingCreated(ctx, domain.ProductOfferingCreatedEvent{ProductOffering: &domain.ProductOffering{ID: "off1"}})
			},
			routingKey: "evt.catalog.offering.created",
		},
		{
			name: "PublishCatalogUpdated",
			action: func() error {
				return pub.PublishCatalogUpdated(ctx, domain.CatalogUpdatedEvent{Catalog: &domain.Catalog{ID: "cat1"}})
			},
			routingKey: "evt.catalog.catalog.updated",
		},
		{
			name: "PublishCatalogDeleted",
			action: func() error {
				return pub.PublishCatalogDeleted(ctx, domain.CatalogDeletedEvent{ID: "cat1"})
			},
			routingKey: "evt.catalog.catalog.deleted",
		},
		{
			name: "PublishCategoryUpdated",
			action: func() error {
				return pub.PublishCategoryUpdated(ctx, domain.CategoryUpdatedEvent{Category: &domain.Category{ID: "cat1"}})
			},
			routingKey: "evt.catalog.category.updated",
		},
		{
			name: "PublishCategoryDeleted",
			action: func() error {
				return pub.PublishCategoryDeleted(ctx, domain.CategoryDeletedEvent{ID: "cat1"})
			},
			routingKey: "evt.catalog.category.deleted",
		},
		{
			name: "PublishProductSpecificationUpdated",
			action: func() error {
				return pub.PublishProductSpecificationUpdated(ctx, domain.ProductSpecificationUpdatedEvent{ProductSpecification: &domain.ProductSpecification{ID: "spec1"}})
			},
			routingKey: "evt.catalog.specification.updated",
		},
		{
			name: "PublishProductSpecificationDeleted",
			action: func() error {
				return pub.PublishProductSpecificationDeleted(ctx, domain.ProductSpecificationDeletedEvent{ID: "spec1"})
			},
			routingKey: "evt.catalog.specification.deleted",
		},
		{
			name: "PublishProductOfferingUpdated",
			action: func() error {
				return pub.PublishProductOfferingUpdated(ctx, domain.ProductOfferingUpdatedEvent{ProductOffering: &domain.ProductOffering{ID: "off1"}})
			},
			routingKey: "evt.catalog.offering.updated",
		},
		{
			name: "PublishProductOfferingDeleted",
			action: func() error {
				return pub.PublishProductOfferingDeleted(ctx, domain.ProductOfferingDeletedEvent{ID: "off1"})
			},
			routingKey: "evt.catalog.offering.deleted",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Count existing outbox events
			var initialCount int64
			sharedDB.Model(&repository.OutboxEventModel{}).Count(&initialCount)

			err := tc.action()
			assert.NoError(t, err)

			var finalCount int64
			sharedDB.Model(&repository.OutboxEventModel{}).Count(&finalCount)

			assert.Equal(t, initialCount+1, finalCount)

			// Get the newly created record
			var ev repository.OutboxEventModel
			sharedDB.Order("created_at desc").First(&ev)

			assert.Equal(t, tc.routingKey, ev.RoutingKey)
			assert.Equal(t, repository.StatusPending, ev.Status)
		})
	}
}
