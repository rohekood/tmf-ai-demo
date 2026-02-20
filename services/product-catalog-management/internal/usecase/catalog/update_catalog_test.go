package catalog

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateCatalog_Execute(t *testing.T) {
	mockRepo := new(MockCatalogRepo)
	mockPub := new(MockEventPublisher)
	useCase := NewUpdateCatalogUseCase(mockRepo, mockPub, &repository.NoOpTransactionManager{})

	ctx := context.Background()
	id := "cat-1"
	newName := "Updated Name"
	status := "Active"

	existing := &domain.Catalog{ID: id, Name: "Old Name"}

	mockRepo.On("Get", ctx, id).Return(existing, nil)

	mockRepo.On("Update", ctx, mock.MatchedBy(func(c *domain.Catalog) bool {
		return c.ID == id && c.Name == newName && c.LifecycleStatus == status
	})).Return(nil)

	mockPub.On("PublishCatalogUpdated", ctx, mock.MatchedBy(func(e domain.CatalogUpdatedEvent) bool {
		return e.Catalog.ID == id && e.Catalog.Name == newName
	})).Return(nil)

	input := ports.UpdateCatalogInput{
		ID:              id,
		Name:            &newName,
		LifecycleStatus: &status,
	}

	res, err := useCase.Execute(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, newName, res.Name)
	assert.Equal(t, status, res.LifecycleStatus)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
