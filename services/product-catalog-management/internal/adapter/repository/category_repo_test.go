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

func TestCategoryRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCategoryRepo(db)
	ctx := context.Background()

	id := uuid.New().String()
	now := time.Now().UTC().Truncate(time.Second)

	cat := &domain.Category{
		ID:              id,
		Name:            "Smartphones",
		Description:     "All smartphones",
		IsRoot:          true,
		ValidFor:        domain.TimePeriod{StartDateTime: &now},
		LastUpdate:      now,
		LifecycleStatus: "Active",
	}

	// Create
	assert.NoError(t, repo.Create(ctx, cat))

	// Get
	fetched, err := repo.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "Smartphones", fetched.Name)

	// Update
	cat.Name = "Mobile Phones"
	assert.NoError(t, repo.Update(ctx, cat))

	fetched2, err := repo.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "Mobile Phones", fetched2.Name)

	// List
	list, err := repo.List(ctx, map[string]any{"name": "Mobile%"})
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	// Sub-category
	subID := uuid.New().String()
	subCat := &domain.Category{
		ID:              subID,
		Name:            "Apple",
		ParentID:        &id,
		IsRoot:          false,
		ValidFor:        domain.TimePeriod{StartDateTime: &now},
		LastUpdate:      now,
		LifecycleStatus: "Active",
	}
	assert.NoError(t, repo.Create(ctx, subCat))

	// Verify Parent Rel
	fetchedSub, err := repo.Get(ctx, subID)
	assert.NoError(t, err)
	assert.Equal(t, id, *fetchedSub.ParentID)

	// Delete
	assert.NoError(t, repo.Delete(ctx, id))
	// Depending on DB constraints (CASCADE), subCat might be deleted or orphaned.
	// Since we use GORM AutoMigrate without strict foreign keys by default unless specified, it might remain or fail.
	// Our setupDB truncates anyway.
}
