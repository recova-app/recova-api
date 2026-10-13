// Package databaseharness provides reusable helpers for DB integration tests.
package databaseharness

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

// RequireDatabaseURLFromEnv reads DB url from env and skips test when not configured.
func RequireDatabaseURLFromEnv(t testing.TB, envKey string) string {
	t.Helper()

	databaseURL := strings.TrimSpace(os.Getenv(strings.TrimSpace(envKey)))
	if databaseURL == "" {
		t.Skipf("%s is not configured", envKey)
	}
	if !strings.Contains(databaseURL, "_test") {
		t.Skipf("%s must point to a *_test database", envKey)
	}

	return databaseURL
}

// ProjectRoot returns repository root by resolving from this file path.
func ProjectRoot(t testing.TB) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}

	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("resolve project root: %v", err)
	}

	return absRoot
}

// MigrationsPath resolves absolute path to migrations directory.
func MigrationsPath(t testing.TB) string {
	t.Helper()
	return filepath.Join(ProjectRoot(t), "migrations")
}

// ResetMigrations resets migration state with down then up for deterministic integration tests.
func ResetMigrations(t testing.TB, databaseURL string) {
	t.Helper()

	m, err := migrate.New("file://"+MigrationsPath(t), databaseURL)
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
}

// BaselineSeedPath resolves path to baseline seed SQL file.
func BaselineSeedPath(t testing.TB) string {
	t.Helper()
	return filepath.Join(MigrationsPath(t), "seeds", "000001_baseline_seed.sql")
}

// LatestMigrationVersion finds the latest migration timestamp version from migrations directory.
func LatestMigrationVersion(t testing.TB) uint64 {
	t.Helper()

	entries, err := os.ReadDir(MigrationsPath(t))
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var latest uint64
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			parts := strings.SplitN(name, "_", 2)
			if len(parts) > 0 {
				if version, err := strconv.ParseUint(parts[0], 10, 64); err == nil {
					if version > latest {
						latest = version
					}
				}
			}
		}
	}

	if latest == 0 {
		t.Fatal("no up migrations found to determine latest version")
	}

	return latest
}

// ApplyBaselineSeed reads and executes baseline seed SQL on the target GORM database handle.
func ApplyBaselineSeed(t testing.TB, gormDB *gorm.DB) {
	t.Helper()

	seedPath := BaselineSeedPath(t)
	content, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("read baseline seed file %s: %v", seedPath, err)
	}

	if err := gormDB.Exec(string(content)).Error; err != nil {
		t.Fatalf("execute baseline seed: %v", err)
	}
}
