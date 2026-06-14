package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckEligibilityUseCase_Check(t *testing.T) {
	// Logger that discards output
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Common test data
	cmd := domain.CheckEligibilityCommand{
		CorrelationID: "test-correlation-id",
		Address: domain.Address{
			Street: "Main St",
			Number: "123",
			City:   "Tech City",
		},
		CategoryFilter: []string{"Fiber"},
	}

	offers := []domain.EligibleCategory{
		{
			ID:   "offer-1",
			Name: "Fiber 100",
		},
	}

	t.Run("Scenario 1: Success (Qualified)", func(t *testing.T) {
		// Arrange
		mockGIS := new(MockGISClient)
		mockInv := new(MockInventoryClient)
		mockCat := new(MockCatalogClient)
		mockPub := new(MockEventPublisher)

		mockSessionRepo := &MockSessionRepository{}
		mockCustomerClient := &MockCustomerPricingClient{}
		mockCatalogPricing := &MockCatalogPricingClient{}
		uc := NewCheckEligibility(mockGIS, mockInv, mockCat, mockPub, mockSessionRepo, mockCustomerClient, mockCatalogPricing, logger)

		// Expectation: GIS returns true (in polygon)
		mockGIS.On("CheckPolygon", mock.Anything, cmd.Address).Return(true, nil)

		// Expectation: Inventory returns 5 ports
		mockInv.On("GetPortCapacity", mock.Anything, cmd.Address).Return(5, nil)

		// Expectation: Catalog returns offers
		mockCat.On("GetOffersByCategory", mock.Anything, "Fiber").Return(offers, nil)

		// Expectation: Publisher called with Status="Qualified"
		mockPub.On("PublishEligibilityChecked", mock.Anything, mock.MatchedBy(func(result domain.EligibilityResult) bool {
			return result.Status == domain.StatusQualified &&
				result.CorrelationID == cmd.CorrelationID &&
				len(result.EligibleCategories) > 0
		})).Return(nil)

		// Act
		err := uc.Execute(context.Background(), cmd)

		// Assert
		assert.NoError(t, err)
		mockGIS.AssertExpectations(t)
		mockInv.AssertExpectations(t)
		mockCat.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("Anonymous caller gets generic-priced offers", func(t *testing.T) {
		mockGIS := new(MockGISClient)
		mockInv := new(MockInventoryClient)
		mockCat := new(MockCatalogClient)
		mockPub := new(MockEventPublisher)

		mockSessionRepo := &MockSessionRepository{}
		mockCustomerClient := &MockCustomerPricingClient{}
		mockCatalogPricing := &MockCatalogPricingClient{}
		uc := NewCheckEligibility(mockGIS, mockInv, mockCat, mockPub, mockSessionRepo, mockCustomerClient, mockCatalogPricing, logger)

		mockGIS.On("CheckPolygon", mock.Anything, cmd.Address).Return(true, nil)
		mockInv.On("GetPortCapacity", mock.Anything, cmd.Address).Return(5, nil)
		mockCat.On("GetOffersByCategory", mock.Anything, "Fiber").Return(offers, nil)

		var published domain.EligibilityResult
		mockPub.On("PublishEligibilityChecked", mock.Anything, mock.MatchedBy(func(result domain.EligibilityResult) bool {
			published = result
			return true
		})).Return(nil)

		// cmd has no CustomerID -> anonymous path.
		err := uc.Execute(context.Background(), cmd)

		assert.NoError(t, err)
		assert.Len(t, published.QualifiedOffers, 1)
		assert.Equal(t, domain.PriceTypeGeneric, published.QualifiedOffers[0].PriceType)
		assert.Equal(t, 100.0, published.QualifiedOffers[0].Price.Amount) // base price, no discount
	})

	t.Run("Authenticated caller gets customer-priced offers", func(t *testing.T) {
		mockGIS := new(MockGISClient)
		mockInv := new(MockInventoryClient)
		mockCat := new(MockCatalogClient)
		mockPub := new(MockEventPublisher)

		mockSessionRepo := &MockSessionRepository{}
		mockCustomerClient := &MockCustomerPricingClient{}
		mockCatalogPricing := &MockCatalogPricingClient{}
		uc := NewCheckEligibility(mockGIS, mockInv, mockCat, mockPub, mockSessionRepo, mockCustomerClient, mockCatalogPricing, logger)

		authCmd := cmd
		authCmd.CustomerID = "cust-1"

		mockGIS.On("CheckPolygon", mock.Anything, authCmd.Address).Return(true, nil)
		mockInv.On("GetPortCapacity", mock.Anything, authCmd.Address).Return(5, nil)
		mockCat.On("GetOffersByCategory", mock.Anything, "Fiber").Return(offers, nil)

		var published domain.EligibilityResult
		mockPub.On("PublishEligibilityChecked", mock.Anything, mock.MatchedBy(func(result domain.EligibilityResult) bool {
			published = result
			return true
		})).Return(nil)

		err := uc.Execute(context.Background(), authCmd)

		assert.NoError(t, err)
		assert.Len(t, published.QualifiedOffers, 1)
		assert.Equal(t, domain.PriceTypeCustomer, published.QualifiedOffers[0].PriceType)
	})

	t.Run("Scenario 2: Success (Unqualified)", func(t *testing.T) {
		// Arrange
		mockGIS := new(MockGISClient)
		mockInv := new(MockInventoryClient)
		mockCat := new(MockCatalogClient)
		mockPub := new(MockEventPublisher)

		mockSessionRepo := &MockSessionRepository{}
		mockCustomerClient := &MockCustomerPricingClient{}
		mockCatalogPricing := &MockCatalogPricingClient{}
		uc := NewCheckEligibility(mockGIS, mockInv, mockCat, mockPub, mockSessionRepo, mockCustomerClient, mockCatalogPricing, logger)

		// Expectation: GIS returns true
		mockGIS.On("CheckPolygon", mock.Anything, cmd.Address).Return(true, nil)

		// Expectation: Inventory returns 0 ports
		mockInv.On("GetPortCapacity", mock.Anything, cmd.Address).Return(0, nil)

		// Expectation: Catalog returns offers (still queried in parallel)
		mockCat.On("GetOffersByCategory", mock.Anything, "Fiber").Return(offers, nil)

		// Expectation: Publisher called with Status="Unqualified"
		mockPub.On("PublishEligibilityChecked", mock.Anything, mock.MatchedBy(func(result domain.EligibilityResult) bool {
			return result.Status == domain.StatusUnqualified &&
				result.CorrelationID == cmd.CorrelationID &&
				result.UnavailabilityReason == "No network capacity available"
		})).Return(nil)

		// Act
		err := uc.Execute(context.Background(), cmd)

		// Assert
		assert.NoError(t, err)
		mockGIS.AssertExpectations(t)
		mockInv.AssertExpectations(t)
		mockCat.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("Scenario 3: Infrastructure Failure", func(t *testing.T) {
		// Arrange
		mockGIS := new(MockGISClient)
		mockInv := new(MockInventoryClient)
		mockCat := new(MockCatalogClient)
		mockPub := new(MockEventPublisher)

		mockSessionRepo := &MockSessionRepository{}
		mockCustomerClient := &MockCustomerPricingClient{}
		mockCatalogPricing := &MockCatalogPricingClient{}
		uc := NewCheckEligibility(mockGIS, mockInv, mockCat, mockPub, mockSessionRepo, mockCustomerClient, mockCatalogPricing, logger)

		// Expectation: GIS returns error
		gisErr := errors.New("GIS service down")
		mockGIS.On("CheckPolygon", mock.Anything, cmd.Address).Return(false, gisErr)

		// Expectation: Inventory and Catalog might be called because of parallel execution.
		// However, if we want to ensure robustness, we should probably allow them to be called or wait.
		// Since errgroup cancels context on first error, the others might be canceled or return.
		// For simplicity in mocking, let's assume they might be called or context canceled.
		// But in this specific test setup, the prompt implies "GIS returns error" -> "PublishError".
		// To avoid "Unexpected call" errors if goroutines race, we can use .Maybe() or just mock them to return success/error.
		// But errgroup.Wait() returns the first non-nil error.

		// To make it deterministic, we can make the other mocks return successfully or sleep.
		// But since we are testing "GIS returns error", let's mock the others to return values (as they run in parallel).
		mockInv.On("GetPortCapacity", mock.Anything, cmd.Address).Return(5, nil).Maybe()
		mockCat.On("GetOffersByCategory", mock.Anything, "Fiber").Return(offers, nil).Maybe()

		// Expectation: Publisher called with Status="Error"
		mockPub.On("PublishEligibilityChecked", mock.Anything, mock.MatchedBy(func(result domain.EligibilityResult) bool {
			return result.Status == domain.StatusError &&
				result.CorrelationID == cmd.CorrelationID
		})).Return(nil)

		// Act
		err := uc.Execute(context.Background(), cmd)

		// Assert
		assert.NoError(t, err) // The Execute method suppresses the error internally and publishes it?
		// Let's check implementation of Execute:
		// if err := g.Wait(); err != nil { ... return uc.publishResult(...) }
		// publishResult returns uc.publisher.PublishEligibilityChecked(...)
		// Since our mock publisher returns nil, Execute should return nil.

		// We need to wait a bit to ensure async goroutines finish if we want to be strict, but Wait() handles that.

		mockGIS.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})

	t.Run("Catalog Error", func(t *testing.T) {
		mockGis := new(MockGISClient)
		mockInv := new(MockInventoryClient)
		mockCat := new(MockCatalogClient)
		mockPub := new(MockEventPublisher)
		mockSess := new(MockSessionRepository)
		mockCustPricing := new(MockCustomerPricingClient)
		mockCatPricing := new(MockCatalogPricingClient)

		uc := NewCheckEligibility(mockGis, mockInv, mockCat, mockPub, mockSess, mockCustPricing, mockCatPricing, logger)

		cmd := domain.CheckEligibilityCommand{
			Address:        domain.Address{City: "Berlin"},
			CategoryFilter: []string{"Internet"},
		}

		mockGis.On("CheckPolygon", mock.Anything, mock.Anything).Return(true, nil)
		mockInv.On("GetPortCapacity", mock.Anything, mock.Anything).Return(10, nil)
		mockCat.On("GetOffersByCategory", mock.Anything, "Internet").Return(nil, errors.New("catalog err"))

		mockPub.On("PublishEligibilityChecked", mock.Anything, mock.Anything).Return(nil)

		err := uc.Execute(context.Background(), cmd)
		assert.NoError(t, err) // publishResult returns nil if successful
	})

	t.Run("Address outside service area", func(t *testing.T) {
		mockGis := new(MockGISClient)
		mockInv := new(MockInventoryClient)
		mockCat := new(MockCatalogClient)
		mockPub := new(MockEventPublisher)
		mockSess := new(MockSessionRepository)
		mockCustPricing := new(MockCustomerPricingClient)
		mockCatPricing := new(MockCatalogPricingClient)

		uc := NewCheckEligibility(mockGis, mockInv, mockCat, mockPub, mockSess, mockCustPricing, mockCatPricing, logger)

		cmd := domain.CheckEligibilityCommand{
			Address: domain.Address{City: "Nowhere"},
		}

		mockGis.On("CheckPolygon", mock.Anything, mock.Anything).Return(false, nil)
		mockInv.On("GetPortCapacity", mock.Anything, mock.Anything).Return(10, nil)
		mockCat.On("GetOffersByCategory", mock.Anything, mock.Anything).Return([]domain.EligibleCategory{}, nil)

		mockPub.On("PublishEligibilityChecked", mock.Anything, mock.MatchedBy(func(r domain.EligibilityResult) bool {
			return r.Status == domain.StatusUnqualified && r.UnavailabilityReason == "Address outside service area"
		})).Return(nil)

		err := uc.Execute(context.Background(), cmd)
		assert.NoError(t, err)
	})

	t.Run("Pricing Error and Session Error", func(t *testing.T) {
		mockGis := new(MockGISClient)
		mockInv := new(MockInventoryClient)
		mockCat := new(MockCatalogClient)
		mockPub := new(MockEventPublisher)

		// Custom mock implementations for Scenario 6
		mockSess := &mockSessErr{MockSessionRepository{}}
		mockCustPricing := &MockCustomerPricingClient{}
		mockCatPricing := &mockCatPricingErr{MockCatalogPricingClient{}}

		uc := NewCheckEligibility(mockGis, mockInv, mockCat, mockPub, mockSess, mockCustPricing, mockCatPricing, logger)

		cmd := domain.CheckEligibilityCommand{
			Address:    domain.Address{City: "Berlin"},
			CustomerID: "cust-1",
		}

		mockGis.On("CheckPolygon", mock.Anything, mock.Anything).Return(true, nil)
		mockInv.On("GetPortCapacity", mock.Anything, mock.Anything).Return(10, nil)

		offers := []domain.EligibleCategory{
			{ID: "off-1", Name: "Offer 1"},
			{ID: "off-err", Name: "Offer Err"},
		}
		mockCat.On("GetOffersByCategory", mock.Anything, mock.Anything).Return(offers, nil)

		// We removed the invalid .On() calls. The custom classes mockCatPricingErr and
		// mockSessErr handle returning errors automatically when called.
		mockPub.On("PublishEligibilityChecked", mock.Anything, mock.Anything).Return(nil)

		err := uc.Execute(context.Background(), cmd)
		assert.NoError(t, err)
	})
}

// Custom error mocks for Scenario 6
type mockSessErr struct {
	MockSessionRepository
}

func (m *mockSessErr) Create(ctx context.Context, session *domain.QualificationSession) (string, error) {
	return "", errors.New("session err")
}

type mockCatPricingErr struct {
	MockCatalogPricingClient
}

func (m *mockCatPricingErr) GetOffering(ctx context.Context, offeringID string) (*ports.Offering, error) {
	if offeringID == "off-err" {
		return nil, errors.New("pricing err")
	}
	return &ports.Offering{
		ID:        offeringID,
		Name:      "Offer 1",
		BasePrice: 10.0,
		Currency:  "EUR",
	}, nil
}
