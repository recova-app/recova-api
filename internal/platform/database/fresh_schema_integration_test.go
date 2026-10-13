package database

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/recova-app/backend-v2/internal/platform/config"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	databaseharness "github.com/recova-app/backend-v2/test/harness/database"
)

func TestIntegration_FreshSchemaMigrationChain_AndDriftCheck(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	ctx := context.Background()

	m, err := migrate.New("file://"+databaseharness.MigrationsPath(t), databaseURL)
	if err != nil {
		t.Fatalf("new migrate instance: %v", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	// Drop all tables to step 0
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrate down to 0: %v", err)
	}

	// Apply migration chain up
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrate up to latest: %v", err)
	}

	// Verify version matches latest migration
	currentVersion, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("get migration version: %v", err)
	}
	if dirty {
		t.Fatal("database migration is marked dirty")
	}

	expectedLatest := databaseharness.LatestMigrationVersion(t)
	if uint64(currentVersion) != expectedLatest {
		t.Fatalf("expected migration version %d, got %d", expectedLatest, currentVersion)
	}

	client, err := Connect(freshTestConfig(databaseURL))
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Check schema drift against all registered domain models
	report, err := CheckSchemaDrift(ctx, client.Gorm())
	if err != nil {
		t.Fatalf("check schema drift: %v", err)
	}
	if report.HasDrift() {
		t.Fatalf("unexpected schema drift:\n%s", report.Error())
	}

	// Rollback smoke: step down 1 migration and re-apply
	if err := m.Steps(-1); err != nil {
		t.Fatalf("step down 1 migration: %v", err)
	}
	if err := m.Steps(1); err != nil {
		t.Fatalf("step up 1 migration: %v", err)
	}

	// Verify schema drift is clean after rollback and re-apply
	reportAfterReapply, err := CheckSchemaDrift(ctx, client.Gorm())
	if err != nil {
		t.Fatalf("check schema drift after reapply: %v", err)
	}
	if reportAfterReapply.HasDrift() {
		t.Fatalf("unexpected schema drift after reapply:\n%s", reportAfterReapply.Error())
	}
}

func TestIntegration_BaselineSeedValidation(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	ctx := context.Background()

	databaseharness.ResetMigrations(t, databaseURL)

	client, err := Connect(freshTestConfig(databaseURL))
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Apply baseline seed
	databaseharness.ApplyBaselineSeed(t, client.Gorm())

	// Validate seeded records
	type countCheck struct {
		model    any
		minCount int64
		label    string
	}

	checks := []countCheck{
		{model: &models.User{}, minCount: 6, label: "users"},
		{model: &models.Profile{}, minCount: 6, label: "profiles"},
		{model: &models.Streak{}, minCount: 1, label: "streaks"},
		{model: &models.CheckIn{}, minCount: 1, label: "check_ins"},
		{model: &models.Journal{}, minCount: 1, label: "journals"},
		{model: &models.CommunityPost{}, minCount: 1, label: "community_posts"},
		{model: &models.CommunityComment{}, minCount: 1, label: "community_comments"},
		{model: &models.Achievement{}, minCount: 1, label: "achievements"},
		{model: &models.EducationContent{}, minCount: 1, label: "education_contents"},
		{model: &models.DailyMotivation{}, minCount: 1, label: "daily_motivations"},
		{model: &models.DailyChallenge{}, minCount: 1, label: "daily_challenges"},
		{model: &models.DailyPhysicalChallenge{}, minCount: 1, label: "daily_physical_challenges"},
	}

	for _, c := range checks {
		var count int64
		if err := client.Gorm().WithContext(ctx).Model(c.model).Count(&count).Error; err != nil {
			t.Fatalf("count %s: %v", c.label, err)
		}
		if count < c.minCount {
			t.Fatalf("expected at least %d %s, got %d", c.minCount, c.label, count)
		}
	}

	// Test seed idempotency (applying a second time must succeed without error)
	databaseharness.ApplyBaselineSeed(t, client.Gorm())
}

type phantomColumnTestModel struct {
	ID            string `gorm:"primaryKey"`
	PhantomColumn string `gorm:"column:phantom_drift_check_column"`
}

func (phantomColumnTestModel) TableName() string {
	return "users"
}

func TestIntegration_SchemaDrift_DetectsMissingColumn(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	ctx := context.Background()

	databaseharness.ResetMigrations(t, databaseURL)

	client, err := Connect(freshTestConfig(databaseURL))
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	report, err := CheckSchemaDrift(ctx, client.Gorm(), &phantomColumnTestModel{})
	if err != nil {
		t.Fatalf("check schema drift: %v", err)
	}

	if !report.HasDrift() {
		t.Fatal("expected schema drift detected for phantom column, but report was clean")
	}

	cols := report.MissingColumns["users"]
	found := false
	for _, c := range cols {
		if c == "phantom_drift_check_column" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected phantom_drift_check_column in missing columns, got: %v", cols)
	}

	if !strings.Contains(report.Error(), "phantom_drift_check_column") {
		t.Fatalf("expected error string to mention phantom column, got:\n%s", report.Error())
	}
}

type phantomTableTestModel struct {
	ID string `gorm:"primaryKey"`
}

func (phantomTableTestModel) TableName() string {
	return "phantom_unmigrated_table"
}

func TestIntegration_SchemaDrift_DetectsMissingTable(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	ctx := context.Background()

	databaseharness.ResetMigrations(t, databaseURL)

	client, err := Connect(freshTestConfig(databaseURL))
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	report, err := CheckSchemaDrift(ctx, client.Gorm(), &phantomTableTestModel{})
	if err != nil {
		t.Fatalf("check schema drift: %v", err)
	}

	if !report.HasDrift() {
		t.Fatal("expected schema drift detected for phantom table, but report was clean")
	}

	found := false
	for _, tbl := range report.MissingTables {
		if tbl == "phantom_unmigrated_table" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected phantom_unmigrated_table in missing tables, got: %v", report.MissingTables)
	}

	if !strings.Contains(report.Error(), "phantom_unmigrated_table") {
		t.Fatalf("expected error string to mention phantom table, got:\n%s", report.Error())
	}
}

func TestIntegration_SchemaDrift_DetectsUntrackedTable(t *testing.T) {
	databaseURL := databaseharness.RequireDatabaseURLFromEnv(t, "RECOVA_DB_INTEGRATION_URL")
	ctx := context.Background()

	databaseharness.ResetMigrations(t, databaseURL)

	client, err := Connect(freshTestConfig(databaseURL))
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	const createTempTable = `CREATE TABLE temp_untracked_drift_table (id uuid PRIMARY KEY);`
	const dropTempTable = `DROP TABLE IF EXISTS temp_untracked_drift_table;`

	if err := client.Gorm().WithContext(ctx).Exec(createTempTable).Error; err != nil {
		t.Fatalf("create temp untracked table: %v", err)
	}
	defer func() {
		_ = client.Gorm().WithContext(ctx).Exec(dropTempTable).Error
	}()

	report, err := CheckSchemaDrift(ctx, client.Gorm())
	if err != nil {
		t.Fatalf("check schema drift: %v", err)
	}

	if !report.HasDrift() {
		t.Fatal("expected drift detected for untracked table, got clean")
	}

	found := false
	for _, tbl := range report.UntrackedTables {
		if tbl == "temp_untracked_drift_table" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected temp_untracked_drift_table in untracked tables, got: %v", report.UntrackedTables)
	}

	// Clean up and recheck
	if err := client.Gorm().WithContext(ctx).Exec(dropTempTable).Error; err != nil {
		t.Fatalf("drop temp table: %v", err)
	}

	cleanReport, err := CheckSchemaDrift(ctx, client.Gorm())
	if err != nil {
		t.Fatalf("check schema drift after cleanup: %v", err)
	}
	if cleanReport.HasDrift() {
		t.Fatalf("expected clean schema after table drop, got:\n%s", cleanReport.Error())
	}
}

func freshTestConfig(databaseURL string) config.Config {
	return config.Config{
		Database: config.DatabaseConfig{
			URL:                databaseURL,
			MaxOpenConns:       10,
			MaxIdleConns:       5,
			ConnMaxLifetimeSec: 300,
		},
		Observability: config.ObservabilityConfig{
			HealthCheckTimeoutMs: 3000,
		},
	}
}
