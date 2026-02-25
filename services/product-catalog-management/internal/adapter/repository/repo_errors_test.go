package repository_test

import (
	"context"
	"testing"
	"time"

	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestRepo_NotFound_And_Duplicates(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	
	catRepo := repository.NewCatalogRepo(db)
	cgRepo := repository.NewCategoryRepo(db)
	spRepo := repository.NewProductSpecificationRepo(db)
	offRepo := repository.NewProductOfferingRepo(db)

	// 1. NotFound
	_, err := catRepo.Get(ctx, "non-existent")
	assert.Equal(t, domain.ErrNotFound, err)
	err = catRepo.Update(ctx, &domain.Catalog{ID: "non-existent"})
	assert.Equal(t, domain.ErrNotFound, err)
	err = catRepo.Delete(ctx, "non-existent")
	assert.Equal(t, domain.ErrNotFound, err)

	_, err = cgRepo.Get(ctx, "non-existent")
	assert.Equal(t, domain.ErrNotFound, err)
	err = cgRepo.Update(ctx, &domain.Category{ID: "non-existent"})
	assert.Equal(t, domain.ErrNotFound, err)
	err = cgRepo.Delete(ctx, "non-existent")
	assert.Equal(t, domain.ErrNotFound, err)

	_, err = spRepo.Get(ctx, "non-existent")
	assert.Equal(t, domain.ErrNotFound, err)
	err = spRepo.Update(ctx, &domain.ProductSpecification{ID: "non-existent"})
	assert.Equal(t, domain.ErrNotFound, err)
	err = spRepo.Delete(ctx, "non-existent")
	assert.Equal(t, domain.ErrNotFound, err)

	_, err = offRepo.Get(ctx, "non-existent")
	assert.Equal(t, domain.ErrNotFound, err)
	err = offRepo.Update(ctx, &domain.ProductOffering{ID: "non-existent"})
	assert.Equal(t, domain.ErrNotFound, err)
	err = offRepo.Delete(ctx, "non-existent")
	assert.Equal(t, domain.ErrNotFound, err)

	// 2. Duplicates
	cat := &domain.Catalog{ID: "dup-1", Name: "Dup"}
	err = catRepo.Create(ctx, cat)
	assert.NoError(t, err)
	err = catRepo.Create(ctx, cat)
	assert.NotNil(t, err)

	cg := &domain.Category{ID: "dup-2", Name: "Dup"}
	err = cgRepo.Create(ctx, cg)
	assert.NoError(t, err)
	err = cgRepo.Create(ctx, cg)
	assert.NotNil(t, err)

	sp := &domain.ProductSpecification{ID: "dup-3", Name: "Dup"}
	err = spRepo.Create(ctx, sp)
	assert.NoError(t, err)
	err = spRepo.Create(ctx, sp)
	assert.NotNil(t, err)
	
	// test spRepo List
	spList, err := spRepo.List(ctx, map[string]interface{}{"name": "Dup"})
	assert.NoError(t, err)
	assert.Len(t, spList, 1)

	now := time.Now()
	off := &domain.ProductOffering{ID: "dup-4", Name: "Dup", LastUpdate: now, ValidFor: domain.TimePeriod{StartDateTime: &now}}
	err = offRepo.Create(ctx, off)
	assert.NoError(t, err)
	err = offRepo.Create(ctx, off)
	assert.NotNil(t, err)
	
	// test offRepo List and Update
	off.Name = "Dup Updated"
	err = offRepo.Update(ctx, off)
	assert.NoError(t, err)
	offList, err := offRepo.List(ctx, map[string]interface{}{"name": "Dup%"})
	assert.NoError(t, err)
	assert.Len(t, offList, 1)
}
