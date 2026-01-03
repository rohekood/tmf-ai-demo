package catalog

import (
	"context"
	"testing"

	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateCatalog_Execute(t *testing.T) {
	mockRepo := new(MockCatalogRepo)
	mockPub := new(MockEventPublisher)
	useCase := NewCreateCatalog(mockRepo, mockPub)

	input := ports.CreateCatalogInput{
		Name:        "Test Catalog",
		Description: "A test catalog",
		ValidFor:    domain.TimePeriod{},
	}

	// Expect Create to be called once
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.Catalog) bool {
		return c.Name == "Test Catalog"
	})).Return(nil)

	mockPub.On("PublishCatalogCreated", mock.Anything, mock.MatchedBy(func(e domain.CatalogCreatedEvent) bool {
		return e.Catalog.Name == "Test Catalog"
	})).Return(nil)

	ctx := context.Background()
	result, err := useCase.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Catalog", result.Name)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
