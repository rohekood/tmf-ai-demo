package postgres

import (
	"context"

	"gorm.io/gorm"
)

// TransactionKey is the key type for storing transaction in context
type TransactionKey struct{}

// TxKey is the exported key for context
var TxKey = TransactionKey{}

// TransactionManager interface defines the contract for transaction management
type TransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

type GormTransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

// Run executes the given function within a transaction.
func (tm *GormTransactionManager) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.Transaction(func(tx *gorm.DB) error {
		// Store the transaction in the context
		txCtx := context.WithValue(ctx, TxKey, tx)
		return fn(txCtx)
	})
}

// GetTx retrieves the transaction from the context if it exists,
// otherwise returns the default db passed as argument.
func GetTx(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(TxKey).(*gorm.DB)
	if ok && tx != nil {
		return tx
	}
	// If no transaction in context, use the default DB but attached to context
	return defaultDB.WithContext(ctx)
}
