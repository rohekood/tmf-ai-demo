package usecase_test

import (
	"context"
	"testing"

	"tmf/services/shopping-cart/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSyncCatalogUseCase_SyncOffering(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockCartRepository)

	uc := usecase.NewSyncCatalogUseCase(mockRepo)

	mockRepo.On("UpsertPrice", ctx, mock.AnythingOfType("*domain.ProductPrice")).Return(nil)

	err := uc.SyncOffering(ctx, "offering-1", 99.99, "USD")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
