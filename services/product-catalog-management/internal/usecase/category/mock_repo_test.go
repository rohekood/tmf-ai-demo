package category

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockCategoryRepo struct {
	mock.Mock
}

func (m *MockCategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepo) Get(ctx context.Context, id string) (*domain.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *MockCategoryRepo) List(ctx context.Context, filters map[string]interface{}) ([]*domain.Category, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*domain.Category), args.Error(1)
}

func (m *MockCategoryRepo) Update(ctx context.Context, category *domain.Category) error {
	args := m.Called(ctx, category)
	return args.Error(0)
}

func (m *MockCategoryRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
