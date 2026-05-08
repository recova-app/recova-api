package database

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/recova-app/backend-v2/internal/platform/config"
	"gorm.io/gorm"
)

func TestIntegration_ConnectAndMigrationRoundTrip(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("RECOVA_DB_INTEGRATION_URL"))
	if databaseURL == "" {
		t.Skip("RECOVA_DB_INTEGRATION_URL tidak diatur")
	}
	if !strings.Contains(databaseURL, "_test") {
		t.Skip("RECOVA_DB_INTEGRATION_URL wajib mengarah ke database *_test")
	}

	cfg := config.Config{
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

	client, err := Connect(cfg)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	if err := client.Ping(nil); err != nil {
		t.Fatalf("ping error: %v", err)
	}

	migrationsPath, err := migrationsAbsPath()
	if err != nil {
		t.Fatalf("resolve migrations path: %v", err)
	}

	m, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		t.Fatalf("new migrate instance: %v", err)
	}
	t.Cleanup(func() {
		_, _ = m.Close()
	})

	_ = m.Down()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migration up failed: %v", err)
	}

	if err := assertTableExists(client.Gorm(), "users"); err != nil {
		t.Fatalf("expected users table exists after up: %v", err)
	}

	if err := m.Steps(-1); err != nil {
		t.Fatalf("migration down one step failed: %v", err)
	}

	if err := assertTableExists(client.Gorm(), "users"); err == nil {
		t.Fatal("expected users table absent after one-step down")
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migration up (rerun) failed: %v", err)
	}
}

func migrationsAbsPath() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime caller unavailable")
	}

	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	return filepath.Abs(filepath.Join(root, "migrations"))
}

func assertTableExists(gormDB *gorm.DB, tableName string) error {
	const query = `
SELECT EXISTS (
	SELECT 1
	FROM information_schema.tables
	WHERE table_schema = 'public' AND table_name = $1
)`

	var exists bool
	if err := gormDB.Raw(query, tableName).Scan(&exists).Error; err != nil {
		return err
	}

	if !exists {
		return errors.New("table not found")
	}

	return nil
}
