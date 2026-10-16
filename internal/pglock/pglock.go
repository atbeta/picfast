// Package pglock coordinates background work across app instances with
// Postgres advisory locks, so a job run by a ticker, an HTTP trigger, or a
// CLI invocation executes on at most one instance at a time.
//
// Locks are transaction-scoped (pg_try_advisory_xact_lock), not
// session-scoped: transaction-mode connection poolers (pgbouncer, Neon
// -pooler hosts, Supabase :6543) may route consecutive statements on one
// pooled connection to different server backends, which breaks session
// locks — the unlock can land on a different backend and the lock leaks.
// Holding the lock inside an open transaction pins the statements to one
// server backend for the job duration and frees the lock automatically at
// transaction end or on connection loss.
package pglock

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
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
	// KeyBootstrapDefaults serializes first-boot seeding (default group,
	// guest group, default storage strategy, admin user) so several instances
	// starting against the same fresh database cannot seed duplicates.
	KeyBootstrapDefaults int64 = 727004
)

const unlockTimeout = 5 * time.Second

// Acquire tries to take the transaction-scoped advisory lock key on a
// dedicated pooled connection. It returns (nil, false, nil) when the lock is
// held by another session. When ok is true the caller owns the lock until
// release is invoked; release is idempotent and uses a fresh context so a
// canceled caller context cannot leak the lock. The lock is also freed if
// the connection dies.
func Acquire(ctx context.Context, pool *pgxpool.Pool, key int64) (release func(), ok bool, err error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("pglock: acquire conn: %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		conn.Release()
		return nil, false, fmt.Errorf("pglock: begin tx: %w", err)
	}

	var locked bool
	if err := tx.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock($1)", key).Scan(&locked); err != nil {
		rollbackTx(tx)
		conn.Release()
		return nil, false, fmt.Errorf("pglock: try lock %d: %w", key, err)
	}
	if !locked {
		rollbackTx(tx)
		conn.Release()
		return nil, false, nil
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			defer conn.Release()
			// The xact lock is released by ending the transaction; use a
			// fresh context so a canceled caller ctx cannot leak it.
			unlockCtx, cancel := context.WithTimeout(context.Background(), unlockTimeout)
			defer cancel()
			if err := tx.Rollback(unlockCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				slog.Warn("pglock: lock tx rollback failed; lock releases when the connection closes", "key", key, "error", err)
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

func rollbackTx(tx pgx.Tx) {
	unlockCtx, cancel := context.WithTimeout(context.Background(), unlockTimeout)
	defer cancel()
	if err := tx.Rollback(unlockCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		slog.Warn("pglock: lock tx rollback failed", "error", err)
	}
}
