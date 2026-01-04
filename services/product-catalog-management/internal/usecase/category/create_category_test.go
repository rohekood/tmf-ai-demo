package category

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/adapter/repository"

	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateCategory_Execute(t *testing.T) {
	mockRepo := new(MockCategoryRepo)
	mockPub := new(MockEventPublisher)
	useCase := NewCreateCategory(mockRepo, mockPub, &repository.NoOpTransactionManager{})

	input := ports.CreateCategoryInput{
		Name:        "Smartphones",
		Description: "A test category",
		IsRoot:      true,
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Category) bool {
		return c.Name == "Smartphones" && c.CatalogID == nil && c.ParentID == nil
	})).Return(nil)

	mockPub.On("PublishCategoryCreated", mock.Anything, mock.MatchedBy(func(e domain.CategoryCreatedEvent) bool {
		return e.Category.Name == "Smartphones"
	})).Return(nil)

	ctx := context.Background()
	result, err := useCase.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Smartphones", result.Name)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
