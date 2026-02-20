package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"tmf/services/customer-management/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutboxPublisher_Publish(t *testing.T) {
	if sharedDB == nil {
		t.Fatal("Shared DB not initialized")
	}
	// Truncate outbox_events
	sharedDB.Exec("TRUNCATE TABLE outbox_events")

	repo := NewOutboxRepository(sharedDB)
	pub := NewOutboxPublisher(repo)
	ctx := context.Background()

	// 1. Publish simple event
	payload := map[string]string{"foo": "bar"}
	err := pub.Publish(ctx, "test.routing.key", payload)
	require.NoError(t, err)

	// Verify DB
	var events []domain.OutboxEvent
	err = sharedDB.Find(&events).Error
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "test.routing.key", events[0].RoutingKey)
	assert.Equal(t, "PENDING", events[0].Status)

	var savedPayload map[string]string
	err = json.Unmarshal(events[0].Payload, &savedPayload)
	require.NoError(t, err)
	assert.Equal(t, "bar", savedPayload["foo"])

	// 2. Publish with Context Headers
	ctxWithUser := context.WithValue(ctx, domain.UserContextKey, "user-123")
	ctxWithUser = context.WithValue(ctxWithUser, domain.AuthContextKey, "Bearer token")

	err = pub.Publish(ctxWithUser, "test.headers", payload)
	require.NoError(t, err)

	// Check latest event headers (re-fetch)
	var latest domain.OutboxEvent
	err = sharedDB.Order("created_at desc").First(&latest).Error
	require.NoError(t, err)

	assert.Equal(t, "test.headers", latest.RoutingKey)

	var headers map[string]string
	err = json.Unmarshal(latest.Headers, &headers)
	require.NoError(t, err)
	assert.Equal(t, "user-123", headers["user"])
	assert.Equal(t, "Bearer token", headers["Authorization"])
}
