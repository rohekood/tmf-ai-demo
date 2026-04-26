package http

import (
	"context"
)

// MockRPCClient is a mock implementation of RPCClient for testing
type MockRPCClient struct {
	CallRPCFunc func(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error)
}

func (m *MockRPCClient) CallRPC(ctx context.Context, exchange, routingKey string, payload any, headers map[string]any) ([]byte, error) {
	if m.CallRPCFunc != nil {
		return m.CallRPCFunc(ctx, exchange, routingKey, payload, headers)
	}
	return nil, nil // Return empty response by default
}
