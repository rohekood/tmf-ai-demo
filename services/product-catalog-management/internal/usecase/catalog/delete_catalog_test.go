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

func TestDeleteCatalog_Execute(t *testing.T) {
	mockRepo := new(MockCatalogRepo)
	mockPub := new(MockEventPublisher)
	useCase := NewDeleteCatalogUseCase(mockRepo, mockPub, &repository.NoOpTransactionManager{})

	ctx := context.Background()
	id := "cat-del-1"

	existing := &domain.Catalog{ID: id, Name: "To Delete"}

	// Delete typically checks existence first or returns Not Found if not there.
	// Let's see implementation. Assuming strict delete or idempotency?
	// Usually: Get -> Delete -> Publish.
	// Or just Delete -> Publish if repo handles "not found".
	// Let's assume standard flow: Get check inside usecase or just delete.
	// Checking `delete_catalog.go`:
	/*
		func (uc *DeleteCatalogUseCase) Execute(ctx context.Context, input ports.DeleteCatalogInput) error {
			// 1. Get to verify existence and get data for event
			catalog, err := uc.repo.Get(ctx, input.ID)
			if err != nil { return err } ...
	*/

	mockRepo.On("Get", ctx, id).Return(existing, nil)
	mockRepo.On("Delete", ctx, id).Return(nil)

	mockPub.On("PublishCatalogDeleted", ctx, mock.MatchedBy(func(e domain.CatalogDeletedEvent) bool {
		return e.ID == id
	})).Return(nil)

	input := ports.DeleteCatalogInput{ID: id}
	err := useCase.Execute(ctx, input)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
