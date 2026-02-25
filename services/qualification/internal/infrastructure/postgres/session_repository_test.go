package postgres_test

import (
	"context"
	"database/sql"
	"os"

	"testing"
	"time"

	"tmf/services/qualification/internal/core/domain"
	"tmf/services/qualification/internal/infrastructure/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		dbURL = "postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable"
	}

	m, err := migrate.New("file://../../../internal/infrastructure/postgres/migrations", dbURL)
	require.NoError(t, err)
	_ = m.Up() // ignore ErrNoChange

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)

	_, err = db.Exec("TRUNCATE TABLE qualification_sessions CASCADE")
	if err != nil {
		t.Logf("truncate failed: %v", err)
	}

	return db
}

func TestSessionRepository(t *testing.T) {
	db := setupDB(t)
	repo := postgres.NewSessionRepository(db)
	ctx := context.Background()

	t.Run("Create and Get Session", func(t *testing.T) {
		session := &domain.QualificationSession{
			CustomerID: "cust-1",
			Address: domain.Address{
				Street: "123 Main St",
			},
			QualifiedOffers: []domain.QualifiedOffer{
				{
					OfferingID:  "off-1",
					Eligibility: "QUALIFIED",
					Price:       domain.Price{Amount: 10.0, Currency: "USD"},
				},
			},
			Status: "Active",
		}

		id, err := repo.Create(ctx, session)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		retrieved, err := repo.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, retrieved.ID)
		assert.Equal(t, "cust-1", retrieved.CustomerID)
		assert.Equal(t, "123 Main St", retrieved.Address.Street)
		assert.Len(t, retrieved.QualifiedOffers, 1)
		assert.Equal(t, "off-1", retrieved.QualifiedOffers[0].OfferingID)
	})

	t.Run("Update Session", func(t *testing.T) {
		id, err := repo.Create(ctx, &domain.QualificationSession{})
		require.NoError(t, err)

		updateSess := &domain.QualificationSession{
			ID:         id,
			CustomerID: "cust-2",
			Address:    domain.Address{Street: "456 Oak"},
			Status:     "Expired",
		}

		err = repo.Update(ctx, updateSess)
		require.NoError(t, err)

		retrieved, _ := repo.Get(ctx, id)
		assert.Equal(t, "cust-2", retrieved.CustomerID)
		assert.Equal(t, "Expired", retrieved.Status)
	})

	t.Run("Delete Session", func(t *testing.T) {
		id, err := repo.Create(ctx, &domain.QualificationSession{})
		require.NoError(t, err)

		err = repo.Delete(ctx, id)
		require.NoError(t, err)

		_, err = repo.Get(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("FindExpired", func(t *testing.T) {
		_, err := repo.Create(ctx, &domain.QualificationSession{
			Status:    "Active",
			ExpiresAt: time.Now().UTC().Add(-1 * time.Hour), // Expired
		})
		require.NoError(t, err)
		_, err = repo.Create(ctx, &domain.QualificationSession{
			Status:    "Active",
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour), // Not expired
		})
		require.NoError(t, err)

		expired, err := repo.FindExpired(ctx)
		require.NoError(t, err)
		assert.True(t, len(expired) >= 1)
	})

	t.Run("Not Found Errors", func(t *testing.T) {
		badID := uuid.New().String()
		_, err := repo.Get(ctx, badID)
		assert.Error(t, err)

		err = repo.Update(ctx, &domain.QualificationSession{ID: badID})
		assert.Error(t, err)

		err = repo.Delete(ctx, badID)
		assert.Error(t, err)
	})

	t.Run("Create Errors", func(t *testing.T) {
		// Mock errors via bad JSON or db drop
		err := repo.Update(ctx, &domain.QualificationSession{ID: ""})
		assert.Error(t, err) // ID is empty so no rows affected
	})

	t.Run("JSON Unmarshal Errors", func(t *testing.T) {
		// Insert syntactically valid JSON that fails struct unmarshaling
		badID := uuid.New().String()
		_, err := db.ExecContext(ctx, "INSERT INTO qualification_sessions (id, customer_id, address, qualified_offers, status, created_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			badID, "test-cust", "123", "[]", "Active", time.Now().UTC(), time.Now().UTC().Add(time.Hour))
		require.NoError(t, err)

		_, err = repo.Get(ctx, badID)
		assert.ErrorContains(t, err, "failed to unmarshal address")

		// Fix address, break offers
		_, err = db.ExecContext(ctx, "UPDATE qualification_sessions SET address = '{}', qualified_offers = '123' WHERE id = $1", badID)
		require.NoError(t, err)

		_, err = repo.Get(ctx, badID)
		assert.ErrorContains(t, err, "failed to unmarshal qualified offers")

		// Make it expired to test FindExpired
		_, err = db.ExecContext(ctx, "UPDATE qualification_sessions SET expires_at = $1 WHERE id = $2", time.Now().UTC().Add(-time.Hour), badID)
		require.NoError(t, err)

		_, err = repo.FindExpired(ctx)
		assert.ErrorContains(t, err, "failed to unmarshal qualified offers")

		// Fix offers, break address
		_, err = db.ExecContext(ctx, "UPDATE qualification_sessions SET address = '123', qualified_offers = '[]' WHERE id = $1", badID)
		require.NoError(t, err)

		_, err = repo.FindExpired(ctx)
		assert.ErrorContains(t, err, "failed to unmarshal address")
	})

	t.Run("DB Errors", func(t *testing.T) {
		// Create a separate db connection to test errors like closed db
		dbBad, err := sql.Open("postgres", "postgres://user:pass@127.0.0.1:1/bad?sslmode=disable&connect_timeout=1")
		require.NoError(t, err)
		repoBad := postgres.NewSessionRepository(dbBad)

		_, err = repoBad.Create(ctx, &domain.QualificationSession{})
		assert.Error(t, err)

		_, err = repoBad.Get(ctx, "some-id")
		assert.Error(t, err)

		err = repoBad.Update(ctx, &domain.QualificationSession{})
		assert.Error(t, err)

		err = repoBad.Delete(ctx, "some-id")
		assert.Error(t, err)

		_, err = repoBad.FindExpired(ctx)
		assert.Error(t, err)
	})
}
