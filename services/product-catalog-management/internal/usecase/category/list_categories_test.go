package category

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
)

func TestListCategories_Execute(t *testing.T) {
	mockRepo := new(MockCategoryRepo)
	useCase := NewListCategories(mockRepo)

	ctx := context.Background()
	filters := map[string]any{"name": "Test"}
	expected := []*domain.Category{
		{ID: "c1", Name: "Test Cat"},
	}

	mockRepo.On("List", ctx, filters).Return(expected, nil)

	res, err := useCase.Execute(ctx, ports.ListCategoriesInput{Filters: filters})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)

	mockRepo.AssertExpectations(t)
}
