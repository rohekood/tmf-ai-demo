package repository

import (
	"context"
	"errors"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"gorm.io/gorm"
)

type CatalogRepo struct {
	db *gorm.DB
}

func NewCatalogRepo(db *gorm.DB) ports.CatalogRepository {
	return &CatalogRepo{db: db}
}

func (r *CatalogRepo) Create(ctx context.Context, catalog *domain.Catalog) error {
	model := FromDomainCatalog(catalog)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *CatalogRepo) Get(ctx context.Context, id string) (*domain.Catalog, error) {
	var model CatalogModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *CatalogRepo) List(ctx context.Context, filters map[string]any) ([]*domain.Catalog, error) {
	var models []CatalogModel
	query := r.db.WithContext(ctx)

	// Apply filters (basic implementation)
	if name, ok := filters["name"]; ok {
		query = query.Where("name LIKE ?", name)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	catalogs := make([]*domain.Catalog, len(models))
	for i, m := range models {
		catalogs[i] = m.ToDomain()
	}
	return catalogs, nil
}

func (r *CatalogRepo) Update(ctx context.Context, catalog *domain.Catalog) error {
	model := FromDomainCatalog(catalog)
	result := r.db.WithContext(ctx).Model(&CatalogModel{}).Where("id = ?", catalog.ID).Updates(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CatalogRepo) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&CatalogModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
