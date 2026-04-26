package offering

import (
	"context"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockOfferingRepo struct {
	mock.Mock
}

func (m *MockOfferingRepo) Create(ctx context.Context, offering *domain.ProductOffering) error {
	args := m.Called(ctx, offering)
	return args.Error(0)
}

func (m *MockOfferingRepo) Get(ctx context.Context, id string) (*domain.ProductOffering, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.ProductOffering), args.Error(1)
}

func (m *MockOfferingRepo) List(ctx context.Context, filters map[string]any) ([]*domain.ProductOffering, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*domain.ProductOffering), args.Error(1)
}

func (m *MockOfferingRepo) Update(ctx context.Context, offering *domain.ProductOffering) error {
	args := m.Called(ctx, offering)
	return args.Error(0)
}

func (m *MockOfferingRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSpecRepo struct {
	mock.Mock
}

func (m *MockSpecRepo) Create(ctx context.Context, spec *domain.ProductSpecification) error {
	args := m.Called(ctx, spec)
	return args.Error(0)
}

func (m *MockSpecRepo) Get(ctx context.Context, id string) (*domain.ProductSpecification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProductSpecification), args.Error(1)
}

func (m *MockSpecRepo) List(ctx context.Context, filters map[string]any) ([]*domain.ProductSpecification, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*domain.ProductSpecification), args.Error(1)
}

func (m *MockSpecRepo) Update(ctx context.Context, spec *domain.ProductSpecification) error {
	args := m.Called(ctx, spec)
	return args.Error(0)
}

func (m *MockSpecRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
