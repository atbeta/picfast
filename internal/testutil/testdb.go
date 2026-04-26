package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pbeta/imgapi/internal/sqlc"
)

func testDBURL() string {
	if u := os.Getenv("IMGAPI_TEST_DB_URL"); u != "" {
		return u
	}
	return "postgres://imgapi:imgapi@localhost:5432/imgapi_test?sslmode=disable"
}

// SetupDB connects to the test database, runs migrations, and returns a pool + queries.
// It registers t.Cleanup to truncate all tables and close the pool.
func SetupDB(t *testing.T) (*pgxpool.Pool, *sqlc.Queries) {
	t.Helper()

	url := testDBURL()

	// Run migrations
	migrationsDir := migrationsPath(t)
	m, err := migrate.New("file://"+migrationsDir, url)
	if err != nil {
		t.Fatalf("create migrator: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("run migrations: %v", err)
	}
	m.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse db config: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping test db: %v", err)
	}

	t.Cleanup(func() {
		TruncateAll(t, pool)
		pool.Close()
	})

	return pool, sqlc.New(pool)
}

// TruncateAll clears all table data in dependency order.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	tables := []string{
		"refresh_tokens",
		"images",
		"albums",
		"users",
		"group_strategies",
		"strategies",
		"groups",
	}
	for _, tbl := range tables {
		if _, err := pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tbl)); err != nil {
			t.Logf("truncate %s: %v", tbl, err)
		}
	}
}

// migrationsPath returns the absolute path to the project's migrations/ directory.
func migrationsPath(t *testing.T) string {
	t.Helper()
	// This file: internal/testutil/testdb.go
	// Migrations:  ../../migrations/ (relative to this file)
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("resolve migrations path: %v", err)
	}
	return abs
}
