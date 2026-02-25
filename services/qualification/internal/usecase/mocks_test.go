package usecase

import (
	"context"
	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"

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
	if args.Get(0) != nil {
		return args.Get(0).([]domain.EligibleCategory), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishEligibilityChecked(ctx context.Context, result domain.EligibilityResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

// New mocks for session and pricing
type MockSessionRepository struct{}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.QualificationSession) (string, error) {
	return "mock-session-id", nil
}

func (m *MockSessionRepository) Get(ctx context.Context, sessionID string) (*domain.QualificationSession, error) {
	return nil, nil
}

func (m *MockSessionRepository) Update(ctx context.Context, session *domain.QualificationSession) error {
	return nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID string) error {
	return nil
}

func (m *MockSessionRepository) FindExpired(ctx context.Context) ([]*domain.QualificationSession, error) {
	return nil, nil
}

type MockCustomerPricingClient struct{}

func (m *MockCustomerPricingClient) GetCustomer(ctx context.Context, customerID string) (*ports.Customer, error) {
	return &ports.Customer{
		ID:      customerID,
		Tier:    "Standard",
		Segment: "Residential",
		Status:  "Active",
	}, nil
}

type MockCatalogPricingClient struct{}

func (m *MockCatalogPricingClient) GetOffering(ctx context.Context, offeringID string) (*ports.Offering, error) {
	return &ports.Offering{
		ID:        offeringID,
		Name:      "Test Offering",
		BasePrice: 100.0,
		Currency:  "EUR",
	}, nil
}
