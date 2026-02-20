package offering

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteProductOffering_Execute(t *testing.T) {
	mockRepo := new(MockOfferingRepo)
	mockPub := new(MockEventPublisher)
	useCase := NewDeleteProductOfferingUseCase(mockRepo, mockPub, &repository.NoOpTransactionManager{})

	ctx := context.Background()
	id := "off-del-1"

	existing := &domain.ProductOffering{ID: id, Name: "To Delete"}

	mockRepo.On("Get", ctx, id).Return(existing, nil)
	mockRepo.On("Delete", ctx, id).Return(nil)
	mockPub.On("PublishProductOfferingDeleted", ctx, mock.MatchedBy(func(e domain.ProductOfferingDeletedEvent) bool {
		return e.ID == id
	})).Return(nil)

	err := useCase.Execute(ctx, ports.DeleteProductOfferingInput{ID: id})
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
