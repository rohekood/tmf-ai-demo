package postgres

import (
	"context"
	"fmt"
	"tmf/services/customer-management/internal/domain"

	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) CreateCustomer(ctx context.Context, c *domain.Customer) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CustomerRepository) GetCustomer(ctx context.Context, id string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).
		Preload("Accounts").
		Preload("CreditProfiles").
		Preload("ContactMediums").
		First(&customer, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update main customer record
		if err := tx.Save(c).Error; err != nil {
			return err
		}

		// Update associated entities - standard GORM Save handles this if using full objects,
		// but explicit management is often safer for complex relations.
		// For now we assume the full object is passed and GORM's full save logic applies.
		// If we need more granular control (like preventing deletion of unrelated accounts),
		// we'd add it here.
		return nil
	})
}

func (r *CustomerRepository) PatchCustomer(ctx context.Context, id string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&domain.Customer{}).Where("id = ?", id).Updates(updates).Error
}

func (r *CustomerRepository) DeleteCustomer(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Customer{}, "id = ?", id).Error
}

func (r *CustomerRepository) SearchCustomers(ctx context.Context, criteria map[string]interface{}) ([]domain.Customer, error) {
	var customers []domain.Customer
	query := r.db.WithContext(ctx)

	for key, value := range criteria {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}

	err := query.Find(&customers).Error
	return customers, err
}
