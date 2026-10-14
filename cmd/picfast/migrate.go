package main

import (
	"context"
	"fmt"
	"io"

	"github.com/atbeta/picfast/internal/config"
	"github.com/golang-migrate/migrate/v4"
)

// runMigrateCommand implements `picfast migrate [up]`. It applies pending
// database migrations without starting the server (no vips/web startup).
func runMigrateCommand(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "help", "-h", "--help":
			printMigrateUsage(stdout)
			return 0
		case "up":
			args = args[1:]
		default:
			_, _ = fmt.Fprintf(stderr, "unknown migrate command: %s\n\n", args[0])
			printMigrateUsage(stderr)
			return 2
		}
	}
	if len(args) > 0 {
		_, _ = fmt.Fprintf(stderr, "unexpected migrate argument: %s\n\n", args[0])
		printMigrateUsage(stderr)
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	m, err := newMigrator(cfg.Database.URL)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed to open migration source: %v\n", err)
		return 1
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			_, _ = fmt.Fprintln(stdout, "migrations already up to date")
			return 0
		}
		_, _ = fmt.Fprintf(stderr, "migration failed: %v\n", err)
		return 1
	}

	if version, dirty, err := m.Version(); err == nil {
		_, _ = fmt.Fprintf(stdout, "migrations applied: version=%d dirty=%t\n", version, dirty)
	} else {
		_, _ = fmt.Fprintln(stdout, "migrations applied")
	}
	return 0
}

func printMigrateUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, `Usage:
  picfast migrate [up]

Commands:
  up                  Apply all pending database migrations (default)`)
}
