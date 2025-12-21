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
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
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

	err := repo.CreateIndividual(ind)
	assert.NoError(t, err)

	savedInd, err := repo.GetIndividual("ind-1")
	assert.NoError(t, err)
	assert.Equal(t, ind.ID, savedInd.ID)
	assert.Equal(t, ind.GivenName, savedInd.GivenName)
	assert.Equal(t, ind.FamilyName, savedInd.FamilyName)
}

func TestCreateOrganization(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPartyRepository(db)

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

	err := repo.CreateOrganization(org)
	assert.NoError(t, err)

	savedOrg, err := repo.GetOrganization("org-1")
	assert.NoError(t, err)
	assert.Equal(t, org.ID, savedOrg.ID)
	assert.Equal(t, org.TradingName, savedOrg.TradingName)
	assert.Equal(t, org.IsLegalEntity, savedOrg.IsLegalEntity)
}
