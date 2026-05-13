package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// WithTransaction executes fn inside one database transaction and returns callback error directly.
func WithTransaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	if db == nil {
		return fmt.Errorf("database handle is required")
	}

	if fn == nil {
		return fmt.Errorf("transaction callback is required")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
