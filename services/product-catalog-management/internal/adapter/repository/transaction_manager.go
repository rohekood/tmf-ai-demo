package repository

import (
	"context"

	"gorm.io/gorm"
)

// TransactionKey is the key type for storing transaction in context
type TransactionKey struct{}

// TxKey is the exported key for context - exposed so other packages can potentially inspect it if needed,
// but primarily for internal use by TransactionManager and Repositories.
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
// If the function returns an error, the transaction is rolled back.
// If the function returns nil, the transaction is committed.
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
// This helper is useful for repositories to decide whether to use an existing tx or the base db connection.
func GetTx(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(TxKey).(*gorm.DB)
	if ok && tx != nil {
		return tx
	}
	return defaultDB.WithContext(ctx)
}
