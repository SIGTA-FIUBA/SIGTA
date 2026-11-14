// Package dbtest gives the tests that need PostgreSQL a schema of their own,
// migrated and with the same roles and privileges as the real one.
//
// SIGTA_TEST_DATABASE_URL points to a database as a superuser, for example
// the database of compose.yaml, where deploy/postgres/sql/roles.sql already
// created the roles. If they are missing, as in a cluster made with initdb,
// they are created without login. Without the variable the tests are skipped,
// except in the continuous integration, where they fail: there they must run.
package dbtest

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/migrations"
)

const (
	urlVariable = "SIGTA_TEST_DATABASE_URL"

	// The roles that deploy/postgres/sql/roles.sql creates.
	AppRole        = "backend"
	MigrationsRole = "backend_migrations"

	// rolesLockKey is an arbitrary number, other than the one of the
	// migrations.
	rolesLockKey = 7_365_829_105
)

// Isolation is the default isolation level of the transactions of a pool.
type Isolation string

const (
	ReadCommitted Isolation = "read committed"
	Serializable  Isolation = "serializable"
)

// Schema is a schema created for one test and dropped when it ends.
type Schema struct {
	Name string
	url  string
}

// URL returns SIGTA_TEST_DATABASE_URL, and skips the test if it is not set.
func URL(t testing.TB) string {
	t.Helper()
	url := os.Getenv(urlVariable)
	if url != "" {
		return url
	}
	if InCI() {
		t.Fatalf("%s is not set: the PostgreSQL tests must run in the continuous integration", urlVariable)
	}
	t.Skipf("set %s to run the PostgreSQL tests (see CONTRIBUTING.md)", urlVariable)
	return ""
}

// InCI reports whether the tests run in the continuous integration.
func InCI() bool { return os.Getenv("CI") != "" }

// NewSchema creates a schema with NewEmptySchema and runs the migrations on it.
func NewSchema(t testing.TB) *Schema {
	t.Helper()
	s := NewEmptySchema(t)
	if _, err := migrations.Apply(t.Context(), s.Pool(t, MigrationsRole, ReadCommitted)); err != nil {
		t.Fatalf("apply the migrations: %v", err)
	}
	return s
}

// NewEmptySchema creates a schema owned by the migrations role, without
// running the migrations. Its owner and privileges copy those of the backend
// schema in deploy/postgres/sql/roles.sql: a change to one goes to the other.
func NewEmptySchema(t testing.TB) *Schema {
	t.Helper()
	s := &Schema{Name: "test_" + strings.ToLower(rand.Text()[:12]), url: URL(t)}
	admin := connect(t, s.url)
	ctx := t.Context()
	createRoles(t, admin)
	for _, statement := range []string{
		fmt.Sprintf("CREATE SCHEMA %s AUTHORIZATION %s", s.Name, MigrationsRole),
		fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s", s.Name, AppRole),
	} {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(context.WithoutCancel(ctx), fmt.Sprintf("DROP SCHEMA %s CASCADE", s.Name)); err != nil {
			t.Errorf("drop schema %s: %v", s.Name, err)
		}
	})
	return s
}

// Pool connects as the superuser of SIGTA_TEST_DATABASE_URL and then takes
// the given role, so every query is checked against that role's privileges.
// With an empty role it stays the superuser.
func (s *Schema) Pool(t testing.TB, role string, isolation Isolation) *pgxpool.Pool {
	t.Helper()
	config, err := pgxpool.ParseConfig(s.url)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = s.Name
	config.ConnConfig.RuntimeParams["default_transaction_isolation"] = string(isolation)
	// Tells the sessions of this test apart in pg_stat_activity.
	config.ConnConfig.RuntimeParams["application_name"] = s.Name
	if role != "" {
		config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, "SET ROLE "+pgx.Identifier{role}.Sanitize())
			return err
		}
	}
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// Admin connects to the schema as the superuser.
func (s *Schema) Admin(t testing.TB) *pgxpool.Pool {
	t.Helper()
	return s.Pool(t, "", ReadCommitted)
}

// WaitForLocks waits until n sessions of the pools of this schema are blocked
// on a lock, which lets a test order two transactions without sleeping.
func (s *Schema) WaitForLocks(t testing.TB, n int) {
	t.Helper()
	admin := s.Admin(t)
	for range 1000 {
		var waiting int
		if err := admin.QueryRow(t.Context(),
			"SELECT count(*) FROM pg_stat_activity WHERE application_name = $1 AND wait_event_type = 'Lock'", s.Name).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting >= n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%d sessions never blocked on a lock", n)
}

// createRoles creates the roles that do not exist yet. The lock keeps the
// test packages, which run in parallel, from creating the same role twice.
func createRoles(t testing.TB, admin *pgxpool.Pool) {
	t.Helper()
	ctx := t.Context()
	tx, err := admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", rolesLockKey); err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{AppRole, MigrationsRole} {
		var exists bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT FROM pg_roles WHERE rolname = $1)", role).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists {
			continue
		}
		if _, err := tx.Exec(ctx, "CREATE ROLE "+pgx.Identifier{role}.Sanitize()+" NOLOGIN"); err != nil {
			t.Fatalf("create role %s: %v", role, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func connect(t testing.TB, url string) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(t.Context(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(t.Context()); err != nil {
		t.Fatalf("connect to %s: %v", urlVariable, err)
	}
	return pool
}
