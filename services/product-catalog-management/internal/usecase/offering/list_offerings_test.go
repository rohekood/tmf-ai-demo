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
	expected := []*domain.ProductOffering{
		{ID: "o1", Name: "Fiber"},
	}

	// Let's assume we filter by Name for test
	name := "Fiber"
	inputFilters := ports.ProductOfferingFilters{Name: &name}

	// Mock Expectation needs to match what Execute calls repo with.
	mockRepo.On("List", ctx, mock.MatchedBy(func(f map[string]any) bool {
		return f["name"] == "Fiber"
	})).Return(expected, nil)

	res, err := useCase.Execute(ctx, ports.ListProductOfferingsInput{Filters: inputFilters})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)

	mockRepo.AssertExpectations(t)
}
