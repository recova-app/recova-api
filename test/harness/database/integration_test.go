package databaseharness

import (
	"path/filepath"
	"testing"
)

func TestProjectRoot_ResolvesRepositoryRoot(t *testing.T) {
	root := ProjectRoot(t)
	if filepath.Base(root) != "recova-backend-v2" {
		t.Fatalf("unexpected root path: %s", root)
	}
}

func TestMigrationsPath_UsesProjectRoot(t *testing.T) {
	path := MigrationsPath(t)
	if filepath.Base(path) != "migrations" {
		t.Fatalf("unexpected migrations path: %s", path)
	}
}
