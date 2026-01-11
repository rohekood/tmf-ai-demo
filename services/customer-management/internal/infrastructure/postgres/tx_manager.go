package postgres

import (
	"context"
	"tmf/services/customer-management/internal/domain"

	"gorm.io/gorm"
)

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.Transaction(func(tx *gorm.DB) error {
		// Set audit context if user is present
		if userID, ok := ctx.Value(domain.UserContextKey).(string); ok {
			if err := tx.Exec("SELECT set_config('app.current_user', ?, true)", userID).Error; err != nil {
				return err
			}
		}

		// Inject tx into context
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}
