package postgres

import (
	"context"
	"testing"
	"tmf/services/party-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestOutboxPublisher(t *testing.T) {
	// Setup DB
	db, err := gorm.Open(gormPostgres.Open("host=localhost user=postgres password=postgres dbname=testdb port=5432 sslmode=disable"), &gorm.Config{})
	require.NoError(t, err)

	repo := NewOutboxRepository(db)
	pub := NewOutboxPublisher(repo)

	ctx := context.WithValue(context.Background(), domain.UserContextKey, "test-user")
	ctx = context.WithValue(ctx, domain.AuthContextKey, "test-auth")

	err = pub.Publish(ctx, "exchange", "routing.key", map[string]string{"foo": "bar"})
	require.NoError(t, err)

	// Verify
	events, err := repo.FetchPending(ctx, 10)
	require.NoError(t, err)

	found := false
	for _, e := range events {
		if e.RoutingKey == "routing.key" {
			found = true
			assert.Contains(t, string(e.Payload), "bar")
			assert.Contains(t, string(e.Headers), "test-user")
			assert.Contains(t, string(e.Headers), "test-auth")
		}
	}
	assert.True(t, found)
}
