package database

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/lib/pq"
	"gorm.io/gorm/schema"
)

func TestDriftReport_HasDrift(t *testing.T) {
	var nilReport *DriftReport
	if nilReport.HasDrift() {
		t.Fatal("expected nil report to have no drift")
	}
	if nilReport.Error() != "" {
		t.Fatalf("expected nil report error to be empty, got: %q", nilReport.Error())
	}

	clean := &DriftReport{}
	if clean.HasDrift() {
		t.Fatal("expected no drift on clean report")
	}
	if clean.Error() != "" {
		t.Fatalf("expected empty error on clean report, got: %q", clean.Error())
	}

	withMissingTable := &DriftReport{MissingTables: []string{"users"}}
	if !withMissingTable.HasDrift() {
		t.Fatal("expected drift when table missing")
	}

	withUntrackedTable := &DriftReport{UntrackedTables: []string{"legacy_data"}}
	if !withUntrackedTable.HasDrift() {
		t.Fatal("expected drift when untracked table present")
	}

	withMissingCol := &DriftReport{MissingColumns: map[string][]string{"users": {"phone"}}}
	if !withMissingCol.HasDrift() {
		t.Fatal("expected drift when column missing")
	}

	withMismatch := &DriftReport{TypeMismatches: []string{"users.id: mismatch"}}
	if !withMismatch.HasDrift() {
		t.Fatal("expected drift when type mismatch present")
	}
}

func TestDriftReport_ErrorFormatting(t *testing.T) {
	report := &DriftReport{
		MissingTables:   []string{"audit_logs"},
		UntrackedTables: []string{"temp_scratch"},
		MissingColumns:  map[string][]string{"users": {"avatar_id"}},
		TypeMismatches:  []string{"users.age: Go integer mapped to incompatible DB type boolean"},
	}

	errStr := report.Error()
	if errStr == "" {
		t.Fatal("expected non-empty error string")
	}

	expectedSubstrings := []string{
		"schema drift detected:",
		"missing tables: audit_logs",
		"untracked tables: temp_scratch",
		`table "users" missing columns: avatar_id`,
		"users.age: Go integer mapped to incompatible DB type boolean",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(errStr, sub) {
			t.Fatalf("expected error string to contain %q, got:\n%s", sub, errStr)
		}
	}
}

func TestCheckSchemaDrift_NilDatabase_ReturnsError(t *testing.T) {
	_, err := CheckSchemaDrift(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when db is nil")
	}
}

func TestValidateTypeCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		field     *schema.Field
		col       ColumnInfo
		wantError bool
	}{
		{
			name: "bool to boolean match",
			field: &schema.Field{
				DataType: schema.Bool,
			},
			col: ColumnInfo{
				DataType: "boolean",
			},
			wantError: false,
		},
		{
			name: "bool to integer mismatch",
			field: &schema.Field{
				DataType: schema.Bool,
			},
			col: ColumnInfo{
				DataType: "integer",
			},
			wantError: true,
		},
		{
			name: "time to timestamptz match",
			field: &schema.Field{
				DataType: schema.Time,
			},
			col: ColumnInfo{
				DataType: "timestamp with time zone",
			},
			wantError: false,
		},
		{
			name: "time to date match",
			field: &schema.Field{
				DataType: schema.Time,
			},
			col: ColumnInfo{
				DataType: "date",
			},
			wantError: false,
		},
		{
			name: "time to boolean mismatch",
			field: &schema.Field{
				DataType: schema.Time,
			},
			col: ColumnInfo{
				DataType: "boolean",
			},
			wantError: true,
		},
		{
			name: "integer to integer match",
			field: &schema.Field{
				DataType: schema.Int,
			},
			col: ColumnInfo{
				DataType: "integer",
			},
			wantError: false,
		},
		{
			name: "integer to smallint match",
			field: &schema.Field{
				DataType: schema.Int,
			},
			col: ColumnInfo{
				DataType: "smallint",
			},
			wantError: false,
		},
		{
			name: "integer to text mismatch",
			field: &schema.Field{
				DataType: schema.Int,
			},
			col: ColumnInfo{
				DataType: "text",
			},
			wantError: true,
		},
		{
			name: "bytes to jsonb match",
			field: &schema.Field{
				DataType: schema.Bytes,
			},
			col: ColumnInfo{
				DataType: "jsonb",
			},
			wantError: false,
		},
		{
			name: "string array to text array match",
			field: &schema.Field{
				FieldType: reflect.TypeOf(pq.StringArray{}),
			},
			col: ColumnInfo{
				DataType: "ARRAY",
				UDTName:  "_text",
			},
			wantError: false,
		},
		{
			name: "string array to integer array mismatch",
			field: &schema.Field{
				FieldType: reflect.TypeOf(pq.StringArray{}),
			},
			col: ColumnInfo{
				DataType: "ARRAY",
				UDTName:  "_int4",
			},
			wantError: true,
		},
		{
			name: "string array to scalar text mismatch",
			field: &schema.Field{
				FieldType: reflect.TypeOf(pq.StringArray{}),
			},
			col: ColumnInfo{
				DataType: "text",
				UDTName:  "text",
			},
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := validateTypeCompatibility("test_table", "test_col", tc.field, tc.col)
			if tc.wantError && res == "" {
				t.Fatalf("expected mismatch error, got empty")
			}
			if !tc.wantError && res != "" {
				t.Fatalf("expected no mismatch error, got: %s", res)
			}
		})
	}
}
