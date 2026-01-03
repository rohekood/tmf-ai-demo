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

func (m *MockOfferingRepo) List(ctx context.Context, filters map[string]interface{}) ([]*domain.ProductOffering, error) {
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
