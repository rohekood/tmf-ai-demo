package repository

import (
	"context"
)

type NoOpTransactionManager struct{}

func (tm *NoOpTransactionManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
