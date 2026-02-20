package category

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
)

func TestGetCategory_Execute(t *testing.T) {
	mockRepo := new(MockCategoryRepo)
	useCase := NewGetCategory(mockRepo)

	ctx := context.Background()
	id := "cat-1"
	expected := &domain.Category{ID: id, Name: "Found"}

	mockRepo.On("Get", ctx, id).Return(expected, nil)

	res, err := useCase.Execute(ctx, ports.GetCategoryInput{ID: id})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)

	mockRepo.AssertExpectations(t)
}
