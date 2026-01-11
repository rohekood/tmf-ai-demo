package postgres

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// GetTx returns the transaction from the context if present, otherwise returns the db.
func GetTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db.WithContext(ctx)
}
