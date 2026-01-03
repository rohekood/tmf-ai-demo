package catalog

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

// MockCatalogRepo is a mock of CatalogRepository
type MockCatalogRepo struct {
	mock.Mock
}

func (m *MockCatalogRepo) Create(ctx context.Context, catalog *domain.Catalog) error {
	args := m.Called(ctx, catalog)
	return args.Error(0)
}

func (m *MockCatalogRepo) Get(ctx context.Context, id string) (*domain.Catalog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Catalog), args.Error(1)
}

func (m *MockCatalogRepo) List(ctx context.Context, filters map[string]interface{}) ([]*domain.Catalog, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*domain.Catalog), args.Error(1)
}

func (m *MockCatalogRepo) Update(ctx context.Context, catalog *domain.Catalog) error {
	args := m.Called(ctx, catalog)
	return args.Error(0)
}

func (m *MockCatalogRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
