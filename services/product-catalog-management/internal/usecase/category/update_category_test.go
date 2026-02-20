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

func TestUpdateCategory_Execute(t *testing.T) {
	mockRepo := new(MockCategoryRepo)
	mockPub := new(MockEventPublisher)
	useCase := NewUpdateCategoryUseCase(mockRepo, mockPub, &repository.NoOpTransactionManager{})

	ctx := context.Background()
	id := "cat-1"
	newName := "Updated Cat"

	existing := &domain.Category{ID: id, Name: "Old"}

	mockRepo.On("Get", ctx, id).Return(existing, nil)

	mockRepo.On("Update", ctx, mock.MatchedBy(func(c *domain.Category) bool {
		return c.ID == id && c.Name == newName
	})).Return(nil)

	mockPub.On("PublishCategoryUpdated", ctx, mock.MatchedBy(func(e domain.CategoryUpdatedEvent) bool {
		return e.Category.Name == newName
	})).Return(nil)

	input := ports.UpdateCategoryInput{
		ID:   id,
		Name: &newName,
	}

	res, err := useCase.Execute(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, newName, res.Name)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
