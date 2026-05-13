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
