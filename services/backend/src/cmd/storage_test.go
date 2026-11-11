package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"backend/src/migrations"
	"backend/src/migrations/dbtest"
)

// TestCheckDatabase refuses a schema that is not migrated, and then the role
// that owns the tables. The cases of the role are in the migrations package.
func TestCheckDatabase(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewEmptySchema(t)
	app := schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
	owner := schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	if err := checkDatabase(ctx, app); !errors.Is(err, migrations.ErrPending) {
		t.Fatalf("checkDatabase() before migrating = %v, want ErrPending", err)
	}
	if _, err := migrations.Apply(ctx, owner); err != nil {
		t.Fatal(err)
	}
	if err := checkDatabase(ctx, app); err != nil {
		t.Fatalf("checkDatabase() as %s = %v, want nil", dbtest.AppRole, err)
	}
	if err := checkDatabase(ctx, owner); err == nil || !strings.Contains(err.Error(), "application role") {
		t.Errorf("checkDatabase() as %s = %v, want an error that names the role of the server", dbtest.MigrationsRole, err)
	}
}

func TestMigrateNeedsItsURL(t *testing.T) {
	err := migrateCommand(env(map[string]string{"SIGTA_DATABASE_URL": "postgres://backend@db/sigta"}), slog.New(slog.DiscardHandler))
	if err == nil || !strings.Contains(err.Error(), "SIGTA_MIGRATIONS_DATABASE_URL") {
		t.Fatalf("migrateCommand() = %v, want an error about SIGTA_MIGRATIONS_DATABASE_URL", err)
	}
}

// An old database keeps the schema of the application role, or lacks the
// migrations role, since roles.sql runs only when its volume is created.
func TestMigrateHintsAnOldDatabase(t *testing.T) {
	ctx := t.Context()
	for _, tc := range []struct {
		name  string
		usage bool
		code  string
	}{
		{"schema of another role", false, invalidSchemaName},
		{"schema of another role that it can use", true, insufficientPrivilege},
	} {
		t.Run(tc.name, func(t *testing.T) {
			schema := dbtest.NewEmptySchema(t)
			admin := schema.Admin(t)
			statements := []string{fmt.Sprintf("ALTER SCHEMA %s OWNER TO %s", schema.Name, dbtest.AppRole)}
			if tc.usage {
				statements = append(statements, fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s", schema.Name, dbtest.MigrationsRole))
			}
			for _, statement := range statements {
				if _, err := admin.Exec(ctx, statement); err != nil {
					t.Fatal(err)
				}
			}
			_, err := applyMigrations(ctx, schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted))
			assertOldDatabaseHint(t, err, tc.code)
		})
	}
	t.Run("missing role", func(t *testing.T) {
		u, err := url.Parse(dbtest.URL(t))
		if err != nil {
			t.Fatal(err)
		}
		u.User = url.User("sigta_missing_role")
		err = migrateCommand(env(map[string]string{"SIGTA_MIGRATIONS_DATABASE_URL": u.String()}), slog.New(slog.DiscardHandler))
		assertOldDatabaseHint(t, err, invalidAuthorization)
	})
}

func assertOldDatabaseHint(t *testing.T, err error, code string) {
	t.Helper()
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != code {
		t.Errorf("migrate = %v, want SQLSTATE %s", err, code)
	}
	if err == nil || !strings.Contains(err.Error(), migrations.OldDatabaseHint) {
		t.Errorf("migrate = %v, want the hint to recreate the database", err)
	}
}
