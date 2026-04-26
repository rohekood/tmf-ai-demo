package specification

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockSpecificationRepo struct {
	mock.Mock
}

func (m *MockSpecificationRepo) Create(ctx context.Context, spec *domain.ProductSpecification) error {
	args := m.Called(ctx, spec)
	return args.Error(0)
}

func (m *MockSpecificationRepo) Get(ctx context.Context, id string) (*domain.ProductSpecification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.ProductSpecification), args.Error(1)
}

func (m *MockSpecificationRepo) List(ctx context.Context, filters map[string]any) ([]*domain.ProductSpecification, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*domain.ProductSpecification), args.Error(1)
}

func (m *MockSpecificationRepo) Update(ctx context.Context, spec *domain.ProductSpecification) error {
	args := m.Called(ctx, spec)
	return args.Error(0)
}

func (m *MockSpecificationRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
