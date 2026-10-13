package migrations

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestEmbeddedFSContainsAllMigrationFiles(t *testing.T) {
	embedded, err := fs.Glob(FS, "*.sql")
	if err != nil {
		t.Fatalf("glob embedded FS: %v", err)
	}
	if len(embedded) == 0 {
		t.Fatal("embedded FS contains no migration files")
	}

	// Tests run with the package directory as CWD, so "*.sql" lists the
	// migration files on disk that should have been embedded.
	onDisk, err := filepath.Glob("*.sql")
	if err != nil {
		t.Fatalf("glob disk migrations: %v", err)
	}

	sort.Strings(embedded)
	sort.Strings(onDisk)
	if len(embedded) != len(onDisk) {
		t.Fatalf("embedded %d sql files, disk has %d: embedded=%v onDisk=%v", len(embedded), len(onDisk), embedded, onDisk)
	}
	for i := range embedded {
		if embedded[i] != onDisk[i] {
			t.Fatalf("embedded migration %q does not match disk file %q", embedded[i], onDisk[i])
		}
		embeddedBytes, err := fs.ReadFile(FS, onDisk[i])
		if err != nil {
			t.Fatalf("read embedded %s: %v", onDisk[i], err)
		}
		diskBytes, err := os.ReadFile(onDisk[i])
		if err != nil {
			t.Fatalf("read disk %s: %v", onDisk[i], err)
		}
		if !bytes.Equal(embeddedBytes, diskBytes) {
			t.Fatalf("embedded %s content differs from disk", onDisk[i])
		}
	}
}

func TestEmbeddedFSContainsLatestMigration(t *testing.T) {
	// Glob on disk for the newest *.up.sql and assert it exists in the FS.
	onDisk, err := filepath.Glob("*.up.sql")
	if err != nil {
		t.Fatalf("glob disk migrations: %v", err)
	}
	if len(onDisk) == 0 {
		t.Fatal("no *.up.sql files on disk")
	}
	sort.Strings(onDisk)
	latest := onDisk[len(onDisk)-1]

	data, err := fs.ReadFile(FS, latest)
	if err != nil {
		t.Fatalf("latest migration %s missing from embedded FS: %v", latest, err)
	}
	if len(data) == 0 {
		t.Fatalf("embedded latest migration %s is empty", latest)
	}
}
