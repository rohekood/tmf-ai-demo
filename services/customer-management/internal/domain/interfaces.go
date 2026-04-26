package domain

import (
	"context"
)

type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type EventPublisher interface {
	Publish(ctx context.Context, routingKey string, payload any) error
}
