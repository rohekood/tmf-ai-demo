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

func TestUpdateOffering_Execute(t *testing.T) {
	mockRepo := new(MockOfferingRepo)
	mockSpecRepo := new(MockSpecRepo)
	mockPub := new(MockEventPublisher)

	useCase := NewUpdateProductOfferingUseCase(mockRepo, mockSpecRepo, mockPub, &repository.NoOpTransactionManager{})

	offeringID := "off-123"
	specID := "spec-123"

	tests := []struct {
		name        string
		input       ports.UpdateProductOfferingInput
		setupMocks  func()
		expectedErr error
	}{
		{
			name: "Success: Draft to Active",
			input: ports.UpdateProductOfferingInput{
				ID:              offeringID,
				LifecycleStatus: strPtr(domain.LifecycleStatusActive),
			},
			setupMocks: func() {
				mockRepo.On("Get", mock.Anything, offeringID).Return(&domain.ProductOffering{
					ID:                     offeringID,
					LifecycleStatus:        domain.LifecycleStatusDraft,
					ProductSpecificationID: &specID,
					Name:                   "Test Offering",
				}, nil).Once()

				// Spec Check
				mockSpecRepo.On("Get", mock.Anything, specID).Return(&domain.ProductSpecification{
					ID:              specID,
					LifecycleStatus: domain.SpecLifecycleStatusActive,
				}, nil).Once()

				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(o *domain.ProductOffering) bool {
					return o.LifecycleStatus == domain.LifecycleStatusActive
				})).Return(nil).Once()

				mockPub.On("PublishProductOfferingUpdated", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			name: "Failure: Invalid Transition (Retired -> Active)",
			input: ports.UpdateProductOfferingInput{
				ID:              offeringID,
				LifecycleStatus: strPtr(domain.LifecycleStatusActive),
			},
			setupMocks: func() {
				mockRepo.On("Get", mock.Anything, offeringID).Return(&domain.ProductOffering{
					ID:              offeringID,
					LifecycleStatus: domain.LifecycleStatusRetired,
					Name:            "Test Offering",
				}, nil).Once()
			},
			expectedErr: domain.ErrInvalidInput,
		},
		{
			name: "Failure: Spec is Retired on Activation",
			input: ports.UpdateProductOfferingInput{
				ID:              offeringID,
				LifecycleStatus: strPtr(domain.LifecycleStatusActive),
			},
			setupMocks: func() {
				mockRepo.On("Get", mock.Anything, offeringID).Return(&domain.ProductOffering{
					ID:                     offeringID,
					LifecycleStatus:        domain.LifecycleStatusDraft,
					ProductSpecificationID: &specID,
					Name:                   "Test Offering",
				}, nil).Once()

				mockSpecRepo.On("Get", mock.Anything, specID).Return(&domain.ProductSpecification{
					ID:              specID,
					LifecycleStatus: domain.SpecLifecycleStatusRetired,
				}, nil).Once()
			},
			expectedErr: domain.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			ctx := context.Background()
			_, err := useCase.Execute(ctx, tt.input)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
