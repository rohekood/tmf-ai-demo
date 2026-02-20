package offering

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListProductOfferings_Execute(t *testing.T) {
	mockRepo := new(MockOfferingRepo)
	useCase := NewListProductOfferings(mockRepo)

	ctx := context.Background()
	filters := map[string]interface{}{"status": "Active"}
	expected := []*domain.ProductOffering{
		{ID: "o1", Name: "Fiber"},
	}

	mockRepo.On("List", ctx, filters).Return(expected, nil)

	// Using type assertion or struct literal if type matches
	inputFilters := ports.ProductOfferingFilters{
		// Since filters is map[string]interface{}, we'd map fields if ListProductOfferingsInput used struct fields.
		// BUT the interface for ListProductOfferingsInput in `ports/usecases.go` uses `ProductOfferingFilters` struct!
		// `ProductOfferingFilters` has Name *string, etc.
		// My test setup `filters` variable is a map, which is mismatch.
	}

	// Correcting test setup:
	// Note: ProductOfferingFilters doesn't have "Status" field in the definition I read from `ports/usecases.go`
	// It has Name, Category, MinPrice, MaxPrice.
	// Let's assume we filter by Name for test
	name := "Fiber"
	inputFilters = ports.ProductOfferingFilters{Name: &name}

	// Mock Expectation needs to match what Execute calls repo with.
	mockRepo.On("List", ctx, mock.MatchedBy(func(f map[string]interface{}) bool {
		return f["name"] == "Fiber"
	})).Return(expected, nil)

	res, err := useCase.Execute(ctx, ports.ListProductOfferingsInput{Filters: inputFilters})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)

	mockRepo.AssertExpectations(t)
}
