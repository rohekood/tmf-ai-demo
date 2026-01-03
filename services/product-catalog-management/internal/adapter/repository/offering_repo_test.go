package repository_test

import (
	"context"
	"testing"
	"time"

	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOfferingRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewProductOfferingRepo(db)
	specRepo := repository.NewProductSpecificationRepo(db)
	ctx := context.Background()

	// Need a spec first? GORM AutoMigrate doesn't strictly enforce FK unless defined in tag, but good practice.
	specID := uuid.New().String()
	spec := &domain.ProductSpecification{
		ID:            specID,
		Name:          "Base Spec",
		ProductNumber: "123",
	}
	_ = specRepo.Create(ctx, spec)

	id := uuid.New().String()
	now := time.Now().UTC().Truncate(time.Second)

	offering := &domain.ProductOffering{
		ID:                     id,
		Name:                   "iPhone 13 Offer",
		LifecycleStatus:        "Active",
		ValidFor:               domain.TimePeriod{StartDateTime: &now},
		LastUpdate:             now,
		ProductSpecificationID: &specID,
		ProductOfferingPrice: []domain.ProductOfferingPrice{
			{
				ID:        "price-1",
				PriceType: "recurring",
				Price:     domain.Money{Unit: "USD", Value: 99.99},
			},
		},
	}

	// Create
	assert.NoError(t, repo.Create(ctx, offering))

	// Get
	fetched, err := repo.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "iPhone 13 Offer", fetched.Name)
	assert.Len(t, fetched.ProductOfferingPrice, 1)
	assert.Equal(t, 99.99, fetched.ProductOfferingPrice[0].Price.Value)

	// List
	list, err := repo.List(ctx, map[string]interface{}{"name": "iPhone%"})
	assert.NoError(t, err)
	assert.NotEmpty(t, list)

	// Delete
	assert.NoError(t, repo.Delete(ctx, id))
}
