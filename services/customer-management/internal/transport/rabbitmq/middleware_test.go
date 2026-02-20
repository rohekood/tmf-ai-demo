package rabbitmq

import (
	"context"
	"testing"
	"tmf/services/customer-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	middleware := AuthMiddleware()

	t.Run("with user header", func(t *testing.T) {
		d := amqp.Delivery{
			Headers: amqp.Table{
				"user": "test-user-123",
			},
		}

		handler := middleware(func(ctx context.Context, d amqp.Delivery) error {
			userID := ctx.Value(domain.UserContextKey)
			assert.Equal(t, "test-user-123", userID)
			return nil
		})

		err := handler(context.Background(), d)
		assert.NoError(t, err)
	})

	t.Run("without user header", func(t *testing.T) {
		d := amqp.Delivery{
			Headers: amqp.Table{},
		}

		handler := middleware(func(ctx context.Context, d amqp.Delivery) error {
			userID := ctx.Value(domain.UserContextKey)
			assert.Nil(t, userID)
			return nil
		})

		err := handler(context.Background(), d)
		assert.NoError(t, err)
	})
}

func TestTracingMiddleware(t *testing.T) {
	middleware := TracingMiddleware("test-service")

	handler := func(ctx context.Context, d amqp.Delivery) error {
		return nil
	}

	// Just verify it doesn't panic and calls next
	err := middleware(handler)(context.Background(), amqp.Delivery{CorrelationId: "corr-1"})
	assert.NoError(t, err)
}

func TestAMQPHeaderCarrier(t *testing.T) {
	carrier := make(amqpHeaderCarrier)

	// Test Set
	carrier.Set("key1", "value1")
	carrier.Set("key2", "123") // Header values are usually strings or handled as interface{}

	// Test Get
	assert.Equal(t, "value1", carrier.Get("key1"))
	assert.Equal(t, "123", carrier.Get("key2"))
	assert.Equal(t, "", carrier.Get("key3"))

	// Test Keys
	keys := carrier.Keys()
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
	assert.Len(t, keys, 2)
}

func TestJWTMiddleware(t *testing.T) {
	middleware := JWTMiddleware()

	t.Run("with authorization header", func(t *testing.T) {
		d := amqp.Delivery{
			Headers: amqp.Table{
				"Authorization": "Bearer token123",
			},
		}

		handler := func(ctx context.Context, d amqp.Delivery) error {
			authVal := ctx.Value("authorization")
			assert.Equal(t, "Bearer token123", authVal)
			return nil
		}

		err := middleware(handler)(context.Background(), d)
		assert.NoError(t, err)
	})
}

func TestChain(t *testing.T) {
	var calls []string

	m1 := func(next func(ctx context.Context, d amqp.Delivery) error) func(ctx context.Context, d amqp.Delivery) error {
		return func(ctx context.Context, d amqp.Delivery) error {
			calls = append(calls, "m1-start")
			err := next(ctx, d)
			calls = append(calls, "m1-end")
			return err
		}
	}

	m2 := func(next func(ctx context.Context, d amqp.Delivery) error) func(ctx context.Context, d amqp.Delivery) error {
		return func(ctx context.Context, d amqp.Delivery) error {
			calls = append(calls, "m2-start")
			err := next(ctx, d)
			calls = append(calls, "m2-end")
			return err
		}
	}

	handler := func(ctx context.Context, d amqp.Delivery) error {
		calls = append(calls, "handler")
		return nil
	}

	chained := Chain(handler, m1, m2)
	_ = chained(context.Background(), amqp.Delivery{})

	expected := []string{"m1-start", "m2-start", "handler", "m2-end", "m1-end"}
	assert.Equal(t, expected, calls)
}
