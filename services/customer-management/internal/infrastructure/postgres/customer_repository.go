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
	return r.withUser(ctx, func(tx *gorm.DB) error {
		return tx.Create(c).Error
	})
}

func (r *CustomerRepository) GetCustomer(ctx context.Context, id string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).
		Preload("Accounts").
		Preload("CreditProfiles").
		Preload("ContactMediums").
		Preload("Characteristics").
		First(&customer, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	return r.withUser(ctx, func(tx *gorm.DB) error {
		return tx.Save(c).Error
	})
}

func (r *CustomerRepository) PatchCustomer(ctx context.Context, id string, updates map[string]interface{}) error {
	return r.withUser(ctx, func(tx *gorm.DB) error {
		return tx.Model(&domain.Customer{}).Where("id = ?", id).Updates(updates).Error
	})
}

func (r *CustomerRepository) DeleteCustomer(ctx context.Context, id string) error {
	return r.withUser(ctx, func(tx *gorm.DB) error {
		return tx.Delete(&domain.Customer{}, "id = ?", id).Error
	})
}

func (r *CustomerRepository) SearchCustomers(ctx context.Context, criteria map[string]interface{}) ([]domain.Customer, error) {
	var customers []domain.Customer
	query := r.db.WithContext(ctx)

	for key, value := range criteria {
		switch key {
		case "id":
			query = query.Where("id = ?", value)
		case "name":
			query = query.Where("name = ?", value)
		case "status":
			query = query.Where("status = ?", value)
		case "party_id":
			query = query.Where("party_id = ?", value)
		default:
			return nil, fmt.Errorf("invalid search criteria: %s", key)
		}
	}

	err := query.Find(&customers).Error
	return customers, err
}

// Helpers

func (r *CustomerRepository) withUser(ctx context.Context, fn func(tx *gorm.DB) error) error {
	userID, _ := ctx.Value(domain.UserContextKey).(string)
	if userID == "" {
		userID = "unknown"
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Set local session variable for audit trigger
		if err := tx.Exec("SELECT set_config('app.current_user', ?, true)", userID).Error; err != nil {
			return err
		}
		return fn(tx)
	})
}
