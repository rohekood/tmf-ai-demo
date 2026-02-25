package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"tmf/services/product-catalog-management/internal/adapter/repository"
)

func TestTransactionManager_Run(t *testing.T) {
	tm := repository.NewTransactionManager(sharedDB)
	
	err := tm.Run(context.Background(), func(ctx context.Context) error {
		tx := repository.GetTx(ctx, sharedDB)
		assert.NotNil(t, tx)
		return nil
	})
	assert.NoError(t, err)

	expectedErr := errors.New("simulated error")
	err = tm.Run(context.Background(), func(ctx context.Context) error {
		return expectedErr
	})
	assert.Equal(t, expectedErr, err)
}

func TestNoOpTransactionManager_Run(t *testing.T) {
	tm := &repository.NoOpTransactionManager{}
	err := tm.Run(context.Background(), func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, err)
}

func TestOutboxModel(t *testing.T) {
	model := &repository.OutboxEventModel{}
	err := model.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, model.ID)
	assert.Equal(t, repository.StatusPending, model.Status)
	assert.Equal(t, "outbox_events", model.TableName())
}
