package database

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// ColumnInfo holds metadata for a database column retrieved from information_schema.
type ColumnInfo struct {
	TableName  string `gorm:"column:table_name"`
	ColumnName string `gorm:"column:column_name"`
	DataType   string `gorm:"column:data_type"`
	UDTName    string `gorm:"column:udt_name"`
	IsNullable string `gorm:"column:is_nullable"`
}

// DriftReport summarizes schema differences between runtime models and database.
type DriftReport struct {
	MissingTables   []string            `json:"missing_tables,omitempty"`
	UntrackedTables []string            `json:"untracked_tables,omitempty"`
	MissingColumns  map[string][]string `json:"missing_columns,omitempty"`
	TypeMismatches  []string            `json:"type_mismatches,omitempty"`
}

// HasDrift returns true if any drift is detected.
func (r *DriftReport) HasDrift() bool {
	if r == nil {
		return false
	}
	return len(r.MissingTables) > 0 ||
		len(r.UntrackedTables) > 0 ||
		len(r.MissingColumns) > 0 ||
		len(r.TypeMismatches) > 0
}

// Error returns a formatted error string if drift exists, or empty string if clean.
func (r *DriftReport) Error() string {
	if r == nil || !r.HasDrift() {
		return ""
	}

	var b strings.Builder
	b.WriteString("schema drift detected:")

	if len(r.MissingTables) > 0 {
		b.WriteString(fmt.Sprintf("\n  missing tables: %s", strings.Join(r.MissingTables, ", ")))
	}
	if len(r.UntrackedTables) > 0 {
		b.WriteString(fmt.Sprintf("\n  untracked tables: %s", strings.Join(r.UntrackedTables, ", ")))
	}
	if len(r.MissingColumns) > 0 {
		var tables []string
		for tbl := range r.MissingColumns {
			tables = append(tables, tbl)
		}
		sort.Strings(tables)
		for _, tbl := range tables {
			cols := r.MissingColumns[tbl]
			b.WriteString(fmt.Sprintf("\n  table %q missing columns: %s", tbl, strings.Join(cols, ", ")))
		}
	}
	if len(r.TypeMismatches) > 0 {
		for _, m := range r.TypeMismatches {
			b.WriteString(fmt.Sprintf("\n  type mismatch: %s", m))
		}
	}

	return b.String()
}

// IgnoredSystemTables lists tables managed by migration tooling or system extensions.
var IgnoredSystemTables = map[string]struct{}{
	"schema_migrations": {},
}

// CheckSchemaDrift compares registered domain models against current PostgreSQL schema.
func CheckSchemaDrift(ctx context.Context, db *gorm.DB, registeredModels ...any) (*DriftReport, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}

	if len(registeredModels) == 0 {
		registeredModels = models.RegisteredModels()
	}

	// 1. Fetch all public tables from information_schema
	const tablesQuery = `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY table_name;`

	var dbTables []string
	if err := db.WithContext(ctx).Raw(tablesQuery).Scan(&dbTables).Error; err != nil {
		return nil, fmt.Errorf("query information_schema.tables: %w", err)
	}

	dbTableMap := make(map[string]struct{}, len(dbTables))
	for _, tbl := range dbTables {
		dbTableMap[tbl] = struct{}{}
	}

	// 2. Fetch all columns from information_schema
	const columnsQuery = `
SELECT table_name, column_name, data_type, udt_name, is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
ORDER BY table_name, ordinal_position;`

	var columns []ColumnInfo
	if err := db.WithContext(ctx).Raw(columnsQuery).Scan(&columns).Error; err != nil {
		return nil, fmt.Errorf("query information_schema.columns: %w", err)
	}

	dbColumnsByTable := make(map[string]map[string]ColumnInfo)
	for _, col := range columns {
		if _, ok := dbColumnsByTable[col.TableName]; !ok {
			dbColumnsByTable[col.TableName] = make(map[string]ColumnInfo)
		}
		dbColumnsByTable[col.TableName][col.ColumnName] = col
	}

	report := &DriftReport{
		MissingColumns: make(map[string][]string),
	}

	modelTableMap := make(map[string]struct{}, len(registeredModels))

	// 3. Inspect each registered model
	for _, model := range registeredModels {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(model); err != nil {
			return nil, fmt.Errorf("parse model schema %T: %w", model, err)
		}

		tableName := stmt.Schema.Table
		modelTableMap[tableName] = struct{}{}

		if _, exists := dbTableMap[tableName]; !exists {
			report.MissingTables = append(report.MissingTables, tableName)
			continue
		}

		tableCols := dbColumnsByTable[tableName]

		// Check each field in model
		for _, field := range stmt.Schema.Fields {
			if field.DBName == "" || field.DBName == "-" {
				continue
			}

			colInfo, colExists := tableCols[field.DBName]
			if !colExists {
				report.MissingColumns[tableName] = append(report.MissingColumns[tableName], field.DBName)
				continue
			}

			// Validate basic type compatibility
			if mismatch := validateTypeCompatibility(tableName, field.DBName, field, colInfo); mismatch != "" {
				report.TypeMismatches = append(report.TypeMismatches, mismatch)
			}
		}
	}

	// 4. Check for untracked tables in database
	for _, tbl := range dbTables {
		if _, ignored := IgnoredSystemTables[tbl]; ignored {
			continue
		}
		if _, known := modelTableMap[tbl]; !known {
			report.UntrackedTables = append(report.UntrackedTables, tbl)
		}
	}

	sort.Strings(report.MissingTables)
	sort.Strings(report.UntrackedTables)
	sort.Strings(report.TypeMismatches)
	for _, cols := range report.MissingColumns {
		sort.Strings(cols)
	}

	return report, nil
}

// validateTypeCompatibility checks basic type affinity between GORM field and PostgreSQL column.
func validateTypeCompatibility(tableName, colName string, field *schema.Field, col ColumnInfo) string {
	if field == nil {
		return ""
	}

	pgType := strings.ToLower(col.DataType)
	pgUDT := strings.ToLower(col.UDTName)

	switch field.DataType {
	case schema.Bool:
		if pgType != "boolean" {
			return fmt.Sprintf("%s.%s: Go bool mapped to incompatible DB type %s", tableName, colName, pgType)
		}
	case schema.Time:
		validTimeTypes := map[string]bool{
			"timestamp with time zone":    true,
			"timestamp without time zone": true,
			"date":                        true,
			"time without time zone":      true,
			"time with time zone":         true,
		}
		if !validTimeTypes[pgType] {
			return fmt.Sprintf("%s.%s: Go time.Time mapped to incompatible DB type %s", tableName, colName, pgType)
		}
	case schema.Int, schema.Uint:
		validIntTypes := map[string]bool{
			"smallint": true,
			"integer":  true,
			"bigint":   true,
			"numeric":  true,
		}
		if !validIntTypes[pgType] {
			return fmt.Sprintf("%s.%s: Go integer mapped to incompatible DB type %s", tableName, colName, pgType)
		}
	case schema.Float:
		validFloatTypes := map[string]bool{
			"numeric":          true,
			"real":             true,
			"double precision": true,
		}
		if !validFloatTypes[pgType] {
			return fmt.Sprintf("%s.%s: Go float mapped to incompatible DB type %s", tableName, colName, pgType)
		}
	case schema.Bytes:
		validByteTypes := map[string]bool{
			"bytea": true,
			"jsonb": true,
			"json":  true,
		}
		if !validByteTypes[pgType] {
			return fmt.Sprintf("%s.%s: Go bytes mapped to incompatible DB type %s", tableName, colName, pgType)
		}
	}

	// Array checking for pq.StringArray
	if field.FieldType != nil && field.FieldType.Kind() == reflect.Slice {
		if field.FieldType.Elem().Kind() == reflect.String {
			if pgType != "array" || (pgUDT != "_text" && pgUDT != "_varchar") {
				return fmt.Sprintf("%s.%s: Go string array mapped to non-array DB type %s (%s)", tableName, colName, pgType, pgUDT)
			}
		}
	}

	return ""
}
