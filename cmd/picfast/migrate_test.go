package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/atbeta/picfast/migrations"
)

func TestRunMigrateCommandUsage(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{"help", []string{"help"}, 0},
		{"short help", []string{"-h"}, 0},
		{"long help", []string{"--help"}, 0},
		{"unknown command", []string{"bogus"}, 2},
		{"extra argument", []string{"up", "extra"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runMigrateCommand(context.Background(), tt.args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("runMigrateCommand(%v) = %d, want %d (stdout=%q stderr=%q)", tt.args, code, tt.wantCode, stdout.String(), stderr.String())
			}
		})
	}
}

// TestRunMigrateCommandUpAppliesMigrations runs `picfast migrate up` against a
// scratch database, proving the embedded migration source works end to end.
func TestRunMigrateCommandUpAppliesMigrations(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("PICFAST_TEST_DB_URL")
	if baseURL == "" {
		baseURL = "postgres://picfast:picfast@localhost:5432/picfast_test?sslmode=disable"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse test db url: %v", err)
	}
	adminURL := *parsed
	adminURL.Path = "/postgres"

	admin, err := pgx.Connect(ctx, adminURL.String())
	if err != nil {
		t.Skipf("test postgres unavailable: %v", err)
	}
	defer admin.Close(ctx)

	dbName := fmt.Sprintf("picfast_migrate_cli_%d_%d", os.Getpid(), time.Now().UnixNano())
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{dbName}.Sanitize()); err != nil {
		t.Fatalf("create scratch database: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(), "DROP DATABASE IF EXISTS "+pgx.Identifier{dbName}.Sanitize())
	})

	target := *parsed
	target.Path = "/" + dbName
	t.Setenv("PICFAST_DATABASE_URL", target.String())

	var stdout, stderr bytes.Buffer
	if code := runMigrateCommand(ctx, []string{"up"}, &stdout, &stderr); code != 0 {
		t.Fatalf("runMigrateCommand(up) = %d, want 0 (stdout=%q stderr=%q)", code, stdout.String(), stderr.String())
	}

	conn, err := pgx.Connect(ctx, target.String())
	if err != nil {
		t.Fatalf("connect scratch database: %v", err)
	}
	defer conn.Close(ctx)

	var version int64
	var dirty bool
	if err := conn.QueryRow(ctx, "SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatalf("read schema_migrations: %v", err)
	}
	if dirty {
		t.Fatal("schema_migrations is dirty after migrate up")
	}
	if want := latestEmbeddedMigrationVersion(t); version != want {
		t.Fatalf("migration version = %d, want %d", version, want)
	}

	stdout.Reset()
	stderr.Reset()
	if code := runMigrateCommand(ctx, []string{"up"}, &stdout, &stderr); code != 0 {
		t.Fatalf("second migrate up = %d, want 0 (stderr=%q)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "up to date") {
		t.Fatalf("expected 'up to date' output, got %q", stdout.String())
	}
}

func latestEmbeddedMigrationVersion(t *testing.T) int64 {
	t.Helper()
	names, err := fs.Glob(migrations.FS, "*.up.sql")
	if err != nil || len(names) == 0 {
		t.Fatalf("glob embedded migrations: %v (n=%d)", err, len(names))
	}
	var latest int64
	for _, name := range names {
		prefix, _, _ := strings.Cut(name, "_")
		v, err := strconv.ParseInt(prefix, 10, 64)
		if err != nil {
			t.Fatalf("parse migration version from %q: %v", name, err)
		}
		latest = max(latest, v)
	}
	return latest
}
