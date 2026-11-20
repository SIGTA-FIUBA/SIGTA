// Package postgres stores the auth module in PostgreSQL, with the tables of
// src/migrations and the privileges of the application role.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Arbitrary numbers, other than those of src/migrations, that identify the
// advisory lock under which each table numbers its new rows.
const (
	areasLockKey       = 7_365_829_106
	membershipsLockKey = 7_365_829_107
)

// inReadCommitted runs write in a READ COMMITTED transaction whatever the
// default of the database. Its writes need no stricter level: a new row
// takes its number under an advisory lock, and each statement after the lock
// sees what the previous holder committed, which at a stricter level it would
// not, since the snapshot is taken before the lock.
func inReadCommitted(ctx context.Context, pool *pgxpool.Pool, write func(tx pgx.Tx) error) error {
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

func lock(ctx context.Context, tx pgx.Tx, key int64) error {
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", key)
	return err
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func valueOf(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func timeOf(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.UTC()
}

// array keeps an empty list as an empty array: the columns are NOT NULL, and
// pgx sends a nil slice as NULL.
func array[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// list reads an empty array back as nil, as the other adapters keep it.
func list[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return s
}
