package observability

import (
	"time"

	"gorm.io/gorm"
)

const dbCallbackStartKey = "recova.observability.db.started_at"

// RegisterDatabaseMetrics registers GORM callbacks for DB latency metrics.
func RegisterDatabaseMetrics(db *gorm.DB, recorder *Recorder) error {
	if db == nil || recorder == nil {
		return nil
	}

	before := func(tx *gorm.DB) {
		tx.InstanceSet(dbCallbackStartKey, time.Now())
	}
	after := func(operation string) func(*gorm.DB) {
		return func(tx *gorm.DB) {
			startedAt := time.Now()
			if value, ok := tx.InstanceGet(dbCallbackStartKey); ok {
				if parsed, castOK := value.(time.Time); castOK {
					startedAt = parsed
				}
			}

			duration := time.Since(startedAt)
			table := "unknown"
			if tx != nil && tx.Statement != nil && tx.Statement.Table != "" {
				table = tx.Statement.Table
			}
			recorder.RecordDBOperation(operation, table, duration, tx.Error)
		}
	}

	db.Callback().Create().Before("*").Remove("recova:obs:before_create")
	if err := db.Callback().Create().Before("*").Register("recova:obs:before_create", before); err != nil {
		return err
	}
	db.Callback().Create().After("*").Remove("recova:obs:after_create")
	if err := db.Callback().Create().After("*").Register("recova:obs:after_create", after("create")); err != nil {
		return err
	}

	db.Callback().Query().Before("*").Remove("recova:obs:before_query")
	if err := db.Callback().Query().Before("*").Register("recova:obs:before_query", before); err != nil {
		return err
	}
	db.Callback().Query().After("*").Remove("recova:obs:after_query")
	if err := db.Callback().Query().After("*").Register("recova:obs:after_query", after("query")); err != nil {
		return err
	}

	db.Callback().Update().Before("*").Remove("recova:obs:before_update")
	if err := db.Callback().Update().Before("*").Register("recova:obs:before_update", before); err != nil {
		return err
	}
	db.Callback().Update().After("*").Remove("recova:obs:after_update")
	if err := db.Callback().Update().After("*").Register("recova:obs:after_update", after("update")); err != nil {
		return err
	}

	db.Callback().Delete().Before("*").Remove("recova:obs:before_delete")
	if err := db.Callback().Delete().Before("*").Register("recova:obs:before_delete", before); err != nil {
		return err
	}
	db.Callback().Delete().After("*").Remove("recova:obs:after_delete")
	if err := db.Callback().Delete().After("*").Register("recova:obs:after_delete", after("delete")); err != nil {
		return err
	}

	db.Callback().Row().Before("*").Remove("recova:obs:before_row")
	if err := db.Callback().Row().Before("*").Register("recova:obs:before_row", before); err != nil {
		return err
	}
	db.Callback().Row().After("*").Remove("recova:obs:after_row")
	if err := db.Callback().Row().After("*").Register("recova:obs:after_row", after("row")); err != nil {
		return err
	}

	db.Callback().Raw().Before("*").Remove("recova:obs:before_raw")
	if err := db.Callback().Raw().Before("*").Register("recova:obs:before_raw", before); err != nil {
		return err
	}
	db.Callback().Raw().After("*").Remove("recova:obs:after_raw")
	if err := db.Callback().Raw().After("*").Register("recova:obs:after_raw", after("raw")); err != nil {
		return err
	}

	return nil
}
