package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/jobs"
	"github.com/atbeta/picfast/internal/sqlc"
)

// runJobsCommand implements `picfast jobs run`: one synchronous pass over the
// background jobs (webhook dispatch + cleanup) for external schedulers when
// jobs.embedded=false. No vips startup is needed for this subcommand.
func runJobsCommand(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printJobsUsage(stderr)
		return 2
	}
	switch args[0] {
	case "help", "-h", "--help":
		printJobsUsage(stdout)
		return 0
	case "run":
		args = args[1:]
	default:
		fmt.Fprintf(stderr, "unknown jobs command: %s\n\n", args[0])
		printJobsUsage(stderr)
		return 2
	}
	if len(args) > 0 {
		fmt.Fprintf(stderr, "unexpected jobs argument: %s\n\n", args[0])
		printJobsUsage(stderr)
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(stderr, "failed to connect to database: %v\n", err)
		return 1
	}
	defer pool.Close()
	if n := pool.Config().MaxConns; n < 3 {
		fmt.Fprintf(stderr, "warning: pool max_conns=%d is small for background jobs: a running job holds one connection for its advisory lock while issuing queries on others; set pool_max_conns >= 3\n", n)
	}

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(stderr, "failed to ping database: %v\n", err)
		return 1
	}

	report := jobs.NewRunner(sqlc.New(pool), pool, cfg).RunOnce(ctx)

	out, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "failed to encode report: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, string(out))

	if report.Failed() {
		return 1
	}
	return 0
}

func printJobsUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  picfast jobs run

Commands:
  run                 Run one pass of the background jobs (webhook dispatch + cleanup)`)
}
