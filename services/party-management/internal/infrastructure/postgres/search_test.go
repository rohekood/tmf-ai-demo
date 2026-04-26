package postgres

import (
	"context"
	"testing"
	"time"
	"tmf/services/party-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchParties_Generic(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	ind1 := &domain.Individual{
		Party: domain.Party{
			ID:        "gen-ind-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "John",
		FamilyName: "Smith",
	}
	org1 := &domain.Organization{
		Party: domain.Party{
			ID:        "gen-org-1",
			Type:      domain.PartyTypeOrganization,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "Smith Enterprises",
		IsLegalEntity: true,
	}
	ind2 := &domain.Individual{
		Party: domain.Party{
			ID:        "gen-ind-2",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Alice",
		FamilyName: "Wonder",
	}

	require.NoError(t, repo.CreateIndividual(ctx, ind1))
	require.NoError(t, repo.CreateOrganization(ctx, org1))
	require.NoError(t, repo.CreateIndividual(ctx, ind2))

	results, err := repo.SearchParties(ctx, map[string]any{
		"search": "Smith",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
	ids := map[string]bool{}
	for _, p := range results {
		ids[p.ID] = true
	}
	assert.True(t, ids["gen-ind-1"])
	assert.True(t, ids["gen-org-1"])

	results, err = repo.SearchParties(ctx, map[string]any{
		"search": "John",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
	ids = map[string]bool{}
	for _, p := range results {
		ids[p.ID] = true
	}
	assert.True(t, ids["gen-ind-1"])

	results, err = repo.SearchParties(ctx, map[string]any{
		"search": "Enterprises",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
	ids = map[string]bool{}
	for _, p := range results {
		ids[p.ID] = true
	}
	assert.True(t, ids["gen-org-1"])

	results, err = repo.SearchParties(ctx, map[string]any{
		"search": "gen-ind-2",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
	ids = map[string]bool{}
	for _, p := range results {
		ids[p.ID] = true
	}
	assert.True(t, ids["gen-ind-2"])

	results, err = repo.SearchParties(ctx, map[string]any{
		"search": "Individual",
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)
	ids = map[string]bool{}
	for _, p := range results {
		ids[p.ID] = true
	}
	assert.True(t, ids["gen-ind-1"])
	assert.True(t, ids["gen-ind-2"])
}
