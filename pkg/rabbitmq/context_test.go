package rabbitmq

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestInjectContextHeaders(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected amqp.Table
	}{
		{
			name: "All headers present",
			ctx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, Key(HeaderCorrelationID), "12345")
				ctx = context.WithValue(ctx, Key(HeaderUser), "test-user")
				ctx = context.WithValue(ctx, Key(HeaderAuthorization), "Bearer token")
				return ctx
			}(),
			expected: amqp.Table{
				HeaderCorrelationID: "12345",
				HeaderUser:          "test-user",
				HeaderAuthorization: "Bearer token",
			},
		},
		{
			name:     "No headers",
			ctx:      context.Background(),
			expected: amqp.Table{},
		},
		{
			name: "Partial headers",
			ctx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, Key(HeaderCorrelationID), "abcde")
				return ctx
			}(),
			expected: amqp.Table{
				HeaderCorrelationID: "abcde",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectContextHeaders(tt.ctx)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestExtractContextHeaders(t *testing.T) {
	tests := []struct {
		name         string
		headers      amqp.Table
		expectedKeys map[string]string
	}{
		{
			name: "All headers present",
			headers: amqp.Table{
				HeaderCorrelationID: "12345",
				HeaderUser:          "test-user",
				HeaderAuthorization: "Bearer token",
			},
			expectedKeys: map[string]string{
				HeaderCorrelationID: "12345",
				HeaderUser:          "test-user",
				HeaderAuthorization: "Bearer token",
			},
		},
		{
			name:         "No headers",
			headers:      nil,
			expectedKeys: map[string]string{},
		},
		{
			name:         "Empty headers",
			headers:      amqp.Table{},
			expectedKeys: map[string]string{},
		},
		{
			name: "Partial headers",
			headers: amqp.Table{
				HeaderCorrelationID: "abcde",
			},
			expectedKeys: map[string]string{
				HeaderCorrelationID: "abcde",
				HeaderUser:          "", // missing
				HeaderAuthorization: "", // missing
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := extractContextHeaders(tt.headers)

			for k, want := range tt.expectedKeys {
				gotVal := ctx.Value(Key(k))
				if want == "" {
					assert.Nil(t, gotVal, "expected key %s to be missing/nil", k)
				} else {
					assert.Equal(t, want, gotVal, "expected key %s = %v", k, want)
				}
			}
		})
	}
}
