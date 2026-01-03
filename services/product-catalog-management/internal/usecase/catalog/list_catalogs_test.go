package catalog

import (
	"context"
	"testing"

	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
)

func TestListCatalogs_Execute(t *testing.T) {
	mockRepo := new(MockCatalogRepo)
	useCase := NewListCatalogs(mockRepo)
	ctx := context.Background()

	expectedList := []*domain.Catalog{
		{ID: "cat-1", Name: "Catalog 1"},
		{ID: "cat-2", Name: "Catalog 2"},
	}

	filters := map[string]interface{}{"name": "Catalog%"}
	input := ports.ListCatalogsInput{Filters: filters}

	// 1. Success
	mockRepo.On("List", ctx, filters).Return(expectedList, nil)

	result, err := useCase.Execute(ctx, input)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	mockRepo.AssertExpectations(t)
}
