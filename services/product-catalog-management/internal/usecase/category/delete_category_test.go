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

func TestDeleteCategory_Execute(t *testing.T) {
	mockRepo := new(MockCategoryRepo)
	mockPub := new(MockEventPublisher)
	useCase := NewDeleteCategoryUseCase(mockRepo, mockPub, &repository.NoOpTransactionManager{})

	ctx := context.Background()
	id := "cat-del-1"

	existing := &domain.Category{ID: id, Name: "To Delete"}

	// Delete use case calls Get first
	mockRepo.On("Get", ctx, id).Return(existing, nil)
	mockRepo.On("Delete", ctx, id).Return(nil)
	mockPub.On("PublishCategoryDeleted", ctx, mock.MatchedBy(func(e domain.CategoryDeletedEvent) bool {
		return e.ID == id
	})).Return(nil)

	err := useCase.Execute(ctx, ports.DeleteCategoryInput{ID: id})
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
