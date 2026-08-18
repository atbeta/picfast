package testutil

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testDBURL() string {
	if u := os.Getenv("PICFAST_TEST_DB_URL"); u != "" {
		return u
	}
	return "postgres://picfast:picfast@localhost:5432/picfast_test?sslmode=disable"
}

// migrateDatabaseURL 把 postgres:// 形式的 DSN 改写成 golang-migrate pgx/v5
// 驱动注册的 pgx5:// scheme，避免引入 lib/pq 驱动（其有未修复的 CVE）。
func migrateDatabaseURL(dbURL string) string {
	for _, scheme := range []string{"postgres://", "postgresql://"} {
		if strings.HasPrefix(dbURL, scheme) {
			return "pgx5://" + strings.TrimPrefix(dbURL, scheme)
		}
	}
	return dbURL
}

// SetupDB connects to the test database, runs migrations, and returns a pool + queries.
// It registers t.Cleanup to truncate all tables and close the pool.
func SetupDB(t *testing.T) (*pgxpool.Pool, *sqlc.Queries) {
	t.Helper()

	dbURL := isolatedTestDBURL(t, testDBURL())
	if err := ensureDatabaseExists(dbURL); err != nil {
		t.Fatalf("ensure test database: %v", err)
	}

	// Run migrations
	migrationsDir := migrationsPath(t)
	m, err := migrate.New("file://"+migrationsDir, migrateDatabaseURL(dbURL))
	if err != nil {
		t.Fatalf("create migrator: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("run migrations: %v", err)
	}
	m.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dbURL)
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

	TruncateAll(t, pool)

	t.Cleanup(func() {
		TruncateAll(t, pool)
		pool.Close()
		if err := dropDatabase(dbURL); err != nil {
			t.Logf("drop test database: %v", err)
		}
	})

	return pool, sqlc.New(pool)
}

func isolatedTestDBURL(t *testing.T, rawURL string) string {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Path == "" {
		return rawURL
	}
	baseName := strings.TrimPrefix(parsed.Path, "/")
	suffix := sanitizeDatabaseSuffix(t.Name())
	pid := fmt.Sprintf("_%d", os.Getpid())
	maxSuffixLen := 63 - len(baseName) - len(pid) - 1
	if maxSuffixLen < 8 {
		maxSuffixLen = 8
	}
	if len(suffix) > maxSuffixLen {
		suffix = suffix[:maxSuffixLen]
	}
	parsed.Path = "/" + baseName + "_" + suffix + pid
	return parsed.String()
}

func sanitizeDatabaseSuffix(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return strings.Trim(b.String(), "_")
}

// TruncateAll clears all table data in dependency order.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	tables := []string{
		"site_settings",
		"email_verification_tokens",
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

func ensureDatabaseExists(dbURL string) error {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("parse db config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err == nil {
		pingErr := pool.Ping(ctx)
		pool.Close()
		if pingErr == nil {
			return nil
		}
		if !isMissingDatabaseError(pingErr, config.ConnConfig.Database) {
			return nil
		}
	} else if !isMissingDatabaseError(err, config.ConnConfig.Database) {
		return nil
	}

	adminConfig := config.ConnConfig.Copy()
	adminConfig.Database = "postgres"

	conn, err := pgx.ConnectConfig(ctx, adminConfig)
	if err != nil {
		return fmt.Errorf("connect admin database: %w", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(
		ctx,
		fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{config.ConnConfig.Database}.Sanitize()),
	)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("create database %q: %w", config.ConnConfig.Database, err)
	}

	return nil
}

func dropDatabase(dbURL string) error {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("parse db config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	adminConfig := config.ConnConfig.Copy()
	dbName := config.ConnConfig.Database
	adminConfig.Database = "postgres"

	conn, err := pgx.ConnectConfig(ctx, adminConfig)
	if err != nil {
		return fmt.Errorf("connect admin database: %w", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", pgx.Identifier{dbName}.Sanitize()))
	if err != nil {
		return fmt.Errorf("drop database %q: %w", dbName, err)
	}
	return nil
}

func isMissingDatabaseError(err error, database string) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), fmt.Sprintf("database %q does not exist", database))
}
