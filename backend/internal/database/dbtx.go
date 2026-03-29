package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX abstracts *pgxpool.Pool and pgx.Tx so repositories can work with either.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TxRunner runs a function inside a database transaction.
type TxRunner interface {
	RunInTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type poolTxRunner struct {
	pool *pgxpool.Pool
}

// NewTxRunner creates a TxRunner backed by a pgxpool.Pool.
func NewTxRunner(pool *pgxpool.Pool) TxRunner {
	return &poolTxRunner{pool: pool}
}

func (r *poolTxRunner) RunInTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
