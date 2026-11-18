// Package migrations creates and updates the tables of the backend schema.
// The files are numbered, embedded in the binary and run in order by the role
// that owns the tables (ADR-022). A file that was already applied is never
// edited: a change is a new file.
package migrations

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"regexp"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var files embed.FS

// lockKey is an arbitrary number that identifies the migrations among the
// advisory locks of the database.
const lockKey = 7_365_829_104

const undefinedTable = "42P01"

var (
	// ErrPending means the schema is behind the binary.
	ErrPending = errors.New("pending migrations")
	// ErrChanged means a migration was edited after it was applied.
	ErrChanged = errors.New("migration changed after it was applied")
	// ErrUnknown means the schema has a migration that the binary does not,
	// so the binary is older than the schema.
	ErrUnknown = errors.New("migration applied but unknown to this binary")
	// ErrOutOfOrder means a migration that was not applied sorts before one
	// that was, as when two branches add a migration with the same number.
	ErrOutOfOrder = errors.New("migration sorts before an applied one")
	// ErrInvalidName means a file is not named like 0001_description.sql or
	// shares its number with another.
	ErrInvalidName = errors.New("invalid migration name")
	// ErrTransactionControl means a migration ended the transaction of Apply,
	// for example with a COMMIT. What it ran up to there stays applied, so the
	// schema has to be fixed by hand. A fresh database in the continuous
	// integration finds it before it reaches a shared one.
	ErrTransactionControl = errors.New("migration ended the transaction of the migrations")
)

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type migration struct {
	name     string
	sql      string
	checksum string
}

// Apply runs, in one transaction, the migrations that were not applied yet,
// and returns how many it applied. The lock lets two runs overlap: the first
// one migrates and the second finds nothing to do. The transaction is READ
// COMMITTED whatever the default of the database: at a stricter level its
// snapshot would be taken before the lock, and a run that waited would not
// see what the first one applied.
func Apply(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	return apply(ctx, pool, files)
}

func apply(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) (int, error) {
	all, err := load(fsys)
	if err != nil {
		return 0, err
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return 0, fmt.Errorf("begin the migrations: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey); err != nil {
		return 0, fmt.Errorf("lock the migrations: %w", err)
	}
	// Bounds how long a migration waits for a table that the server is using,
	// so it does not queue every query behind it. It is set after taking the
	// lock of the migrations, so a run waits for another one as long as ctx
	// allows.
	if _, err := tx.Exec(ctx, "SET LOCAL lock_timeout = '10s'"); err != nil {
		return 0, fmt.Errorf("set the lock timeout: %w", err)
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		name       text PRIMARY KEY,
		checksum   text NOT NULL,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return 0, fmt.Errorf("create schema_migrations: %w", err)
	}
	pending, err := pendingOf(ctx, tx, all)
	if err != nil {
		return 0, err
	}
	started, err := transactionID(ctx, tx)
	if err != nil {
		return 0, err
	}
	for _, m := range pending {
		if _, err := tx.Exec(ctx, m.sql); err != nil {
			return 0, fmt.Errorf("migration %s: %w", m.name, err)
		}
		if current, err := transactionID(ctx, tx); err != nil || current != started {
			return 0, errors.Join(fmt.Errorf("%w: %s", ErrTransactionControl, m.name), err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (name, checksum) VALUES ($1, $2)", m.name, m.checksum); err != nil {
			return 0, fmt.Errorf("record migration %s: %w", m.name, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit the migrations: %w", err)
	}
	return len(pending), nil
}

// transactionID changes if a migration ends the transaction, even if it opens
// another one afterwards.
func transactionID(ctx context.Context, tx pgx.Tx) (int64, error) {
	var id int64
	if err := tx.QueryRow(ctx, "SELECT txid_current()").Scan(&id); err != nil {
		return 0, fmt.Errorf("read the transaction id: %w", err)
	}
	return id, nil
}

// Check fails with ErrPending if a migration of the binary was not applied,
// so the application refuses to start against an old schema instead of
// failing on the first query.
func Check(ctx context.Context, pool *pgxpool.Pool) error {
	all, err := load(files)
	if err != nil {
		return err
	}
	pending, err := pendingOf(ctx, pool, all)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == undefinedTable {
		return fmt.Errorf("%w: no migration was applied", ErrPending)
	}
	if err != nil {
		return err
	}
	if len(pending) > 0 {
		return fmt.Errorf("%w, from %s on", ErrPending, pending[0].name)
	}
	return nil
}

// pendingOf also fails if an applied migration was edited or is not in the
// binary, which means the binary is older than the schema, or if a pending
// one sorts before an applied one.
func pendingOf(ctx context.Context, db querier, all []migration) ([]migration, error) {
	rows, err := db.Query(ctx, "SELECT name, checksum FROM schema_migrations ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	applied := map[string]string{}
	var name, checksum string
	if _, err := pgx.ForEachRow(rows, []any{&name, &checksum}, func() error {
		applied[name] = checksum
		return nil
	}); err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}

	var pending []migration
	for _, m := range all {
		checksum, ok := applied[m.name]
		switch {
		case !ok:
			pending = append(pending, m)
		case checksum != m.checksum:
			return nil, fmt.Errorf("%w: %s", ErrChanged, m.name)
		case len(pending) > 0:
			return nil, fmt.Errorf("%w: %s was not applied and %s was", ErrOutOfOrder, pending[0].name, m.name)
		}
		delete(applied, m.name)
	}
	if unknown := slices.Sorted(maps.Keys(applied)); len(unknown) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrUnknown, unknown[0])
	}
	return pending, nil
}

// load reads the migrations of fsys in order. Each name is four digits, an
// underscore and a description, and no two share a number.
func load(fsys fs.FS) ([]migration, error) {
	names, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return nil, fmt.Errorf("list the migrations: %w", err)
	}
	slices.Sort(names)
	all := make([]migration, 0, len(names))
	for i, name := range names {
		if !validName.MatchString(name) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidName, name)
		}
		if i > 0 && name[:4] == names[i-1][:4] {
			return nil, fmt.Errorf("%w: %s and %s share a number", ErrInvalidName, names[i-1], name)
		}
		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}
		sum := sha256.Sum256(content)
		all = append(all, migration{name: name, sql: string(content), checksum: hex.EncodeToString(sum[:])})
	}
	return all, nil
}

var validName = regexp.MustCompile(`^[0-9]{4}_[a-z0-9_]+\.sql$`)
