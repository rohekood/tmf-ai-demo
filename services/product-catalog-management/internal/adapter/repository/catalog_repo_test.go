package repository_test

import (
	"context"
	"testing"
	"time"

	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	if sharedDB == nil {
		t.Fatal("Shared DB not initialized. Ensure TestMain is running.")
	}

	tables := []string{
		"product_offerings", // link FK
		"categories",        // link FK
		"product_specifications",
		"catalogs",
	}

	for _, table := range tables {
		if err := sharedDB.Exec("TRUNCATE TABLE " + table + " CASCADE").Error; err != nil {
			t.Logf("Failed to truncate table %s: %v", table, err)
		}
	}

	return sharedDB
}

func TestCatalogRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCatalogRepo(db)
	ctx := context.Background()

	// 1. Create
	id := uuid.New().String()
	start := time.Now().UTC().Truncate(time.Second) // Postgres doesn't save monotonic clock
	catalog := &domain.Catalog{
		ID:          id,
		Name:        "Integration Catalog",
		Description: "Test Desc",
		ValidFor: domain.TimePeriod{
			StartDateTime: &start,
		},
		LastUpdate:      start,
		LifecycleStatus: "Active",
	}

	err := repo.Create(ctx, catalog)
	assert.NoError(t, err)

	// 2. Get
	fetched, err := repo.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, catalog.Name, fetched.Name)
	assert.Equal(t, catalog.ID, fetched.ID)
	// Time comparison can be tricky with timezones, usually truncated to micro/millis
	// assert.True(t, catalog.ValidFor.StartDateTime.Equal(*fetched.ValidFor.StartDateTime))

	// 3. Update
	fetched.Name = "Updated Catalog"
	err = repo.Update(ctx, fetched)
	assert.NoError(t, err)

	fetched2, err := repo.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Catalog", fetched2.Name)

	// 4. List
	list, err := repo.List(ctx, map[string]any{"name": "Updated%"})
	assert.NoError(t, err)
	assert.NotEmpty(t, list)
	assert.Equal(t, id, list[0].ID)

	// 5. Delete
	err = repo.Delete(ctx, id)
	assert.NoError(t, err)

	_, err = repo.Get(ctx, id)
	assert.Equal(t, domain.ErrNotFound, err)
}
