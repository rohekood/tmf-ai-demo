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

func TestUpdateProductSpecification_Execute(t *testing.T) {
	mockRepo := new(MockSpecificationRepo)
	mockPub := new(MockEventPublisher)
	useCase := NewUpdateProductSpecificationUseCase(mockRepo, mockPub, &repository.NoOpTransactionManager{})

	ctx := context.Background()
	id := "spec-upd-1"
	newName := "Updated Spec"

	existing := &domain.ProductSpecification{
		ID:            id,
		Name:          "Old Name",
		ProductNumber: "PN-001",
	}

	mockRepo.On("Get", ctx, id).Return(existing, nil)

	mockRepo.On("Update", ctx, mock.MatchedBy(func(s *domain.ProductSpecification) bool {
		return s.ID == id && s.Name == newName
	})).Return(nil)

	mockPub.On("PublishProductSpecificationUpdated", ctx, mock.MatchedBy(func(e domain.ProductSpecificationUpdatedEvent) bool {
		return e.ProductSpecification.Name == newName
	})).Return(nil)

	input := ports.UpdateProductSpecificationInput{
		ID:   id,
		Name: &newName,
	}

	res, err := useCase.Execute(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, newName, res.Name)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
