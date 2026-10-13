package main

import (
	"os"
	"path/filepath"
	"testing"
)

// unsetEnv removes an env var for the duration of the test, restoring any
// previous value afterwards.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, had := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unsetenv %s: %v", key, err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, old)
		}
	})
}

func TestNewMigrationSourceFallsBackToEmbedded(t *testing.T) {
	// No migrations dir relative to CWD and no PICFAST_MIGRATIONS_DIR:
	// the embedded FS must be used. (/migrations and /app/migrations do not
	// exist on dev machines or CI runners.)
	t.Chdir(t.TempDir())
	unsetEnv(t, "PICFAST_MIGRATIONS_DIR")

	name, src, err := newMigrationSource()
	if err != nil {
		t.Fatalf("newMigrationSource: %v", err)
	}
	defer src.Close()

	if name != "iofs" {
		t.Fatalf("source name = %q, want iofs (embedded)", name)
	}
	version, err := src.First()
	if err != nil {
		t.Fatalf("src.First: %v", err)
	}
	if version != 1 {
		t.Fatalf("first embedded migration version = %d, want 1", version)
	}
}

func TestNewMigrationSourceIgnoresInvalidEnvDir(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("PICFAST_MIGRATIONS_DIR", filepath.Join(t.TempDir(), "does-not-exist"))

	name, src, err := newMigrationSource()
	if err != nil {
		t.Fatalf("newMigrationSource: %v", err)
	}
	defer src.Close()

	if name != "iofs" {
		t.Fatalf("source name = %q, want iofs (embedded)", name)
	}
}

func TestNewMigrationSourcePrefersEnvDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "900_test.up.sql"), []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatalf("write test migration: %v", err)
	}
	t.Setenv("PICFAST_MIGRATIONS_DIR", dir)

	name, src, err := newMigrationSource()
	if err != nil {
		t.Fatalf("newMigrationSource: %v", err)
	}
	defer src.Close()

	if name != "file" {
		t.Fatalf("source name = %q, want file", name)
	}
	version, err := src.First()
	if err != nil {
		t.Fatalf("src.First: %v", err)
	}
	if version != 900 {
		t.Fatalf("first migration version = %d, want 900", version)
	}
}
