package databaseharness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectRoot_ResolvesRepositoryRoot(t *testing.T) {
	root := ProjectRoot(t)
	if root == "" {
		t.Fatal("project root is empty")
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("project root missing go.mod: %s (%v)", root, err)
	}
}

func TestMigrationsPath_UsesProjectRoot(t *testing.T) {
	path := MigrationsPath(t)
	if filepath.Base(path) != "migrations" {
		t.Fatalf("unexpected migrations path: %s", path)
	}
}

func TestBaselineSeedPath_Exists(t *testing.T) {
	path := BaselineSeedPath(t)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("baseline seed file not found: %s (%v)", path, err)
	}
}

func TestLatestMigrationVersion_ReturnsNonZero(t *testing.T) {
	version := LatestMigrationVersion(t)
	if version == 0 {
		t.Fatal("expected latest migration version to be non-zero")
	}
}
