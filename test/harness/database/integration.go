// Package databaseharness provides reusable helpers for DB integration tests.
package databaseharness

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
)

// RequireDatabaseURLFromEnv reads DB url from env and skips test when not configured.
func RequireDatabaseURLFromEnv(t testing.TB, envKey string) string {
	t.Helper()

	databaseURL := strings.TrimSpace(os.Getenv(strings.TrimSpace(envKey)))
	if databaseURL == "" {
		t.Skipf("%s tidak diatur", envKey)
	}
	if !strings.Contains(databaseURL, "_test") {
		t.Skipf("%s wajib mengarah ke database *_test", envKey)
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
