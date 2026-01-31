package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"tmf/services/qualification/internal/core/domain"

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

		uc := NewCheckEligibility(mockGIS, mockInv, mockCat, mockPub, logger)

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

	t.Run("Scenario 2: Success (Unqualified)", func(t *testing.T) {
		// Arrange
		mockGIS := new(MockGISClient)
		mockInv := new(MockInventoryClient)
		mockCat := new(MockCatalogClient)
		mockPub := new(MockEventPublisher)

		uc := NewCheckEligibility(mockGIS, mockInv, mockCat, mockPub, logger)

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

		uc := NewCheckEligibility(mockGIS, mockInv, mockCat, mockPub, logger)

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
}
