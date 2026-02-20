package specification

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
)

func TestGetProductSpecification_Execute(t *testing.T) {
	// MockSpecRepo is in mock_repo_test.go, same package
	mockRepo := new(MockSpecificationRepo)
	useCase := NewGetProductSpecification(mockRepo)

	ctx := context.Background()
	id := "spec-1"
	expected := &domain.ProductSpecification{ID: id, Name: "Found"}

	mockRepo.On("Get", ctx, id).Return(expected, nil)

	res, err := useCase.Execute(ctx, ports.GetProductSpecificationInput{ID: id})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)

	mockRepo.AssertExpectations(t)
}
