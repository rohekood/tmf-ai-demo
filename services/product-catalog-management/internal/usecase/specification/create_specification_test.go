package specification

import (
	"context"
	"testing"
	"time"

	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSpecification_Execute(t *testing.T) {
	mockRepo := new(MockSpecificationRepo)
	mockPub := new(MockEventPublisher)

	useCase := NewCreateProductSpecification(mockRepo, mockPub)

	input := ports.CreateProductSpecificationInput{
		Name:            "iPhone 13 Spec",
		ProductNumber:   "TEST-001",
		Description:     "A test specification",
		LifecycleStatus: "Active",
		ValidFor:        domain.TimePeriod{StartDateTime: func() *time.Time { t := time.Now(); return &t }()},
		Characteristics: map[string]domain.ProductSpecCharacteristic{
			"color": {Name: "color", ValueType: "string", ValidValues: []string{"Red", "Blue"}},
		},
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.ProductSpecification) bool {
		return s.Name == "iPhone 13 Spec" && s.ProductNumber == "TEST-001"
	})).Return(nil)

	mockPub.On("PublishProductSpecificationCreated", mock.Anything, mock.MatchedBy(func(e domain.ProductSpecificationCreatedEvent) bool {
		return e.ProductSpecification.Name == "iPhone 13 Spec"
	})).Return(nil)

	ctx := context.Background()
	result, err := useCase.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "iPhone 13 Spec", result.Name)
	assert.Equal(t, "TEST-001", result.ProductNumber)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
