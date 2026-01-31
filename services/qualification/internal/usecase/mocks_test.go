package usecase

import (
	"context"
	"tmf/services/qualification/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockGISClient struct {
	mock.Mock
}

func (m *MockGISClient) CheckPolygon(ctx context.Context, addr domain.Address) (bool, error) {
	args := m.Called(ctx, addr)
	return args.Bool(0), args.Error(1)
}

type MockInventoryClient struct {
	mock.Mock
}

func (m *MockInventoryClient) GetPortCapacity(ctx context.Context, addr domain.Address) (int, error) {
	args := m.Called(ctx, addr)
	return args.Int(0), args.Error(1)
}

type MockCatalogClient struct {
	mock.Mock
}

func (m *MockCatalogClient) GetOffersByCategory(ctx context.Context, category string) ([]domain.EligibleCategory, error) {
	args := m.Called(ctx, category)
	return args.Get(0).([]domain.EligibleCategory), args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishEligibilityChecked(ctx context.Context, result domain.EligibilityResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}
