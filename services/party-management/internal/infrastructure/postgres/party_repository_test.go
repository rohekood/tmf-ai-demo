package postgres

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"tmf/services/party-management/internal/domain"

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

func TestCreateIndividual(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "ind-1",
			Type:      domain.PartyTypeIndividual,
			Href:      "http://example.com/ind-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "John",
		FamilyName: "Doe",
	}

	err := repo.CreateIndividual(ctx, ind)
	assert.NoError(t, err)

	savedInd, err := repo.GetIndividual(ctx, "ind-1")
	assert.NoError(t, err)
	assert.Equal(t, ind.ID, savedInd.ID)
	assert.Equal(t, ind.GivenName, savedInd.GivenName)
	assert.Equal(t, ind.FamilyName, savedInd.FamilyName)
}

func TestCreateOrganization(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	org := &domain.Organization{
		Party: domain.Party{
			ID:        "org-1",
			Type:      domain.PartyTypeOrganization,
			Href:      "http://example.com/org-1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "Acme Corp",
		IsLegalEntity: true,
	}

	err := repo.CreateOrganization(ctx, org)
	assert.NoError(t, err)

	savedOrg, err := repo.GetOrganization(ctx, "org-1")
	assert.NoError(t, err)
	assert.Equal(t, org.ID, savedOrg.ID)
	assert.Equal(t, org.TradingName, savedOrg.TradingName)
	assert.Equal(t, org.IsLegalEntity, savedOrg.IsLegalEntity)
}

func TestUpdateIndividual(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	// Create first
	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "ind-update-1",
			Type:      domain.PartyTypeIndividual,
			Href:      "http://example.com/ind-update-1",
			Status:    "Initialized",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Jane",
		FamilyName: "Doe",
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind))

	// Update
	ind.GivenName = "Janet"
	ind.FamilyName = "Smith"
	ind.Status = "Active"
	ind.UpdatedAt = time.Now()

	err := repo.UpdateIndividual(ctx, ind)
	assert.NoError(t, err)

	// Verify
	updated, err := repo.GetIndividual(ctx, "ind-update-1")
	assert.NoError(t, err)
	assert.Equal(t, "Janet", updated.GivenName)
	assert.Equal(t, "Smith", updated.FamilyName)
	assert.Equal(t, "Active", updated.Status)
}

func TestUpdateOrganization(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	// Create first
	org := &domain.Organization{
		Party: domain.Party{
			ID:        "org-update-1",
			Type:      domain.PartyTypeOrganization,
			Href:      "http://example.com/org-update-1",
			Status:    "Initialized",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "Old Corp",
		IsLegalEntity: false,
	}
	require.NoError(t, repo.CreateOrganization(ctx, org))

	// Update
	org.TradingName = "New Corp"
	org.IsLegalEntity = true
	org.Status = "Validated"
	org.UpdatedAt = time.Now()

	err := repo.UpdateOrganization(ctx, org)
	assert.NoError(t, err)

	// Verify
	updated, err := repo.GetOrganization(ctx, "org-update-1")
	assert.NoError(t, err)
	assert.Equal(t, "New Corp", updated.TradingName)
	assert.Equal(t, true, updated.IsLegalEntity)
	assert.Equal(t, "Validated", updated.Status)
}

func TestDeleteParty_Individual(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	// Create first
	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "ind-delete-1",
			Type:      domain.PartyTypeIndividual,
			Href:      "http://example.com/ind-delete-1",
			Status:    "Initialized",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "ToDelete",
		FamilyName: "Person",
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind))

	// Delete
	err := repo.DeleteParty(ctx, "ind-delete-1")
	assert.NoError(t, err)

	// Verify deleted
	_, err = repo.GetIndividual(ctx, "ind-delete-1")
	assert.Error(t, err)
}

func TestDeleteParty_Organization(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	// Create first
	org := &domain.Organization{
		Party: domain.Party{
			ID:        "org-delete-1",
			Type:      domain.PartyTypeOrganization,
			Href:      "http://example.com/org-delete-1",
			Status:    "Initialized",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "ToDelete Corp",
		IsLegalEntity: true,
	}
	require.NoError(t, repo.CreateOrganization(ctx, org))

	// Delete
	err := repo.DeleteParty(ctx, "org-delete-1")
	assert.NoError(t, err)

	// Verify deleted
	_, err = repo.GetOrganization(ctx, "org-delete-1")
	assert.Error(t, err)
}

func TestSearchParties_ByGivenName(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	// Create test data
	ind1 := &domain.Individual{
		Party: domain.Party{
			ID:        "search-ind-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Alice",
		FamilyName: "Wonder",
	}
	ind2 := &domain.Individual{
		Party: domain.Party{
			ID:        "search-ind-2",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Bob",
		FamilyName: "Builder",
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind1))
	require.NoError(t, repo.CreateIndividual(ctx, ind2))

	// Search by GivenName
	results, err := repo.SearchParties(ctx, map[string]interface{}{
		"given_name": "Alice",
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-ind-1", results[0].ID)
}

func TestSearchParties_ByTradingName(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	// Create test data
	org1 := &domain.Organization{
		Party: domain.Party{
			ID:        "search-org-1",
			Type:      domain.PartyTypeOrganization,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "TechCorp",
		IsLegalEntity: true,
	}
	org2 := &domain.Organization{
		Party: domain.Party{
			ID:        "search-org-2",
			Type:      domain.PartyTypeOrganization,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "FinanceInc",
		IsLegalEntity: true,
	}
	require.NoError(t, repo.CreateOrganization(ctx, org1))
	require.NoError(t, repo.CreateOrganization(ctx, org2))

	// Search by TradingName
	results, err := repo.SearchParties(ctx, map[string]interface{}{
		"trading_name": "TechCorp",
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-org-1", results[0].ID)
}

func TestSearchParties_ByType(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	// Create test data
	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "type-search-ind",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "TypeTest",
		FamilyName: "Individual",
	}
	org := &domain.Organization{
		Party: domain.Party{
			ID:        "type-search-org",
			Type:      domain.PartyTypeOrganization,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TradingName:   "TypeTest Org",
		IsLegalEntity: true,
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind))
	require.NoError(t, repo.CreateOrganization(ctx, org))

	// Search by Type
	results, err := repo.SearchParties(ctx, map[string]interface{}{
		"type": domain.PartyTypeOrganization,
	})
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "type-search-org", results[0].ID)
}
