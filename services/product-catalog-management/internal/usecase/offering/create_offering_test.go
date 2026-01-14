package offering

import (
	"context"
	"testing"
	"time"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOffering_Execute(t *testing.T) {
	mockRepo := new(MockOfferingRepo)
	mockSpecRepo := new(MockSpecRepo)
	mockPub := new(MockEventPublisher)

	useCase := NewCreateProductOffering(mockRepo, mockSpecRepo, mockPub, &repository.NoOpTransactionManager{})

	specID := "spec-123"
	input := ports.CreateProductOfferingInput{
		Name:            "Internet Plan",
		Description:     "High speed internet",
		IsBundle:        false,
		IsSellable:      true,
		LifecycleStatus: "Active",
		ProductSpecID:   &specID,
		Prices: []domain.ProductOfferingPrice{
			{PriceType: "recurring", Price: domain.Money{Value: 50.0, Unit: "USD"}},
		},
		ValidFor: domain.TimePeriod{StartDateTime: func() *time.Time { t := time.Now(); return &t }()},
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *domain.ProductOffering) bool {
		return o.Name == "Internet Plan"
	})).Return(nil)

	// Mock Spec Repo call
	mockSpecRepo.On("Get", mock.Anything, "spec-123").Return(&domain.ProductSpecification{
		ID:              "spec-123",
		LifecycleStatus: domain.SpecLifecycleStatusActive,
	}, nil)

	mockPub.On("PublishProductOfferingCreated", mock.Anything, mock.MatchedBy(func(e domain.ProductOfferingCreatedEvent) bool {
		return e.ProductOffering.Name == "Internet Plan"
	})).Return(nil)

	ctx := context.Background()
	result, err := useCase.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Internet Plan", result.Name)

	mockRepo.AssertExpectations(t)
	mockSpecRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
