package postgres

import (
	"context"
	"testing"
	"time"
	"tmf/services/party-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIndividual_WithReferencesAndTax(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "ind-ref-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Ref",
		FamilyName: "Tester",
	}

	// 1. External References
	ref1 := domain.ExternalReference{
		ID:                  "33333333-3333-3333-3333-333333333333",
		PartyID:             ind.ID,
		ExternalSystemID:    "LegacyCRM",
		ExternalReferenceID: "CUST-001",
	}
	ref2 := domain.ExternalReference{
		ID:                  "44444444-4444-4444-4444-444444444444",
		PartyID:             ind.ID,
		ExternalSystemID:    "Billing",
		ExternalReferenceID: "ACC-999",
	}
	ind.ExternalReferences = []domain.ExternalReference{ref1, ref2}

	// 2. Tax Exemptions
	ind.TaxExemptions = []domain.TaxExemption{
		{
			ID:                  "55555555-5555-5555-5555-555555555555",
			PartyID:             ind.ID,
			CertificateNumber:   "TAX-CERT-001",
			IssuingJurisdiction: "US-CA",
			ValidForStart:       time.Now(),
		},
	}

	// ACTION
	err := repo.CreateIndividual(ctx, ind)
	require.NoError(t, err)

	// VERIFICATION	// Verify
	saved, err := repo.GetIndividual(ctx, "ind-ref-1")
	assert.NoError(t, err)
	assert.Len(t, saved.ExternalReferences, 2)
	assert.Len(t, saved.TaxExemptions, 1)

	assert.Equal(t, "33333333-3333-3333-3333-333333333333", saved.ExternalReferences[0].ID)
	assert.Equal(t, "55555555-5555-5555-5555-555555555555", saved.TaxExemptions[0].ID)
	assert.Len(t, saved.TaxExemptions, 1)
	assert.Equal(t, "TAX-CERT-001", saved.TaxExemptions[0].CertificateNumber)
}

func TestSearchParties_ByExternalReference(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	// Setup Data
	ind1 := &domain.Individual{
		Party:     domain.Party{ID: "search-ext-1", Type: domain.PartyTypeIndividual, Status: "Active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		GivenName: "Alice",
	}
	ind1.ExternalReferences = []domain.ExternalReference{
		{ID: "11111111-1111-1111-1111-111111111111", PartyID: ind1.ID, ExternalSystemID: "SysA", ExternalReferenceID: "12345"},
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind1))

	ind2 := &domain.Individual{
		Party:     domain.Party{ID: "search-ext-2", Type: domain.PartyTypeIndividual, Status: "Active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		GivenName: "Bob",
	}
	ind2.ExternalReferences = []domain.ExternalReference{
		{ID: "22222222-2222-2222-2222-222222222222", PartyID: ind2.ID, ExternalSystemID: "SysA", ExternalReferenceID: "67890"},
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind2))

	// Action: Search
	results, err := repo.SearchParties(ctx, map[string]interface{}{
		"externalReference": "12345",
	})
	require.NoError(t, err)

	// Verify
	assert.Len(t, results, 1)
	assert.Equal(t, "search-ext-1", results[0].ID)
}

func TestUpdateIndividual_References(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	// Initial State
	ind := &domain.Individual{
		Party:     domain.Party{ID: "upd-ref-1", Type: domain.PartyTypeIndividual, Status: "Active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		GivenName: "Original",
	}
	ind.ExternalReferences = []domain.ExternalReference{{ID: "33333333-3333-3333-3333-333333333333", PartyID: ind.ID, ExternalSystemID: "OldSys", ExternalReferenceID: "XXX"}}
	require.NoError(t, repo.CreateIndividual(ctx, ind))

	// Update
	ind.ExternalReferences = []domain.ExternalReference{
		{ID: "44444444-4444-4444-4444-444444444444", PartyID: ind.ID, ExternalSystemID: "NewSys", ExternalReferenceID: "YYY"}, // Replaces old
	}

	err := repo.UpdateIndividual(ctx, ind)
	require.NoError(t, err)

	// Verify
	updated, err := repo.GetIndividual(ctx, "upd-ref-1")
	require.NoError(t, err)

	assert.Len(t, updated.ExternalReferences, 1)
	assert.Equal(t, "NewSys", updated.ExternalReferences[0].ExternalSystemID)
}
