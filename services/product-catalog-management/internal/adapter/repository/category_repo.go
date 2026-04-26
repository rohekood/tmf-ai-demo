package repository

import (
	"context"
	"errors"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"gorm.io/gorm"
)

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) ports.CategoryRepository {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	model := FromDomainCategory(category)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *CategoryRepo) Get(ctx context.Context, id string) (*domain.Category, error) {
	var model CategoryModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *CategoryRepo) List(ctx context.Context, filters map[string]any) ([]*domain.Category, error) {
	var models []CategoryModel
	query := r.db.WithContext(ctx)

	if name, ok := filters["name"]; ok {
		query = query.Where("name LIKE ?", name)
	}
	if catalogID, ok := filters["catalogId"]; ok {
		query = query.Where("catalog_id = ?", catalogID)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	categories := make([]*domain.Category, len(models))
	for i, m := range models {
		categories[i] = m.ToDomain()
	}
	return categories, nil
}

func (r *CategoryRepo) Update(ctx context.Context, category *domain.Category) error {
	model := FromDomainCategory(category)
	result := r.db.WithContext(ctx).Model(&CategoryModel{}).Where("id = ?", category.ID).Updates(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&CategoryModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
