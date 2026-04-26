package rabbitmq

import (
	"context"
	"testing"

	"tmf/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestAmqpHeaderCarrier(t *testing.T) {
	cm := amqpHeaderCarrier(make(map[string]any))
	cm.Set("key1", "val1")
	val := cm.Get("key1")
	assert.Equal(t, "val1", val)

	val2 := cm.Get("unknown")
	assert.Empty(t, val2)

	keys := cm.Keys()
	assert.Len(t, keys, 1)
	assert.Equal(t, "key1", keys[0])
}

func TestTracingMiddleware(t *testing.T) {
	handlerCalled := false
	handler := func(ctx context.Context, d amqp.Delivery) error {
		val := ctx.Value(rabbitmq.ContextKeyCorrelationID)
		if val != nil {
			assert.NotEmpty(t, val)
		}
		handlerCalled = true
		return nil
	}

	mw := TracingMiddleware("test-service")(handler)
	d := amqp.Delivery{CorrelationId: "test-corr-id"}
	err := mw(context.Background(), d)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)

	// Test without correlation ID (should generate uuid)
	_ = mw(context.Background(), amqp.Delivery{})
}

func TestJWTMiddleware(t *testing.T) {
	handlerCalled := false
	handler := func(ctx context.Context, d amqp.Delivery) error {
		handlerCalled = true
		return nil
	}

	mw := JWTMiddleware()(handler)
	d := amqp.Delivery{
		Headers: amqp.Table{
			"Authorization": "Bearer test-jwt",
		},
	}
	err := mw(context.Background(), d)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)

	// Test without header
	d2 := amqp.Delivery{}
	err2 := mw(context.Background(), d2)
	assert.Error(t, err2)
}
