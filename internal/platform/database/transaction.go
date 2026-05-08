package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// WithTransaction executes fn inside one database transaction and returns callback error directly.
func WithTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	if db == nil {
		return fmt.Errorf("database handle wajib tersedia")
	}

	if fn == nil {
		return fmt.Errorf("callback transaksi wajib tersedia")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
