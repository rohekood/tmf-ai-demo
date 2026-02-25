package telemetry

import (
	"context"
	"testing"
)

func TestInitTracer(t *testing.T) {
	// Simple test to initialize and shutdown the tracer
	shutdown, err := InitTracer("test-service")
	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown function should not be nil")
	}

	err = shutdown(context.Background())
	if err != nil {
		t.Errorf("shutdown failed: %v", err)
	}
}
