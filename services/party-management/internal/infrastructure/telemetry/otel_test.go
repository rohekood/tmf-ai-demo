package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitTracer_Success(t *testing.T) {
	// Call InitTracer with a dummy service name
	shutdown, err := InitTracer("test-party-service")
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	// Since we are not doing a real exporter integration test that sends spans,
	// just verify it shuts down gracefully.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = shutdown(ctx)
	assert.NoError(t, err)
}
