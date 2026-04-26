package domain

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, exchange, routingKey string, event any) error
}

type TransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
