package postgres

import (
	"context"
	"testing"
	"time"
	"tmf/services/customer-management/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutboxRepository_FetchAndMark(t *testing.T) {
	if sharedDB == nil {
		t.Fatal("Shared DB not initialized")
	}
	sharedDB.Exec("TRUNCATE TABLE outbox_events")

	repo := NewOutboxRepository(sharedDB)
	ctx := context.Background()

	// Seed events
	event1 := &domain.OutboxEvent{
		ID:         uuid.New().String(),
		RoutingKey: "test.1",
		Payload:    []byte("{}"),
		Status:     "PENDING",
		CreatedAt:  time.Now().Add(-1 * time.Minute),
	}
	event2 := &domain.OutboxEvent{
		ID:         uuid.New().String(),
		RoutingKey: "test.2",
		Payload:    []byte("{}"),
		Status:     "PUBLISHED", // Should not be fetched
		CreatedAt:  time.Now(),
	}

	require.NoError(t, repo.Save(ctx, event1))
	require.NoError(t, repo.Save(ctx, event2))

	// Test FetchPending
	fetched, err := repo.FetchPending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, fetched, 1)
	assert.Equal(t, event1.ID, fetched[0].ID)

	// Test MarkAsProcessed
	err = repo.MarkAsProcessed(ctx, event1.ID)
	require.NoError(t, err)

	// Verify status updated
	var updated domain.OutboxEvent
	err = sharedDB.First(&updated, "id = ?", event1.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "PUBLISHED", updated.Status)
	assert.NotNil(t, updated.ProcessedAt)

	// Verify FetchPending returns nothing now
	fetchedAgain, err := repo.FetchPending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, fetchedAgain, 0)
}
