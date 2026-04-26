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

	ind.TaxExemptions = []domain.TaxExemption{
		{
			ID:                  "55555555-5555-5555-5555-555555555555",
			PartyID:             ind.ID,
			CertificateNumber:   "TAX-CERT-001",
			IssuingJurisdiction: "US-CA",
			ValidForStart:       time.Now(),
		},
	}

	err := repo.CreateIndividual(ctx, ind)
	require.NoError(t, err)

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

	ind1 := &domain.Individual{
		Party:     domain.Party{ID: "search-ext-b-1", Type: domain.PartyTypeIndividual, Status: "Active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		GivenName: "ExtAlice",
	}
	ind1.ExternalReferences = []domain.ExternalReference{
		{ID: "11111111-aaaa-1111-1111-111111111111", PartyID: ind1.ID, ExternalSystemID: "SysA", ExternalReferenceID: "12345"},
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind1))

	ind2 := &domain.Individual{
		Party:     domain.Party{ID: "search-ext-b-2", Type: domain.PartyTypeIndividual, Status: "Active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		GivenName: "ExtBob",
	}
	ind2.ExternalReferences = []domain.ExternalReference{
		{ID: "22222222-aaaa-2222-2222-222222222222", PartyID: ind2.ID, ExternalSystemID: "SysA", ExternalReferenceID: "67890"},
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind2))

	results, err := repo.SearchParties(ctx, map[string]any{
		"externalReference": "12345",
	})
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(results), 1)
	found := false
	for _, p := range results {
		if p.ID == "search-ext-b-1" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestUpdateIndividual_References(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	ind := &domain.Individual{
		Party:     domain.Party{ID: "upd-ref-b-1", Type: domain.PartyTypeIndividual, Status: "Active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		GivenName: "Original",
	}
	ind.ExternalReferences = []domain.ExternalReference{{ID: "33333333-aaaa-3333-3333-333333333333", PartyID: ind.ID, ExternalSystemID: "OldSys", ExternalReferenceID: "XXX"}}
	require.NoError(t, repo.CreateIndividual(ctx, ind))

	ind.ExternalReferences = []domain.ExternalReference{
		{ID: "44444444-aaaa-4444-4444-444444444444", PartyID: ind.ID, ExternalSystemID: "NewSys", ExternalReferenceID: "YYY"},
	}

	err := repo.UpdateIndividual(ctx, ind)
	require.NoError(t, err)

	updated, err := repo.GetIndividual(ctx, "upd-ref-b-1")
	require.NoError(t, err)

	assert.Len(t, updated.ExternalReferences, 1)
	assert.Equal(t, "NewSys", updated.ExternalReferences[0].ExternalSystemID)
}
