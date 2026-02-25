package offering

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

func TestGetOffering_Enrich(t *testing.T) {
	repo := new(MockOfferingRepo)
	specRepo := new(MockSpecRepo)
	catRepo := new(MockCategoryRepoLocal)

	uc := NewGetProductOffering(repo, specRepo, catRepo)
	ctx := context.Background()

	specID := "spec1"
	catID := "cat1"

	off := &domain.ProductOffering{
		ID:                     "1",
		ProductSpecificationID: &specID,
		CategoryIDs:            []string{catID},
	}

	repo.On("Get", ctx, "1").Return(off, nil)
	specRepo.On("Get", ctx, specID).Return(&domain.ProductSpecification{ID: specID, Name: "Spec"}, nil)
	catRepo.On("Get", ctx, catID).Return(&domain.Category{ID: catID, Name: "Cat"}, nil)

	res, err := uc.Execute(ctx, ports.GetProductOfferingInput{ID: "1", Enrich: true})
	assert.NoError(t, err)
	assert.NotNil(t, res.ProductSpecification)
	assert.Equal(t, "Spec", res.ProductSpecification.Name)
	assert.Len(t, res.Categories, 1)
	assert.Equal(t, "Cat", res.Categories[0].Name)
}
