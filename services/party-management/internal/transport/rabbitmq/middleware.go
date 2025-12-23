package rabbitmq

import (
	"context"
	"fmt"
	"tmf/services/party-management/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

// Middleware defines a function that wraps a message handler.
type Middleware func(next func(ctx context.Context, d amqp.Delivery) error) func(ctx context.Context, d amqp.Delivery) error

// TracingMiddleware extracts trace context from AMQP headers and starts a span.
func TracingMiddleware(serviceName string) Middleware {
	return func(next func(ctx context.Context, d amqp.Delivery) error) func(ctx context.Context, d amqp.Delivery) error {
		return func(ctx context.Context, d amqp.Delivery) error {
			// Extract context from headers
			propagator := otel.GetTextMapPropagator()
			ctx = propagator.Extract(ctx, amqpHeaderCarrier(d.Headers))

			tracer := otel.Tracer(serviceName)
			ctx, span := tracer.Start(ctx, fmt.Sprintf("rabbitmq.%s", d.RoutingKey))
			defer span.End()

			return next(ctx, d)
		}
	}
}

// amqpHeaderCarrier maps AMQP headers to OTel carrier interface.
type amqpHeaderCarrier map[string]interface{}

func (c amqpHeaderCarrier) Get(key string) string {
	if v, ok := c[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c amqpHeaderCarrier) Set(key string, value string) {
	c[key] = value
}

func (c amqpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

// AuthMiddleware extracts a user ID from the 'user' header.
func AuthMiddleware() Middleware {
	return func(next func(ctx context.Context, d amqp.Delivery) error) func(ctx context.Context, d amqp.Delivery) error {
		return func(ctx context.Context, d amqp.Delivery) error {
			user, ok := d.Headers["user"].(string)
			if !ok || user == "" {
				return next(ctx, d)
			}

			ctx = context.WithValue(ctx, domain.UserContextKey, user)
			return next(ctx, d)
		}
	}
}

// Chain applies a list of middlewares to a handler.
func Chain(handler func(ctx context.Context, d amqp.Delivery) error, middlewares ...Middleware) func(ctx context.Context, d amqp.Delivery) error {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
