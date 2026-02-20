package rabbitmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListener_GetHandler(t *testing.T) {
	l := &Listener{}
	h := &Handlers{}

	tests := []struct {
		name       string
		routingKey string
		wantValid  bool
	}{
		{"Onboard Customer", "cmd.customer.onboard", true},
		{"Update Customer", "cmd.customer.update", true},
		{"Get Customer", "query.customer.get", false},
		{"Search Customer", "query.customer.search", true},
		{"Delete Customer", "cmd.customer.delete", true},
		{"Unknown Key", "unknown.key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, valid := l.GetHandler(tt.routingKey, h)
			assert.Equal(t, tt.wantValid, valid)
			if tt.wantValid {
				assert.NotNil(t, handler)
			} else {
				assert.Nil(t, handler)
			}
		})
	}
}

func TestListener_Start(t *testing.T) {
	if sharedConn == nil {
		t.Skip("Skipping integration test: sharedConn not initialized")
	}

	l, err := NewListener(sharedConn)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Start in background
	errChan := make(chan error)
	go func() {
		// Mock handlers
		h := &Handlers{}
		errChan <- l.Start(ctx, h)
	}()

	// Give it a moment to start
	// In real test we might want to check if queue exists

	// Stop
	cancel()
	err = <-errChan
	assert.NoError(t, err)
}
