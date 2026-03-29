package postgres

import (
	"context"
	"testing"
	"tmf/services/party-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutboxPublisher(t *testing.T) {
	db, _ := setupTestDB(t)

	repo := NewOutboxRepository(db)
	pub := NewOutboxPublisher(repo)

	ctx := context.WithValue(context.Background(), domain.UserContextKey, "test-user")
	ctx = context.WithValue(ctx, domain.AuthContextKey, "test-auth")

	err := pub.Publish(ctx, "exchange", "routing.key", map[string]string{"foo": "bar"})
	require.NoError(t, err)

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
