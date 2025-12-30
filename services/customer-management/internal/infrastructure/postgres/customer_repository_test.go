package postgres

import (
	"context"
	"testing"
	"tmf/services/customer-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, string) {
	if sharedDB == nil {
		t.Fatal("Shared DB not initialized. Ensure TestMain is running.")
	}

	// Truncate tables to ensure clean state
	// Order matters due to FKs
	tables := []string{
		"applied_billing_rates",
		"customer_interactions",
		"market_segments",
		"payment_methods",
		"related_parties",
		"privacy_consents",
		"tax_exemptions",
		"customer_characteristics",
		"contact_mediums",
		"credit_profiles",
		"customer_accounts",
		"customers",
		"audit.logged_actions",
	}

	for _, table := range tables {
		// Use Unscoped to ensure even soft-deleted records are gone if we were using Delete,
		// but for Truncate it's explicit.
		// Using EXEC for Truncate
		if err := sharedDB.Exec("TRUNCATE TABLE " + table + " CASCADE").Error; err != nil {
			// Some tables might not exist yet if migrations changed, but they should.
			// Or if specific tables are empty it's fine.
			// Warning: CASCADE might be needed.
			t.Logf("Failed to truncate table %s: %v", table, err)
		}
	}

	return sharedDB, sharedConnStr
}

func TestCreateCustomer(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust := &domain.Customer{
		ID:      "cust-1",
		Name:    "Test Customer",
		Status:  domain.CustomerStatusActive,
		PartyID: "party-1",
	}

	err := repo.CreateCustomer(ctx, cust)
	assert.NoError(t, err)

	saved, err := repo.GetCustomer(ctx, "cust-1")
	assert.NoError(t, err)
	assert.Equal(t, cust.ID, saved.ID)
	assert.Equal(t, cust.Name, saved.Name)
}

func TestUpdateCustomer(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust := &domain.Customer{
		ID:      "cust-upd-1",
		Name:    "Original Name",
		Status:  domain.CustomerStatusActive,
		PartyID: "party-upd-1",
	}
	require.NoError(t, repo.CreateCustomer(ctx, cust))

	cust.Name = "Updated Name"
	cust.Status = domain.CustomerStatusSuspended
	err := repo.UpdateCustomer(ctx, cust)
	assert.NoError(t, err)

	updated, err := repo.GetCustomer(ctx, "cust-upd-1")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, domain.CustomerStatusSuspended, updated.Status)
}

func TestPatchCustomer(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust := &domain.Customer{
		ID:      "cust-patch-1",
		Name:    "Patch Me",
		Status:  domain.CustomerStatusActive,
		PartyID: "party-patch-1",
	}
	require.NoError(t, repo.CreateCustomer(ctx, cust))

	updates := map[string]interface{}{
		"name":   "Patched Name",
		"status": domain.CustomerStatusClosed,
		"tax_exemptions": []domain.TaxExemption{
			{
				ID:                "tax-1",
				CertificateNumber: "CERT-123",
			},
		},
		"privacy_consents": []domain.PrivacyConsent{
			{
				ID:          "privacy-1",
				ConsentType: "Marketing",
				Status:      "Given",
			},
		},
	}
	err := repo.PatchCustomer(ctx, "cust-patch-1", updates)
	assert.NoError(t, err)

	patched, err := repo.GetCustomer(ctx, "cust-patch-1")
	assert.NoError(t, err)
	assert.Equal(t, "Patched Name", patched.Name)
	assert.Equal(t, domain.CustomerStatusClosed, patched.Status)
	assert.Len(t, patched.TaxExemptions, 1)
	assert.Equal(t, "CERT-123", patched.TaxExemptions[0].CertificateNumber)
	assert.Len(t, patched.PrivacyConsents, 1)
	assert.Equal(t, "Marketing", patched.PrivacyConsents[0].ConsentType)
}

func TestDeleteCustomer(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust := &domain.Customer{
		ID:      "cust-del-1",
		Name:    "Delete Me",
		Status:  domain.CustomerStatusActive,
		PartyID: "party-del-1",
	}
	require.NoError(t, repo.CreateCustomer(ctx, cust))

	err := repo.DeleteCustomer(ctx, "cust-del-1")
	assert.NoError(t, err)

	_, err = repo.GetCustomer(ctx, "cust-del-1")
	assert.Error(t, err)
}

func TestSearchCustomers(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust1 := &domain.Customer{ID: "s-1", Name: "Alice", Status: domain.CustomerStatusActive, PartyID: "p-1"}
	cust2 := &domain.Customer{ID: "s-2", Name: "Bob", Status: domain.CustomerStatusActive, PartyID: "p-2"}
	require.NoError(t, repo.CreateCustomer(ctx, cust1))
	require.NoError(t, repo.CreateCustomer(ctx, cust2))

	// Search by name
	results, err := repo.SearchCustomers(ctx, map[string]interface{}{"name": "Alice"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "s-1", results[0].ID)

	// Search by status
	results, err = repo.SearchCustomers(ctx, map[string]interface{}{"status": domain.CustomerStatusActive})
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Search by party_id
	results, err = repo.SearchCustomers(ctx, map[string]interface{}{"party_id": "p-2"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "s-2", results[0].ID)
}

func TestAuditTrail(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)

	userID := "test-audit-user"
	ctx := context.WithValue(context.Background(), domain.UserContextKey, userID)

	cust := &domain.Customer{
		ID:     "audit-1",
		Name:   "Audit Test",
		Status: domain.CustomerStatusActive,
	}

	err := repo.CreateCustomer(ctx, cust)
	assert.NoError(t, err)

	// Verify Audit Log
	type LoggedAction struct {
		TableName string `gorm:"column:table_name"`
		UserName  string `gorm:"column:user_name"`
		Action    string `gorm:"column:action"`
	}

	var auditLog LoggedAction
	err = db.Table("audit.logged_actions").
		Where("table_name = ? AND action = ?", "customers", "I").
		Order("action_tstamp_clk DESC").
		First(&auditLog).Error

	require.NoError(t, err)
	assert.Equal(t, userID, auditLog.UserName)
	assert.Equal(t, "customers", auditLog.TableName)
}
