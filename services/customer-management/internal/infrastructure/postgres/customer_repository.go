package postgres

import (
	"context"
	"errors"
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
	tx := GetTx(ctx, r.db)
	if err := tx.Create(c).Error; err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	return nil
}

func (r *CustomerRepository) GetCustomer(ctx context.Context, id string) (*domain.Customer, error) {
	var customer domain.Customer
	// GetTx isn't strictly necessary for reads but good for consistency if reading your own writes within tx
	err := GetTx(ctx, r.db).
		Preload("Accounts").
		Preload("CreditProfiles").
		Preload("ContactMediums").
		Preload("Characteristics").
		Preload("PrivacyConsents").
		Preload("RelatedParties").
		Preload("PaymentMethods").
		Preload("MarketSegments").
		Preload("CustomerInteractions").
		First(&customer, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

func (r *CustomerRepository) UpdateCustomer(ctx context.Context, c *domain.Customer) error {
	tx := GetTx(ctx, r.db)
	result := tx.Save(c)
	if result.Error != nil {
		return fmt.Errorf("failed to update customer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	// Replace sub-resources
	return r.updateSubResources(tx, c.ID, c)
}

func (r *CustomerRepository) PatchCustomer(ctx context.Context, id string, updates map[string]interface{}) error {
	tx := GetTx(ctx, r.db)

	// Handle associations separately

	if privacy, ok := updates["privacy_consents"]; ok {
		if err := tx.Delete(&domain.PrivacyConsent{}, "customer_id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete old privacy consents: %w", err)
		}
		if err := tx.Model(&domain.Customer{ID: id}).Association("PrivacyConsents").Replace(privacy); err != nil {
			return fmt.Errorf("failed to replace privacy consents: %w", err)
		}
		delete(updates, "privacy_consents")
	}
	if accounts, ok := updates["accounts"]; ok {
		if err := tx.Delete(&domain.CustomerAccount{}, "customer_id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete old accounts: %w", err)
		}
		if err := tx.Model(&domain.Customer{ID: id}).Association("Accounts").Replace(accounts); err != nil {
			return fmt.Errorf("failed to replace accounts: %w", err)
		}
		delete(updates, "accounts")
	}
	if creditProfiles, ok := updates["credit_profiles"]; ok {
		if err := tx.Delete(&domain.CreditProfile{}, "customer_id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete old credit profiles: %w", err)
		}
		if err := tx.Model(&domain.Customer{ID: id}).Association("CreditProfiles").Replace(creditProfiles); err != nil {
			return fmt.Errorf("failed to replace credit profiles: %w", err)
		}
		delete(updates, "credit_profiles")
	}
	if related, ok := updates["related_parties"]; ok {
		if err := tx.Delete(&domain.RelatedParty{}, "customer_id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete old related parties: %w", err)
		}
		if err := tx.Model(&domain.Customer{ID: id}).Association("RelatedParties").Replace(related); err != nil {
			return fmt.Errorf("failed to replace related parties: %w", err)
		}
		delete(updates, "related_parties")
	}
	if payments, ok := updates["payment_methods"]; ok {
		if err := tx.Delete(&domain.PaymentMethod{}, "customer_id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete old payment methods: %w", err)
		}
		if err := tx.Model(&domain.Customer{ID: id}).Association("PaymentMethods").Replace(payments); err != nil {
			return fmt.Errorf("failed to replace payment methods: %w", err)
		}
		delete(updates, "payment_methods")
	}
	if segments, ok := updates["market_segments"]; ok {
		if err := tx.Delete(&domain.MarketSegment{}, "customer_id = ?", id).Error; err != nil {
			return fmt.Errorf("failed to delete old market segments: %w", err)
		}
		if err := tx.Model(&domain.Customer{ID: id}).Association("MarketSegments").Replace(segments); err != nil {
			return fmt.Errorf("failed to replace market segments: %w", err)
		}
		delete(updates, "market_segments")
	}

	// Update scalar fields if any remain
	if len(updates) > 0 {
		result := tx.Model(&domain.Customer{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("failed to patch customer: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
	}
	return nil
}

func (r *CustomerRepository) DeleteCustomer(ctx context.Context, id string) error {
	tx := GetTx(ctx, r.db)
	result := tx.Delete(&domain.Customer{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete customer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CustomerRepository) SearchCustomers(ctx context.Context, criteria map[string]interface{}) ([]domain.Customer, error) {
	var customers []domain.Customer
	query := r.db.WithContext(ctx).Preload("Accounts")

	for key, value := range criteria {
		switch key {
		case "id":
			query = query.Where("id = ?", value)
		case "search":
			// Generic search across multiple fields
			searchTerm := "%" + value.(string) + "%"
			query = query.Where("(name ILIKE ? OR status ILIKE ? OR party_id = ? OR party_type ILIKE ?)",
				searchTerm, searchTerm, value, searchTerm)
		case "name":
			query = query.Where("name ILIKE ?", "%"+value.(string)+"%")
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

func (r *CustomerRepository) AddInteraction(ctx context.Context, interaction *domain.CustomerInteraction) error {
	tx := GetTx(ctx, r.db)
	if err := tx.Create(interaction).Error; err != nil {
		return fmt.Errorf("failed to add interaction: %w", err)
	}
	return nil
}

// Helpers

func (r *CustomerRepository) updateSubResources(tx *gorm.DB, customerID string, c *domain.Customer) error {
	// 1. Accounts
	if err := tx.Delete(&domain.CustomerAccount{}, "customer_id = ?", customerID).Error; err != nil {
		return err
	}
	if len(c.Accounts) > 0 {
		for i := range c.Accounts {
			c.Accounts[i].CustomerID = customerID
		}
		if err := tx.Create(&c.Accounts).Error; err != nil {
			return err
		}
	}

	// 2. CreditProfiles
	if err := tx.Delete(&domain.CreditProfile{}, "customer_id = ?", customerID).Error; err != nil {
		return err
	}
	if len(c.CreditProfiles) > 0 {
		for i := range c.CreditProfiles {
			c.CreditProfiles[i].CustomerID = customerID
		}
		if err := tx.Create(&c.CreditProfiles).Error; err != nil {
			return err
		}
	}

	// 3. ContactMediums
	if err := tx.Delete(&domain.ContactMedium{}, "customer_id = ?", customerID).Error; err != nil {
		return err
	}
	if len(c.ContactMediums) > 0 {
		for i := range c.ContactMediums {
			c.ContactMediums[i].CustomerID = customerID
		}
		if err := tx.Create(&c.ContactMediums).Error; err != nil {
			return err
		}
	}

	// 4. Characteristics
	if err := tx.Delete(&domain.CustomerCharacteristic{}, "customer_id = ?", customerID).Error; err != nil {
		return err
	}
	if len(c.Characteristics) > 0 {
		for i := range c.Characteristics {
			c.Characteristics[i].CustomerID = customerID
		}
		if err := tx.Create(&c.Characteristics).Error; err != nil {
			return err
		}
	}

	// 6. PrivacyConsents
	if err := tx.Delete(&domain.PrivacyConsent{}, "customer_id = ?", customerID).Error; err != nil {
		return err
	}
	if len(c.PrivacyConsents) > 0 {
		for i := range c.PrivacyConsents {
			c.PrivacyConsents[i].CustomerID = customerID
		}
		if err := tx.Create(&c.PrivacyConsents).Error; err != nil {
			return err
		}
	}

	// 7. RelatedParties
	if err := tx.Delete(&domain.RelatedParty{}, "customer_id = ?", customerID).Error; err != nil {
		return err
	}
	if len(c.RelatedParties) > 0 {
		for i := range c.RelatedParties {
			c.RelatedParties[i].CustomerID = customerID
		}
		if err := tx.Create(&c.RelatedParties).Error; err != nil {
			return err
		}
	}

	// 8. PaymentMethods
	if err := tx.Delete(&domain.PaymentMethod{}, "customer_id = ?", customerID).Error; err != nil {
		return err
	}
	if len(c.PaymentMethods) > 0 {
		for i := range c.PaymentMethods {
			c.PaymentMethods[i].CustomerID = customerID
		}
		if err := tx.Create(&c.PaymentMethods).Error; err != nil {
			return err
		}
	}

	// 9. MarketSegments
	if err := tx.Delete(&domain.MarketSegment{}, "customer_id = ?", customerID).Error; err != nil {
		return err
	}
	if len(c.MarketSegments) > 0 {
		for i := range c.MarketSegments {
			c.MarketSegments[i].CustomerID = customerID
		}
		if err := tx.Create(&c.MarketSegments).Error; err != nil {
			return err
		}
	}

	return nil
}
