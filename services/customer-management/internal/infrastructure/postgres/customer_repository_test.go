package postgres

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"tmf/services/customer-management/internal/domain"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, string) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Run migrations
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "migrations")

	m, err := migrate.New(
		"file://"+migrationsPath,
		connStr,
	)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	// Connect GORM
	db, err := gorm.Open(gormPostgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	return db, connStr
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
