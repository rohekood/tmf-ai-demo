package specification

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
)

func TestListProductSpecifications_Execute(t *testing.T) {
	mockRepo := new(MockSpecificationRepo)
	useCase := NewListProductSpecifications(mockRepo)

	ctx := context.Background()
	filters := map[string]any{"name": "Spec"}
	expected := []*domain.ProductSpecification{
		{ID: "s1", Name: "Spec 1"},
	}

	mockRepo.On("List", ctx, filters).Return(expected, nil)

	res, err := useCase.Execute(ctx, ports.ListProductSpecificationsInput{Filters: filters})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)

	mockRepo.AssertExpectations(t)
}
