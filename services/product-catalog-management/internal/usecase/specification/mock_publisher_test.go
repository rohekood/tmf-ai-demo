package specification

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishCatalogCreated(ctx context.Context, event domain.CatalogCreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCategoryCreated(ctx context.Context, event domain.CategoryCreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishProductSpecificationCreated(ctx context.Context, event domain.ProductSpecificationCreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishProductOfferingCreated(ctx context.Context, event domain.ProductOfferingCreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCatalogUpdated(ctx context.Context, event domain.CatalogUpdatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCatalogDeleted(ctx context.Context, event domain.CatalogDeletedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCategoryUpdated(ctx context.Context, event domain.CategoryUpdatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCategoryDeleted(ctx context.Context, event domain.CategoryDeletedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishProductSpecificationUpdated(ctx context.Context, event domain.ProductSpecificationUpdatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishProductSpecificationDeleted(ctx context.Context, event domain.ProductSpecificationDeletedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishProductOfferingUpdated(ctx context.Context, event domain.ProductOfferingUpdatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishProductOfferingDeleted(ctx context.Context, event domain.ProductOfferingDeletedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}
