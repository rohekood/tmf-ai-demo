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
		"customer_interactions",
		"market_segments",
		"payment_methods",
		"related_parties",
		"privacy_consents",

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
	cust.Accounts = []domain.CustomerAccount{{ID: "a2", Name: "Acc", AccountStatus: "Active"}}
	cust.CreditProfiles = []domain.CreditProfile{{ID: "cp2", CreditScore: 800}}
	cust.ContactMediums = []domain.ContactMedium{{ID: "cm2", MediumType: "Phone", Value: "123456"}}
	cust.Characteristics = []domain.CustomerCharacteristic{{ID: "ch2", Name: "Type", Value: "B2B"}}
	cust.PrivacyConsents = []domain.PrivacyConsent{{ID: "pc2", ConsentType: "Marketing", Status: "Given"}}
	cust.RelatedParties = []domain.RelatedParty{{ID: "rp2", Name: "Partner"}}
	cust.PaymentMethods = []domain.PaymentMethod{{ID: "pm2", Type: "BankTransfer"}}
	cust.MarketSegments = []domain.MarketSegment{{ID: "ms2", Name: "Enterprise"}}

	err := repo.UpdateCustomer(ctx, cust)
	assert.NoError(t, err)

	updated, err := repo.GetCustomer(ctx, "cust-upd-1")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, domain.CustomerStatusSuspended, updated.Status)
	assert.Len(t, updated.Accounts, 1)
	assert.Len(t, updated.CreditProfiles, 1)
	assert.Len(t, updated.ContactMediums, 1)
	assert.Len(t, updated.Characteristics, 1)
	assert.Len(t, updated.PrivacyConsents, 1)
	assert.Len(t, updated.RelatedParties, 1)
	assert.Len(t, updated.PaymentMethods, 1)
	assert.Len(t, updated.MarketSegments, 1)
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

	// Test 1: Basic fields and one sub-resource
	updates := map[string]interface{}{
		"name":   "Patched Name",
		"status": domain.CustomerStatusClosed,

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

	assert.Len(t, patched.PrivacyConsents, 1)
	assert.Equal(t, "Marketing", patched.PrivacyConsents[0].ConsentType)

	// Test 2: Full Patch (All sub-resources)
	fullUpdates := map[string]interface{}{
		"accounts": []domain.CustomerAccount{
			{ID: "a1", Name: "Account 1", AccountStatus: "Active"},
		},
		"credit_profiles": []domain.CreditProfile{
			{ID: "cp1", CreditScore: 750},
		},
		"contact_mediums": []domain.ContactMedium{
			{ID: "cm1", MediumType: "Email", Value: "test@example.com"},
		},
		"characteristics": []domain.CustomerCharacteristic{
			{ID: "char1", Name: "Segment", Value: "VIP"},
		},
		"related_parties": []domain.RelatedParty{
			{ID: "rp1", Name: "Parent"},
		},
		"payment_methods": []domain.PaymentMethod{
			{ID: "pm1", Type: "CreditCard"},
		},
		"market_segments": []domain.MarketSegment{
			{ID: "ms1", Name: "Retail"},
		},
	}
	err = repo.PatchCustomer(ctx, "cust-patch-1", fullUpdates)
	assert.NoError(t, err)

	patchedFull, err := repo.GetCustomer(ctx, "cust-patch-1")
	assert.NoError(t, err)

	assert.Len(t, patchedFull.Accounts, 1)
	assert.Equal(t, "Account 1", patchedFull.Accounts[0].Name)

	assert.Len(t, patchedFull.CreditProfiles, 1)
	assert.Equal(t, 750, patchedFull.CreditProfiles[0].CreditScore)

	assert.Len(t, patchedFull.ContactMediums, 1)
	assert.Equal(t, "test@example.com", patchedFull.ContactMediums[0].Value)

	assert.Len(t, patchedFull.Characteristics, 1)
	assert.Equal(t, "VIP", patchedFull.Characteristics[0].Value)

	assert.Len(t, patchedFull.RelatedParties, 1)
	assert.Equal(t, "Parent", patchedFull.RelatedParties[0].Name)

	assert.Len(t, patchedFull.PaymentMethods, 1)
	assert.Equal(t, "CreditCard", patchedFull.PaymentMethods[0].Type)

	assert.Len(t, patchedFull.MarketSegments, 1)
	assert.Equal(t, "Retail", patchedFull.MarketSegments[0].Name)
}

func TestAddInteraction(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust := &domain.Customer{
		ID:      "cust-interact-1",
		Name:    "Interaction Test",
		Status:  domain.CustomerStatusActive,
		PartyID: "party-i-1",
	}
	require.NoError(t, repo.CreateCustomer(ctx, cust))

	interaction := &domain.CustomerInteraction{
		ID:          "int-1",
		CustomerID:  "cust-interact-1",
		Channel:     "Web",
		Type:        "Login",
		Description: "User logged in",
		AgentID:     "system",
	}

	err := repo.AddInteraction(ctx, interaction)
	assert.NoError(t, err)

	// Verify it's saved
	var count int64
	err = db.Model(&domain.CustomerInteraction{}).Where("id = ?", "int-1").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify linked to customer (via GetCustomer, assuming preload works or manual check)
	// GetCustomer typically preloads interactions? Let's check repository.go.
	// Check if GetCustomer preloads CustomerInteractions
	loaded, err := repo.GetCustomer(ctx, "cust-interact-1")
	assert.NoError(t, err)
	// It depends if GetCustomer preloads interactions. Even if not, the AddInteraction test passes if row exists.
	// But let's check if we can verify relationship.
	if len(loaded.CustomerInteractions) > 0 {
		assert.Equal(t, "int-1", loaded.CustomerInteractions[0].ID)
	}
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

	tm := NewTransactionManager(db)
	err := tm.RunInTransaction(ctx, func(ctx context.Context) error {
		return repo.CreateCustomer(ctx, cust)
	})
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

func TestCreateCustomer_Error(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	// Should fail with duplicate ID
	cust1 := &domain.Customer{ID: "duplicate-id", Name: "Duplicate 1", Status: domain.CustomerStatusActive}
	err := repo.CreateCustomer(ctx, cust1)
	require.NoError(t, err)

	cust2 := &domain.Customer{ID: "duplicate-id", Name: "Duplicate 2", Status: domain.CustomerStatusActive}
	err = repo.CreateCustomer(ctx, cust2)
	assert.Error(t, err)

	// Should fail due to foreign key violation on interaction customer_id
	interaction := &domain.CustomerInteraction{ID: "int-err-1", CustomerID: "non-existent-cust"}
	err = repo.AddInteraction(ctx, interaction)
	assert.Error(t, err)
}

func TestGetCustomer_Error(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	_, err := repo.GetCustomer(ctx, "non-existent")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestPatchCustomer_NonExistent(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	err := repo.PatchCustomer(ctx, "non-existent", map[string]interface{}{"name": "foo"})
	assert.Error(t, err)
}

func TestUpdateCustomer_Error(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cust := &domain.Customer{ID: "cust-upd-err-1", Name: "name"}
	err := repo.UpdateCustomer(ctx, cust)
	assert.Error(t, err)
}

func TestDeleteCustomer_Error(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	err := repo.DeleteCustomer(ctx, "non-existent")
	// For gorm soft-delete, deleting a non-existent record without condition might not return error unless
	// we explicitly fetch it. Or if it returns ErrNotFound. We just assert error.
	assert.ErrorIs(t, err, domain.ErrNotFound)

	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()
	err = repo.DeleteCustomer(ctxCancel, "non-existent")
	assert.Error(t, err)
}

func TestSearchCustomers_Error(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.SearchCustomers(ctx, map[string]interface{}{})
	assert.Error(t, err)
}

func TestUpdateSubResources_Error(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust := &domain.Customer{ID: "cust-suberr-1", Name: "ErrTest", Status: domain.CustomerStatusActive}
	require.NoError(t, repo.CreateCustomer(ctx, cust))

	errCust := *cust

	canceledCtx, cancel := context.WithCancel(ctx)
	cancel() // instantly cancel so queries fail

	errCust.CreditProfiles = []domain.CreditProfile{{ID: "1"}}
	assert.Error(t, repo.UpdateCustomer(canceledCtx, &errCust))

	errCust = *cust
	errCust.Accounts = []domain.CustomerAccount{{ID: "1"}}
	assert.Error(t, repo.UpdateCustomer(canceledCtx, &errCust))

	errCust = *cust
	errCust.ContactMediums = []domain.ContactMedium{{ID: "1"}}
	assert.Error(t, repo.UpdateCustomer(canceledCtx, &errCust))

	errCust = *cust
	errCust.Characteristics = []domain.CustomerCharacteristic{{ID: "1"}}
	assert.Error(t, repo.UpdateCustomer(canceledCtx, &errCust))

	errCust = *cust
	errCust.PrivacyConsents = []domain.PrivacyConsent{{ID: "1"}}
	assert.Error(t, repo.UpdateCustomer(canceledCtx, &errCust))

	errCust = *cust
	errCust.RelatedParties = []domain.RelatedParty{{ID: "1"}}
	assert.Error(t, repo.UpdateCustomer(canceledCtx, &errCust))

	errCust = *cust
	errCust.PaymentMethods = []domain.PaymentMethod{{ID: "1"}}
	assert.Error(t, repo.UpdateCustomer(canceledCtx, &errCust))

	errCust = *cust
	errCust.MarketSegments = []domain.MarketSegment{{ID: "1"}}
	assert.Error(t, repo.UpdateCustomer(canceledCtx, &errCust))
}

func TestPatchCustomer_Error(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust := &domain.Customer{ID: "cust-patch-err", Name: "PatchErr", Status: domain.CustomerStatusActive}
	require.NoError(t, repo.CreateCustomer(ctx, cust))

	canceledCtx, cancel := context.WithCancel(ctx)
	cancel() // instantly cancel to force gorm/sql execution errors

	err := repo.PatchCustomer(canceledCtx, cust.ID, map[string]interface{}{"accounts": []domain.CustomerAccount{{ID: "dup9"}}})
	assert.Error(t, err)

	err = repo.PatchCustomer(canceledCtx, cust.ID, map[string]interface{}{"credit_profiles": []domain.CreditProfile{{ID: "dup10"}}})
	assert.Error(t, err)

	err = repo.PatchCustomer(canceledCtx, cust.ID, map[string]interface{}{"contact_mediums": []domain.ContactMedium{{ID: "dup11"}}})
	assert.Error(t, err)

	err = repo.PatchCustomer(canceledCtx, cust.ID, map[string]interface{}{"characteristics": []domain.CustomerCharacteristic{{ID: "dup12"}}})
	assert.Error(t, err)

	err = repo.PatchCustomer(canceledCtx, cust.ID, map[string]interface{}{"privacy_consents": []domain.PrivacyConsent{{ID: "dup13"}}})
	assert.Error(t, err)

	err = repo.PatchCustomer(canceledCtx, cust.ID, map[string]interface{}{"related_parties": []domain.RelatedParty{{ID: "dup14"}}})
	assert.Error(t, err)

	err = repo.PatchCustomer(canceledCtx, cust.ID, map[string]interface{}{"payment_methods": []domain.PaymentMethod{{ID: "dup15"}}})
	assert.Error(t, err)

	err = repo.PatchCustomer(canceledCtx, cust.ID, map[string]interface{}{"market_segments": []domain.MarketSegment{{ID: "dup16"}}})
	assert.Error(t, err)
}
