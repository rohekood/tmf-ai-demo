package postgres

import (
	"context"
	"testing"
	"tmf/services/customer-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchCustomers_Generic(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust1 := &domain.Customer{ID: "gen-1", Name: "John Doe", Status: domain.CustomerStatusActive, PartyID: "p-1", PartyType: "Individual"}
	cust2 := &domain.Customer{ID: "gen-2", Name: "Jane Smith", Status: domain.CustomerStatusSuspended, PartyID: "p-2", PartyType: "Organization"}
	require.NoError(t, repo.CreateCustomer(ctx, cust1))
	require.NoError(t, repo.CreateCustomer(ctx, cust2))

	// Search Generic "John" (Name)
	results, err := repo.SearchCustomers(ctx, map[string]interface{}{"search": "John"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "gen-1", results[0].ID)

	// Search Generic "Active" (Status)
	results, err = repo.SearchCustomers(ctx, map[string]interface{}{"search": "Active"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "gen-1", results[0].ID)

	// Search Generic "p-2" (PartyID)
	results, err = repo.SearchCustomers(ctx, map[string]interface{}{"search": "p-2"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "gen-2", results[0].ID)

	// Search Generic "Organization" (PartyType)
	results, err = repo.SearchCustomers(ctx, map[string]interface{}{"search": "Organization"})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "gen-2", results[0].ID)
}

func TestSearchCustomers_WithAccounts(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewCustomerRepository(db)
	ctx := context.Background()

	cust := &domain.Customer{
		ID:        "cust-acc-1",
		Name:      "Customer With Account",
		Status:    domain.CustomerStatusActive,
		PartyID:   "p-acc-1",
		PartyType: "Individual",
		Accounts: []domain.CustomerAccount{
			{ID: "acc-1", Name: "Main Account", AccountStatus: "active", AccountType: "postpaid"},
		},
	}
	require.NoError(t, repo.CreateCustomer(ctx, cust))

	// Search and verify accounts are loaded
	results, err := repo.SearchCustomers(ctx, map[string]interface{}{"name": "Customer With Account"})
	assert.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "cust-acc-1", results[0].ID)
	// This assertion fails if Preload("Accounts") is missing
	assert.Len(t, results[0].Accounts, 1)
	assert.Equal(t, "acc-1", results[0].Accounts[0].ID)
}
