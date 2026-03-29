package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"tmf/services/shopping-cart/internal/adapter/repository"
	"tmf/services/shopping-cart/internal/core/domain"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) *gorm.DB {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable"
	}

	m, err := migrate.New("file://../../../internal/infrastructure/postgres/migrations", dbURL)
	require.NoError(t, err)
	_ = m.Up()

	db, err := gorm.Open(gormPostgres.Open(dbURL), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func TestPostgresRepo_SaveAndGet(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewCartRepository(db)
	ctx := context.Background()

	t.Run("Should save and get cart successfully", func(t *testing.T) {
		cartID := uuid.New().String()
		offeringID := uuid.New().String()
		customerID := uuid.New().String()

		cart := &domain.Cart{
			ID:                 cartID,
			CustomerID:         customerID,
			Status:             domain.CartStatusActive,
			TotalPriceAmount:   100.0,
			TotalPriceCurrency: "USD",
			Version:            1,
			CreatedAt:          time.Now().UTC().Truncate(time.Microsecond),
			UpdatedAt:          time.Now().UTC().Truncate(time.Microsecond),
			Items: []domain.CartItem{
				{
					ID:         uuid.New().String(),
					CartID:     cartID,
					OfferingID: offeringID,
					Quantity:   2,
					UnitAmount: 50.0,
					Currency:   "USD",
				},
			},
		}

		events := []domain.OutboxEvent{
			{
				ID:        uuid.New().String(),
				Topic:     "test-topic",
				Payload:   []byte(`{"foo":"bar"}`),
				Status:    "PENDING",
				CreatedAt: time.Now().UTC(),
			},
		}

		err := repo.Save(ctx, cart, events)
		assert.NoError(t, err)

		retrieved, err := repo.Get(ctx, cartID)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		if retrieved != nil {
			assert.Equal(t, cartID, retrieved.ID)
			assert.Len(t, retrieved.Items, 1)
			if len(retrieved.Items) > 0 {
				assert.Equal(t, 2, retrieved.Items[0].Quantity)
				assert.Equal(t, offeringID, retrieved.Items[0].OfferingID)
			}
		}
	})

	t.Run("Should return nil when cart not found", func(t *testing.T) {
		retrieved, err := repo.Get(ctx, uuid.New().String())
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestPostgresRepo_UpsertAndGetPrice(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewCartRepository(db)
	ctx := context.Background()

	t.Run("Should upsert and get price", func(t *testing.T) {
		offeringID := uuid.New().String()
		price := &domain.ProductPrice{
			ID:         offeringID,
			UnitAmount: 99.99,
			Currency:   "EUR",
			UpdatedAt:  time.Now().UTC().Truncate(time.Microsecond),
		}

		err := repo.UpsertPrice(ctx, price)
		assert.NoError(t, err)

		price.UnitAmount = 89.99
		err = repo.UpsertPrice(ctx, price)
		assert.NoError(t, err)

		retrieved, err := repo.GetPrice(ctx, offeringID)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		if retrieved != nil {
			assert.Equal(t, offeringID, retrieved.ID)
			assert.Equal(t, 89.99, retrieved.UnitAmount)
		}
	})

	t.Run("Should return nil when price not found", func(t *testing.T) {
		retrieved, err := repo.GetPrice(ctx, uuid.New().String())
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestPostgresRepo_GetTxWithContext(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewCartRepository(db)
	type ctxKey string
	var txKey ctxKey = "tx"
	ctx := context.WithValue(context.Background(), txKey, db)
	retrieved, err := repo.Get(ctx, uuid.New().String())
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestPostgresRepo_SaveError(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewCartRepository(db)

	ctx := context.Background()
	cart := &domain.Cart{
		ID:         "invalid-uuid",
		CustomerID: "cust-1",
		Status:     domain.CartStatusActive,
	}

	err := repo.Save(ctx, cart, nil)
	assert.Error(t, err)
}

func TestPostgresRepo_UpsertPriceError(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewCartRepository(db)

	ctx := context.Background()
	price := &domain.ProductPrice{
		ID:         "invalid-uuid",
		UnitAmount: 99.99,
		Currency:   "EUR",
	}

	err := repo.UpsertPrice(ctx, price)
	assert.Error(t, err)
}
