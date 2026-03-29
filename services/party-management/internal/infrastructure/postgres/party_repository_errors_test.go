package postgres

import (
	"context"
	"testing"
	"time"

	"tmf/services/party-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartyRepository_ContextErrors(t *testing.T) {
	db, _ := setupTestDB(t)

	repo := NewPartyRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("GetParty", func(t *testing.T) {
		_, err := repo.GetParty(ctx, "some-id")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("GetIndividual", func(t *testing.T) {
		_, err := repo.GetIndividual(ctx, "some-id")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("GetOrganization", func(t *testing.T) {
		_, err := repo.GetOrganization(ctx, "some-id")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("CreateIndividual", func(t *testing.T) {
		err := repo.CreateIndividual(ctx, &domain.Individual{
			Party: domain.Party{ID: "err-ind-1", Type: domain.PartyTypeIndividual},
		})
		assert.Error(t, err)
	})

	t.Run("CreateOrganization", func(t *testing.T) {
		err := repo.CreateOrganization(ctx, &domain.Organization{
			Party: domain.Party{ID: "err-org-1", Type: domain.PartyTypeOrganization},
		})
		assert.Error(t, err)
	})

	t.Run("UpdateIndividual", func(t *testing.T) {
		err := repo.UpdateIndividual(ctx, &domain.Individual{
			Party: domain.Party{ID: "err-ind-2", Type: domain.PartyTypeIndividual},
		})
		assert.Error(t, err)
	})

	t.Run("UpdateOrganization", func(t *testing.T) {
		err := repo.UpdateOrganization(ctx, &domain.Organization{
			Party: domain.Party{ID: "err-org-2", Type: domain.PartyTypeOrganization},
		})
		assert.Error(t, err)
	})

	t.Run("DeleteParty", func(t *testing.T) {
		err := repo.DeleteParty(ctx, "some-id")
		assert.Error(t, err)
		assert.NotErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("SearchParties", func(t *testing.T) {
		_, err := repo.SearchParties(ctx, map[string]interface{}{"id": "123"})
		assert.Error(t, err)
	})
}

func TestPartyRepository_DuplicateErrors(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "dup-ind-1",
			Type:      domain.PartyTypeIndividual,
			CreatedAt: time.Now(),
		},
		GivenName: "Dup",
	}

	err := repo.CreateIndividual(ctx, ind)
	require.NoError(t, err)

	err = repo.CreateIndividual(ctx, ind)
	assert.Error(t, err)

	org := &domain.Organization{
		Party: domain.Party{
			ID:        "dup-org-1",
			Type:      domain.PartyTypeOrganization,
			CreatedAt: time.Now(),
		},
		TradingName: "Dup Corp",
	}

	err = repo.CreateOrganization(ctx, org)
	require.NoError(t, err)

	err = repo.CreateOrganization(ctx, org)
	assert.Error(t, err)
}
