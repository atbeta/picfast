package sqlc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunInTx begins a transaction, runs fn with a *Queries scoped to that tx,
// and commits on success (or rolls back on error/panic).
func RunInTx(ctx context.Context, pool *pgxpool.Pool, fn func(*Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := New(tx)
	if err := fn(qtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
