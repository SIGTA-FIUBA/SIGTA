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
	"regexp"
	"slices"
	"strings"

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
	// ErrTransactionControl means a migration ends or opens a transaction, or
	// does something that cannot run inside one. Every migration runs in the
	// transaction of Apply, and a COMMIT inside it would leave the schema half
	// migrated.
	ErrTransactionControl = errors.New("migration controls its own transaction")
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
	all, err := load(files)
	if err != nil {
		return 0, err
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return 0, err
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
	for _, m := range pending {
		if _, err := tx.Exec(ctx, m.sql); err != nil {
			return 0, fmt.Errorf("migration %s: %w", m.name, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (name, checksum) VALUES ($1, $2)", m.name, m.checksum); err != nil {
			return 0, fmt.Errorf("record migration %s: %w", m.name, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(pending), nil
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
// binary, which means the binary is older than the schema.
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
		}
		delete(applied, m.name)
	}
	for name := range applied {
		return nil, fmt.Errorf("%w: %s", ErrUnknown, name)
	}
	return pending, nil
}

// load reads the migrations of fsys in order and refuses one that controls
// its own transaction.
func load(fsys fs.FS) ([]migration, error) {
	names, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return nil, err
	}
	slices.Sort(names)
	all := make([]migration, 0, len(names))
	for _, name := range names {
		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}
		if found := transactionControl.FindString(withoutQuotes(string(content))); found != "" {
			return nil, fmt.Errorf("%w: %s has %s", ErrTransactionControl, name, strings.ToUpper(found))
		}
		sum := sha256.Sum256(content)
		all = append(all, migration{name: name, sql: string(content), checksum: hex.EncodeToString(sum[:])})
	}
	return all, nil
}

// transactionControl finds the statements that would end the transaction of
// Apply, and CONCURRENTLY, which cannot run inside one. It may refuse a
// migration that only names them, which is the safe side.
var transactionControl = regexp.MustCompile(`(?i)\b(begin|commit|rollback|start\s+transaction|concurrently)\b`)

// withoutQuotes blanks comments, strings, quoted identifiers and dollar
// quoted bodies, such as the body of a function, where those words do not
// control the transaction of Apply.
func withoutQuotes(sql string) string {
	var out strings.Builder
	for i := 0; i < len(sql); {
		var end int
		switch {
		case strings.HasPrefix(sql[i:], "--"):
			end = closing(sql, i+2, "\n")
		case strings.HasPrefix(sql[i:], "/*"):
			end = closing(sql, i+2, "*/")
		case sql[i] == '\'' || sql[i] == '"':
			end = closing(sql, i+1, sql[i:i+1])
		case sql[i] == '$':
			tag := dollarTag.FindString(sql[i:])
			if tag == "" {
				out.WriteByte(sql[i])
				i++
				continue
			}
			end = closing(sql, i+len(tag), tag)
		default:
			out.WriteByte(sql[i])
			i++
			continue
		}
		out.WriteByte(' ')
		i = end
	}
	return out.String()
}

var dollarTag = regexp.MustCompile(`^\$([A-Za-z_][A-Za-z0-9_]*)?\$`)

// closing returns the index right after the first delimiter at or after
// start, or the end of sql if there is none. A doubled quote inside a string
// reads as the end of one string and the start of another, which blanks the
// same text.
func closing(sql string, start int, delimiter string) int {
	if j := strings.Index(sql[start:], delimiter); j >= 0 {
		return start + j + len(delimiter)
	}
	return len(sql)
}
