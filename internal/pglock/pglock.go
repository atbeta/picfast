// Package pglock coordinates background work across app instances with
// Postgres advisory locks, so a job run by a ticker, an HTTP trigger, or a
// CLI invocation executes on at most one instance at a time.
package pglock

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Advisory lock keys shared by every instance. The values are arbitrary but
// must be distinct and stable across releases so all instances compete for
// the same locks.
const (
	// KeyJobsWebhooks serializes the periodic webhook outbox/delivery pass.
	KeyJobsWebhooks int64 = 727001
	// KeyJobsCleanup serializes the hourly cleanup (expired images, webhook
	// delivery/outbox retention, audit-log retention).
	KeyJobsCleanup int64 = 727002
	// KeyRecalcPHash serializes the admin-triggered pHash recalculation.
	KeyRecalcPHash int64 = 727003
)

const unlockTimeout = 5 * time.Second

// TryWithLock acquires the advisory lock key on a dedicated connection and,
// when acquired, runs fn and reports (true, fn's error). It returns
// (false, nil) when the lock is held by another session. The lock is released
// with a fresh context so a canceled ctx cannot leak it.
func TryWithLock(ctx context.Context, pool *pgxpool.Pool, key int64, fn func(context.Context) error) (acquired bool, err error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("pglock: acquire conn: %w", err)
	}
	defer conn.Release()

	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked); err != nil {
		return false, fmt.Errorf("pglock: try lock %d: %w", key, err)
	}
	if !locked {
		return false, nil
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), unlockTimeout)
		defer cancel()
		if _, err := conn.Exec(unlockCtx, "SELECT pg_advisory_unlock($1)", key); err != nil {
			slog.Warn("pglock: unlock failed; lock releases when the connection closes", "key", key, "error", err)
		}
	}()

	if err := fn(ctx); err != nil {
		return true, err
	}
	return true, nil
}
