package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// injectContextHeaders injects values from context into AMQP headers
func injectContextHeaders(ctx context.Context) amqp.Table {
	headers := amqp.Table{}

	// Standard keys used in application
	if val, ok := ctx.Value(Key(HeaderCorrelationID)).(string); ok {
		headers[HeaderCorrelationID] = val
	}
	if val, ok := ctx.Value(ContextKeyCorrelationID).(string); ok {
		// Fallback or override
		headers[HeaderCorrelationID] = val
	}
	if val, ok := ctx.Value(Key(HeaderUserID)).(string); ok {
		headers[HeaderUserID] = val
	}

	// Additional keys to satisfy tests
	if val, ok := ctx.Value(ContextKeyUser).(string); ok {
		headers[HeaderUser] = val
	}
	if val, ok := ctx.Value(Key(HeaderAuthorization)).(string); ok {
		headers[HeaderAuthorization] = val
	}
	// Fallback for services using string key "authorization"
	if val, ok := ctx.Value("authorization").(string); ok {
		headers[HeaderAuthorization] = val
	}

	return headers
}

// extractContextHeaders extracts AMQP headers into context
func extractContextHeaders(headers amqp.Table) context.Context {
	ctx := context.Background()
	if headers == nil {
		return ctx
	}

	for k, v := range headers {
		if val, ok := v.(string); ok {
			ctx = context.WithValue(ctx, Key(k), val)
		}
	}
	return ctx
}
