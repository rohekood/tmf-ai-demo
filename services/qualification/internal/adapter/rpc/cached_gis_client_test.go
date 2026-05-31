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
	err    error
}

func (m *mockNextGIS) CheckPolygon(ctx context.Context, addr domain.Address) (bool, error) {
	m.called++
	return m.result, m.err
}

func TestCachedGISClient_CheckPolygon(t *testing.T) {
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

	cachedClient := NewCachedGISClient(mock, rdb, logger, "qualification:")
	cachedClient.ttl = 1 * time.Hour

	ctx := context.Background()
	addr := domain.Address{Zip: "1000", City: "Berlin", Street: "Main", Number: "1"}

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

	mr.SetError("redis failure")
	exists, err = cachedClient.CheckPolygon(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error during fallback: %v", err)
	}
	if !exists {
		t.Error("expected exists=true")
	}
	if mock.called != 2 {
		t.Errorf("expected source called 2 times (fallback), got %d", mock.called)
	}

	mr.SetError("")

	mock.err = context.DeadlineExceeded
	addrError := domain.Address{Zip: "error"}
	_, err = cachedClient.CheckPolygon(ctx, addrError)
	if err != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded, got %v", err)
	}
	mock.err = nil

	addrFalse := domain.Address{Zip: "false"}
	mock.result = false
	exists, err = cachedClient.CheckPolygon(ctx, addrFalse)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected exists=false")
	}
}
