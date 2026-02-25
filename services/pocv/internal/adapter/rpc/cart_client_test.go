package rpc

import (
	"context"
	"errors"
	"testing"
)

type mockRPC struct {
	request func(ctx context.Context, routingKey string, payload interface{}) ([]byte, error)
}

func (m *mockRPC) Request(ctx context.Context, routingKey string, payload interface{}) ([]byte, error) {
	if m.request != nil {
		return m.request(ctx, routingKey, payload)
	}
	return nil, nil
}

func TestGetCart(t *testing.T) {
	tests := []struct {
		name        string
		rpcMock     *mockRPC
		expectError bool
	}{
		{
			name: "Success",
			rpcMock: &mockRPC{
				request: func(ctx context.Context, routingKey string, payload interface{}) ([]byte, error) {
					return []byte(`{"items": []}`), nil
				},
			},
			expectError: false,
		},
		{
			name: "RPC Error",
			rpcMock: &mockRPC{
				request: func(ctx context.Context, routingKey string, payload interface{}) ([]byte, error) {
					return nil, errors.New("rpc failed")
				},
			},
			expectError: true,
		},
		{
			name: "Unmarshal Error",
			rpcMock: &mockRPC{
				request: func(ctx context.Context, routingKey string, payload interface{}) ([]byte, error) {
					return []byte(`invalid json`), nil
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewCartClient(tt.rpcMock)
			_, err := client.GetCart(context.Background(), "cart-1")
			if (err != nil) != tt.expectError {
				t.Errorf("GetCart() error = %v, expectError = %v", err, tt.expectError)
			}
		})
	}
}
