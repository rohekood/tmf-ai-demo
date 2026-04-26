package repository

import (
	"context"
	"errors"
	"fmt"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"gorm.io/gorm"
)

type ProductOfferingRepo struct {
	db *gorm.DB
}

func NewProductOfferingRepo(db *gorm.DB) ports.ProductOfferingRepository {
	return &ProductOfferingRepo{db: db}
}

func (r *ProductOfferingRepo) Create(ctx context.Context, offering *domain.ProductOffering) error {
	model := FromDomainOffering(offering)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *ProductOfferingRepo) Get(ctx context.Context, id string) (*domain.ProductOffering, error) {
	// fmt.Printf("DEBUG: ProductOfferingRepo.Get called with id='%s'\n", id)
	var model ProductOfferingModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *ProductOfferingRepo) List(ctx context.Context, filters map[string]any) ([]*domain.ProductOffering, error) {
	var models []ProductOfferingModel
	query := r.db.WithContext(ctx)

	if name, ok := filters["name"]; ok {
		query = query.Where("name LIKE ?", name)
	}

	if catID, ok := filters["category"]; ok {
		// Expecting category to be a string ID
		// Construct JSON string for containment: ["id"]
		// Note: category_ids is stored as JSONB array of strings
		// Query: category_ids @> '"id"'
		jsonStr := fmt.Sprintf(`"%s"`, catID)
		query = query.Where("category_ids @> ?", jsonStr)
	}

	if minPrice, ok := filters["min_price"]; ok {
		// product_offering_price is JSONB array of objects
		// Each object has "price" -> "value"
		query = query.Where("EXISTS (SELECT 1 FROM jsonb_array_elements(product_offering_price) elem WHERE (elem->'price'->>'value')::numeric >= ?)", minPrice)
	}

	if maxPrice, ok := filters["max_price"]; ok {
		query = query.Where("EXISTS (SELECT 1 FROM jsonb_array_elements(product_offering_price) elem WHERE (elem->'price'->>'value')::numeric <= ?)", maxPrice)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	offerings := make([]*domain.ProductOffering, len(models))
	for i, m := range models {
		offerings[i] = m.ToDomain()
	}
	return offerings, nil
}

func (r *ProductOfferingRepo) Update(ctx context.Context, offering *domain.ProductOffering) error {
	model := FromDomainOffering(offering)
	result := r.db.WithContext(ctx).Model(&ProductOfferingModel{}).Where("id = ?", offering.ID).Updates(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductOfferingRepo) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&ProductOfferingModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
