package specification

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteProductSpecification_Execute(t *testing.T) {
	mockRepo := new(MockSpecificationRepo)
	mockPub := new(MockEventPublisher)
	useCase := NewDeleteProductSpecificationUseCase(mockRepo, mockPub, &repository.NoOpTransactionManager{})

	ctx := context.Background()
	id := "spec-del-1"

	existing := &domain.ProductSpecification{ID: id, Name: "To Delete"}

	mockRepo.On("Get", ctx, id).Return(existing, nil)
	mockRepo.On("Delete", ctx, id).Return(nil)
	mockPub.On("PublishProductSpecificationDeleted", ctx, mock.MatchedBy(func(e domain.ProductSpecificationDeletedEvent) bool {
		return e.ID == id
	})).Return(nil)

	err := useCase.Execute(ctx, ports.DeleteProductSpecificationInput{ID: id})
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
