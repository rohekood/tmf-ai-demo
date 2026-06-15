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

func TestSearchParties_ByEmail(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	target := &domain.Individual{
		Party: domain.Party{
			ID:        "email-ind-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ContactMediums: []domain.ContactMedium{
				{ID: "cm-email-1", PartyID: "email-ind-1", MediumType: "email", Preferred: true, Value: "jane.provision@example.com"},
				{ID: "cm-phone-1", PartyID: "email-ind-1", MediumType: "phone", Value: "555-0100"},
			},
		},
		GivenName:  "Jane",
		FamilyName: "Provision",
	}
	other := &domain.Individual{
		Party: domain.Party{
			ID:        "email-ind-2",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ContactMediums: []domain.ContactMedium{
				{ID: "cm-email-2", PartyID: "email-ind-2", MediumType: "email", Value: "someone.else@example.com"},
			},
		},
		GivenName:  "Someone",
		FamilyName: "Else",
	}

	require.NoError(t, repo.CreateIndividual(ctx, target))
	require.NoError(t, repo.CreateIndividual(ctx, other))

	// Exact email match returns exactly the owning party (no duplicate rows).
	results, err := repo.SearchParties(ctx, map[string]any{
		"email": "jane.provision@example.com",
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "email-ind-1", results[0].ID)

	// A non-email medium with the same value must not match the email filter.
	results, err = repo.SearchParties(ctx, map[string]any{
		"email": "555-0100",
	})
	assert.NoError(t, err)
	assert.Empty(t, results)

	// Unknown email returns no rows.
	results, err = repo.SearchParties(ctx, map[string]any{
		"email": "nobody@example.com",
	})
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestPartyEmail_UniqueConstraint(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	first := &domain.Individual{
		Party: domain.Party{
			ID:        "uniq-ind-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ContactMediums: []domain.ContactMedium{
				{ID: "uniq-cm-1", PartyID: "uniq-ind-1", MediumType: "email", Value: "dupe@example.com"},
			},
		},
		GivenName: "First", FamilyName: "Owner",
	}
	require.NoError(t, repo.CreateIndividual(ctx, first))

	// Same email (different case) on another party must be rejected.
	second := &domain.Individual{
		Party: domain.Party{
			ID:        "uniq-ind-2",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ContactMediums: []domain.ContactMedium{
				{ID: "uniq-cm-2", PartyID: "uniq-ind-2", MediumType: "email", Value: "DUPE@example.com"},
			},
		},
		GivenName: "Second", FamilyName: "Owner",
	}
	err := repo.CreateIndividual(ctx, second)
	assert.Error(t, err, "expected unique-email constraint to reject a duplicate email")

	// A different email is fine.
	third := &domain.Individual{
		Party: domain.Party{
			ID:        "uniq-ind-3",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ContactMediums: []domain.ContactMedium{
				{ID: "uniq-cm-3", PartyID: "uniq-ind-3", MediumType: "email", Value: "unique@example.com"},
			},
		},
		GivenName: "Third", FamilyName: "Owner",
	}
	assert.NoError(t, repo.CreateIndividual(ctx, third))
}
