package catalog

import (
	"context"
	"testing"
	"time"

	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
)

func TestGetCatalog_Execute(t *testing.T) {
	mockRepo := new(MockCatalogRepo)
	useCase := NewGetCatalog(mockRepo)
	ctx := context.Background()

	expectedCatalog := &domain.Catalog{
		ID:              "cat-1",
		Name:            "Test Catalog",
		LifecycleStatus: "Active",
		LastUpdate:      time.Now(),
	}

	// 1. Success
	mockRepo.On("Get", ctx, "cat-1").Return(expectedCatalog, nil)

	result, err := useCase.Execute(ctx, ports.GetCatalogInput{ID: "cat-1"})
	assert.NoError(t, err)
	assert.Equal(t, expectedCatalog, result)

	// 2. Not Found
	mockRepo.On("Get", ctx, "cat-unknown").Return((*domain.Catalog)(nil), domain.ErrNotFound)

	result2, err := useCase.Execute(ctx, ports.GetCatalogInput{ID: "cat-unknown"})
	assert.Error(t, err)
	assert.Nil(t, result2)
	assert.Equal(t, domain.ErrNotFound, err)

	mockRepo.AssertExpectations(t)
}
