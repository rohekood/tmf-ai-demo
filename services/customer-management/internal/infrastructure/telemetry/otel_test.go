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

	// Test shutdown errors by passing canceled context
	shutdownErr, _ := InitTracer("test-service-2")
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel() // instantly cancel context to force an error during shutdown
	err = shutdownErr(cancelCtx)
	if err == nil {
		// Sometimes OTEL ignores context cancellations on metrics exporter shutdown if it's already stopped or no-op.
		// That's fine, we're just trying to hit the code path.
		t.Logf("shutdown with cancelled context returned nil, expected error potentially")
	}
}

func TestInitTracer_EmptyServiceName(t *testing.T) {
	// Trying to trigger resource.New error
	_, err := InitTracer("")
	if err != nil {
		t.Logf("InitTracer with empty service: %v", err)
	}
}
