package rpc

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"tmf/services/qualification/internal/core/domain"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type mockNextGIS struct {
	called int
	result bool
}

func (m *mockNextGIS) CheckPolygon(ctx context.Context, addr domain.Address) (bool, error) {
	m.called++
	return m.result, nil
}

func TestCachedGISClient_CheckPolygon(t *testing.T) {
	// Setup Miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mock := &mockNextGIS{result: true}

	// Create Cached Client
	cachedClient := NewCachedGISClient(mock, rdb, logger)
	cachedClient.ttl = 1 * time.Hour // Set TTL

	ctx := context.Background()
	addr := domain.Address{Zip: "1000", City: "Berlin", Street: "Main", Number: "1"}

	// 1. First Call: Should hit Source
	exists, err := cachedClient.CheckPolygon(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
	if mock.called != 1 {
		t.Errorf("expected source called 1 time, got %d", mock.called)
	}

	// 2. Second Call: Should hit Cache
	exists, err = cachedClient.CheckPolygon(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
	if mock.called != 1 {
		t.Errorf("expected source called 1 time (cached), got %d", mock.called)
	}

	// 3. Third Call: Redis Error should Fallback to Source
	mr.SetError("redis failure")
	exists, err = cachedClient.CheckPolygon(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error during fallback: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
	// Should have called source again because cache read failed
	if mock.called != 2 {
		t.Errorf("expected source called 2 times (fallback), got %d", mock.called)
	}
}
