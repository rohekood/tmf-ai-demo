package postgres

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
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

var (
	testDB      *gorm.DB
	testConnStr string
	testDBOnce  sync.Once
	testDBErr   error
)

func setupTestDB(t *testing.T) (*gorm.DB, string) {
	testDBOnce.Do(func() {
		ctx := context.Background()

		defer func() {
			if r := recover(); r != nil {
				testDBErr = fmt.Errorf("testcontainers panic: %v", r)
			}
		}()

		pgContainer, err := postgres.Run(ctx,
			"postgres:15",
			postgres.WithDatabase("testdb"),
			postgres.WithUsername("postgres"),
			postgres.WithPassword("postgres"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*time.Second)),
		)
		if err != nil {
			testDBErr = err
			return
		}

		connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			testDBErr = err
			return
		}

		_, filename, _, _ := runtime.Caller(0)
		migrationsPath := filepath.Join(filepath.Dir(filename), "migrations")

		m, err := migrate.New("file://"+migrationsPath, connStr)
		if err != nil {
			testDBErr = fmt.Errorf("migration new error: %w", err)
			return
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			testDBErr = fmt.Errorf("migration up error: %w", err)
			return
		}

		db, err := gorm.Open(gormPostgres.Open(connStr), &gorm.Config{})
		if err != nil {
			testDBErr = fmt.Errorf("gorm open error: %w", err)
			return
		}

		testDB = db
		testConnStr = connStr
	})

	if testDBErr != nil {
		t.Skipf("Skipping test: PostgreSQL container unavailable: %v", testDBErr)
		return nil, ""
	}

	return testDB, testConnStr
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
		MiddleName: "M",
		BirthDate:  "1990-01-01",
		Gender:     "Male",
	}

	err := repo.CreateIndividual(ctx, ind)
	assert.NoError(t, err)

	savedInd, err := repo.GetIndividual(ctx, "ind-1")
	assert.NoError(t, err)
	assert.Equal(t, ind.ID, savedInd.ID)
	assert.Equal(t, ind.GivenName, savedInd.GivenName)
	assert.Equal(t, ind.FamilyName, savedInd.FamilyName)
	assert.Equal(t, ind.MiddleName, savedInd.MiddleName)
	assert.Equal(t, ind.BirthDate, savedInd.BirthDate)
	assert.Equal(t, ind.Gender, savedInd.Gender)
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
		TradingName:      "Acme Corp",
		IsLegalEntity:    true,
		OrganizationType: "LLC",
	}

	err := repo.CreateOrganization(ctx, org)
	assert.NoError(t, err)

	savedOrg, err := repo.GetOrganization(ctx, "org-1")
	assert.NoError(t, err)
	assert.Equal(t, org.ID, savedOrg.ID)
	assert.Equal(t, org.TradingName, savedOrg.TradingName)
	assert.Equal(t, org.IsLegalEntity, savedOrg.IsLegalEntity)
	assert.Equal(t, org.OrganizationType, savedOrg.OrganizationType)
}

func TestUpdateIndividual(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

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

	ind.GivenName = "Janet"
	ind.FamilyName = "Smith"
	ind.MiddleName = "K"
	ind.BirthDate = "1995-05-05"
	ind.Gender = "Female"
	ind.Status = "Active"
	ind.UpdatedAt = time.Now()

	err := repo.UpdateIndividual(ctx, ind)
	assert.NoError(t, err)

	updated, err := repo.GetIndividual(ctx, "ind-update-1")
	assert.NoError(t, err)
	assert.Equal(t, "Janet", updated.GivenName)
	assert.Equal(t, "Smith", updated.FamilyName)
	assert.Equal(t, "K", updated.MiddleName)
	assert.Equal(t, "1995-05-05", updated.BirthDate)
	assert.Equal(t, "Female", updated.Gender)
	assert.Equal(t, "Active", updated.Status)
}

func TestUpdateOrganization(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

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

	org.TradingName = "New Corp"
	org.IsLegalEntity = true
	org.OrganizationType = "Inc"
	org.Status = "Validated"
	org.UpdatedAt = time.Now()

	err := repo.UpdateOrganization(ctx, org)
	assert.NoError(t, err)

	updated, err := repo.GetOrganization(ctx, "org-update-1")
	assert.NoError(t, err)
	assert.Equal(t, "New Corp", updated.TradingName)
	assert.Equal(t, "New Corp", updated.TradingName)
	assert.Equal(t, true, updated.IsLegalEntity)
	assert.Equal(t, "Inc", updated.OrganizationType)
	assert.Equal(t, "Validated", updated.Status)
}

func TestDeleteParty_Individual(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

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

	err := repo.DeleteParty(ctx, "ind-delete-1")
	assert.NoError(t, err)

	_, err = repo.GetIndividual(ctx, "ind-delete-1")
	assert.Error(t, err)
}

func TestDeleteParty_Organization(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

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

	err := repo.DeleteParty(ctx, "org-delete-1")
	assert.NoError(t, err)

	_, err = repo.GetOrganization(ctx, "org-delete-1")
	assert.Error(t, err)
}

func TestUpdateParty_TypeSwitch(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

	ind := &domain.Individual{
		Party: domain.Party{
			ID:        "switch-party-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GivenName:  "Test",
		FamilyName: "Switcher",
	}
	require.NoError(t, repo.CreateIndividual(ctx, ind))

	org := &domain.Organization{
		Party: domain.Party{
			ID:        "switch-party-1",
			Type:      domain.PartyTypeOrganization,
			Status:    "Active",
			CreatedAt: ind.CreatedAt,
			UpdatedAt: time.Now(),
		},
		TradingName:      "Switch Corp",
		IsLegalEntity:    true,
		OrganizationType: "LLC",
	}

	err := repo.UpdateOrganization(ctx, org)
	assert.NoError(t, err)

	_, err = repo.GetIndividual(ctx, "switch-party-1")
	assert.Error(t, err)

	savedOrg, err := repo.GetOrganization(ctx, "switch-party-1")
	assert.NoError(t, err)
	assert.Equal(t, "Switch Corp", savedOrg.TradingName)
	assert.Equal(t, "LLC", savedOrg.OrganizationType)

	indNew := &domain.Individual{
		Party: domain.Party{
			ID:        "switch-party-1",
			Type:      domain.PartyTypeIndividual,
			Status:    "Active",
			CreatedAt: ind.CreatedAt,
			UpdatedAt: time.Now(),
		},
		GivenName:  "BackTo",
		FamilyName: "Individual",
	}

	err = repo.UpdateIndividual(ctx, indNew)
	assert.NoError(t, err)

	_, err = repo.GetOrganization(ctx, "switch-party-1")
	assert.Error(t, err)

	savedInd, err := repo.GetIndividual(ctx, "switch-party-1")
	assert.NoError(t, err)
	assert.Equal(t, "BackTo", savedInd.GivenName)
}

func TestSearchParties_ByGivenName(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)
	ctx := context.Background()

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

	results, err := repo.SearchParties(ctx, map[string]any{
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

	results, err := repo.SearchParties(ctx, map[string]any{
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

	results, err := repo.SearchParties(ctx, map[string]any{
		"type": domain.PartyTypeOrganization,
	})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)

	found := false
	for _, p := range results {
		if p.ID == "type-search-org" {
			found = true
			break
		}
	}
	assert.True(t, found)
}
