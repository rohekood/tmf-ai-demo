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

func TestSpecificationRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewProductSpecificationRepo(db)
	ctx := context.Background()

	id := uuid.New().String()
	now := time.Now().UTC().Truncate(time.Second)

	spec := &domain.ProductSpecification{
		ID:              id,
		Name:            "iPhone 13 Spec",
		ProductNumber:   "PROD-001",
		LifecycleStatus: "Active",
		ValidFor:        domain.TimePeriod{StartDateTime: &now},
		LastUpdate:      now,
		Characteristics: map[string]domain.ProductSpecCharacteristic{
			"color": {
				Name:         "color",
				ValueType:    "string",
				Configurable: true,
				ValidValues:  []string{"Red", "Blue"},
			},
		},
	}

	// Create
	assert.NoError(t, repo.Create(ctx, spec))

	// Get
	fetched, err := repo.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "iPhone 13 Spec", fetched.Name)
	assert.Equal(t, "Red", fetched.Characteristics["color"].ValidValues[0])

	// Update
	spec.Description = "Updated Desc"
	assert.NoError(t, repo.Update(ctx, spec))

	// Delete
	assert.NoError(t, repo.Delete(ctx, id))
}
