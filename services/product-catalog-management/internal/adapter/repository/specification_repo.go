package repository

import (
	"context"
	"errors"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"gorm.io/gorm"
)

type ProductSpecificationRepo struct {
	db *gorm.DB
}

func NewProductSpecificationRepo(db *gorm.DB) ports.ProductSpecificationRepository {
	return &ProductSpecificationRepo{db: db}
}

func (r *ProductSpecificationRepo) Create(ctx context.Context, spec *domain.ProductSpecification) error {
	model := FromDomainSpecification(spec)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *ProductSpecificationRepo) Get(ctx context.Context, id string) (*domain.ProductSpecification, error) {
	var model ProductSpecificationModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *ProductSpecificationRepo) List(ctx context.Context, filters map[string]interface{}) ([]*domain.ProductSpecification, error) {
	var models []ProductSpecificationModel
	query := r.db.WithContext(ctx)

	if name, ok := filters["name"]; ok {
		query = query.Where("name LIKE ?", name)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	specs := make([]*domain.ProductSpecification, len(models))
	for i, m := range models {
		specs[i] = m.ToDomain()
	}
	return specs, nil
}

func (r *ProductSpecificationRepo) Update(ctx context.Context, spec *domain.ProductSpecification) error {
	model := FromDomainSpecification(spec)
	result := r.db.WithContext(ctx).Model(&ProductSpecificationModel{}).Where("id = ?", spec.ID).Updates(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductSpecificationRepo) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&ProductSpecificationModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
