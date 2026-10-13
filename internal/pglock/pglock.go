// Package pglock coordinates background work across app instances with
// Postgres advisory locks, so a job run by a ticker, an HTTP trigger, or a
// CLI invocation executes on at most one instance at a time.
package pglock

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
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

// Acquire tries to take the advisory lock key on a dedicated pooled
// connection. It returns (nil, false, nil) when the lock is held by another
// session. When ok is true the caller owns the lock until release is invoked;
// release is idempotent and uses a fresh context so a canceled caller context
// cannot leak the lock.
func Acquire(ctx context.Context, pool *pgxpool.Pool, key int64) (release func(), ok bool, err error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("pglock: acquire conn: %w", err)
	}

	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked); err != nil {
		conn.Release()
		return nil, false, fmt.Errorf("pglock: try lock %d: %w", key, err)
	}
	if !locked {
		conn.Release()
		return nil, false, nil
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			defer conn.Release()
			unlockCtx, cancel := context.WithTimeout(context.Background(), unlockTimeout)
			defer cancel()
			if _, err := conn.Exec(unlockCtx, "SELECT pg_advisory_unlock($1)", key); err != nil {
				slog.Warn("pglock: unlock failed; lock releases when the connection closes", "key", key, "error", err)
			}
		})
	}, true, nil
}

// TryWithLock acquires the advisory lock key on a dedicated connection and,
// when acquired, runs fn and reports (true, fn's error). It returns
// (false, nil) when the lock is held by another session.
func TryWithLock(ctx context.Context, pool *pgxpool.Pool, key int64, fn func(context.Context) error) (acquired bool, err error) {
	release, ok, err := Acquire(ctx, pool, key)
	if err != nil || !ok {
		return false, err
	}
	defer release()

	if err := fn(ctx); err != nil {
		return true, err
	}
	return true, nil
}
