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

	// Create test data
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

	// 1. Search for "Smith" (Name - Matches Ind and Org)
	results, err := repo.SearchParties(ctx, map[string]interface{}{
		"search": "Smith",
	})
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	ids := map[string]bool{}
	for _, p := range results {
		ids[p.ID] = true
	}
	assert.True(t, ids["gen-ind-1"])
	assert.True(t, ids["gen-org-1"])

	// 2. Search for "John" (Name - Ind GivenName)
	results, err = repo.SearchParties(ctx, map[string]interface{}{
		"search": "John",
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "gen-ind-1", results[0].ID)

	// 3. Search for "Enterprises" (Name - Org TradingName)
	results, err = repo.SearchParties(ctx, map[string]interface{}{
		"search": "Enterprises",
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "gen-org-1", results[0].ID)

	// 4. Search for "gen-ind-2" (ID)
	results, err = repo.SearchParties(ctx, map[string]interface{}{
		"search": "gen-ind-2",
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "gen-ind-2", results[0].ID)

	// 5. Search for "Individual" (Type)
	results, err = repo.SearchParties(ctx, map[string]interface{}{
		"search": "Individual",
	})
	assert.NoError(t, err)
	// Should match gen-ind-1 and gen-ind-2
	assert.Len(t, results, 2)
	ids = map[string]bool{}
	for _, p := range results {
		ids[p.ID] = true
	}
	assert.True(t, ids["gen-ind-1"])
	assert.True(t, ids["gen-ind-2"])
}
