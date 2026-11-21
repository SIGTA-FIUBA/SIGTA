// Package database has what the PostgreSQL adapters of every module share.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InReadCommitted runs write in a READ COMMITTED transaction whatever the
// default of the database. The writes of the adapters need no stricter
// level: each one first takes a lock, on a row or an advisory one, and each
// statement after the lock sees what the previous holder committed, which at
// a stricter level it would not, since the snapshot is taken before the lock.
func InReadCommitted(ctx context.Context, pool *pgxpool.Pool, write func(tx pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if err := write(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Lock takes the advisory lock key until tx ends.
func Lock(ctx context.Context, tx pgx.Tx, key int64) error {
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", key)
	return err
}

// Nullable stores an empty text as NULL.
func Nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// NullableTime stores a zero time as NULL.
func NullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// ValueOf reads NULL back as an empty text.
func ValueOf(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// TimeOf reads NULL back as a zero time, and a time in UTC, as the domain
// keeps it.
func TimeOf(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.UTC()
}

// Array keeps an empty list as an empty array: the columns are NOT NULL, and
// pgx sends a nil slice as NULL.
func Array[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// List reads an empty array back as nil, as the memory adapters keep it.
func List[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return s
}
