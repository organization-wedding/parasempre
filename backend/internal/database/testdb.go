package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	host := envOrDefault("DB_HOST", "localhost")
	port := envOrDefault("DB_PORT", "5432")
	user := envOrDefault("DB_USER", "postgres")
	password := envOrDefault("DB_PASSWORD", "postgres")
	name := envOrDefault("DB_NAME", "parasempre_test")
	sslmode := envOrDefault("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, name, sslmode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return pool
}

func BeginTestTx(t *testing.T, pool *pgxpool.Pool) pgx.Tx {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin test transaction: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})
	return tx
}

func CleanTable(t *testing.T, pool *pgxpool.Pool, table string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE %s RESTART IDENTITY CASCADE", table))
	if err != nil {
		t.Fatalf("failed to truncate %s: %v", table, err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
