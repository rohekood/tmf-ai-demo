package domain

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, event interface{}) error
}

type TransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
