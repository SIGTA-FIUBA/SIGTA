package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

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
	if err := checkDatabase(ctx, owner); !errors.Is(err, migrations.ErrPrivileged) || !strings.Contains(err.Error(), "migrate command") {
		t.Errorf("checkDatabase() as %s = %v, want ErrPrivileged and the role of the migrate command", dbtest.MigrationsRole, err)
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
			_, err := applyMigrations(ctx, schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted), true)
			assertOldDatabaseHint(t, err, tc.code)
		})
	}
	t.Run("missing role", func(t *testing.T) {
		u, err := url.Parse(dbtest.URL(t))
		if err != nil {
			t.Fatal(err)
		}
		u.User = url.User("sigta_missing_role")
		err = migrateCommand(env(map[string]string{"SIGTA_MIGRATIONS_DATABASE_URL": u.String(), "SIGTA_ENV": envDevelopment}), slog.New(slog.DiscardHandler))
		// Without the role, PostgreSQL answers 28P01 when it asks for a password.
		assertOldDatabaseHint(t, err, invalidAuthorization, invalidPassword)
	})
	// Elsewhere the same codes come from a wrong password or a missing grant.
	t.Run("outside development", func(t *testing.T) {
		schema := dbtest.NewEmptySchema(t)
		if _, err := schema.Admin(t).Exec(ctx, fmt.Sprintf("ALTER SCHEMA %s OWNER TO %s", schema.Name, dbtest.AppRole)); err != nil {
			t.Fatal(err)
		}
		_, err := applyMigrations(ctx, schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted), false)
		if err == nil || strings.Contains(err.Error(), migrations.OldDatabaseHint) {
			t.Errorf("migrate = %v, want an error without the hint to recreate the database", err)
		}
	})
}

func TestActAsDefaultsToTheBackendSchema(t *testing.T) {
	for _, tc := range []struct{ url, want string }{
		{"postgres://sigta@db/sigta", backendSchema},
		{"postgres://sigta@db/sigta?search_path=other", "other"},
	} {
		config, err := pgxpool.ParseConfig(tc.url)
		if err != nil {
			t.Fatal(err)
		}
		actAs(config, appRole)
		if got := config.ConnConfig.RuntimeParams["search_path"]; got != tc.want {
			t.Errorf("search_path of %s = %q, want %q", tc.url, got, tc.want)
		}
	}
}

// TestLoginsThatAreMembersOfTheRoles migrates and serves with logins that are
// members of the roles of roles.sql and not the roles themselves: the tables
// still belong to the migrations role, and the server passes its checks.
func TestLoginsThatAreMembersOfTheRoles(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewEmptySchema(t)
	admin := schema.Admin(t)
	login := func(role string) string {
		name := "sigta_" + schema.Name + "_" + role
		if _, err := admin.Exec(ctx, fmt.Sprintf("CREATE ROLE %s LOGIN PASSWORD 'member' IN ROLE %s", pgx.Identifier{name}.Sanitize(), role)); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			// DROP OWNED takes what the login created if it did not act as role.
			quoted := pgx.Identifier{name}.Sanitize()
			if _, err := admin.Exec(context.WithoutCancel(ctx), "DROP OWNED BY "+quoted+"; DROP ROLE "+quoted); err != nil {
				t.Error(err)
			}
		})
		u, err := url.Parse(dbtest.URL(t))
		if err != nil {
			t.Fatal(err)
		}
		u.User = url.UserPassword(name, "member")
		query := u.Query()
		query.Set("search_path", schema.Name)
		u.RawQuery = query.Encode()
		return u.String()
	}
	migrator, err := connect(ctx, login(dbtest.MigrationsRole), migrationsRole)
	if err != nil {
		t.Fatal(err)
	}
	defer migrator.Close()
	if _, err := applyMigrations(ctx, migrator, false); err != nil {
		t.Fatalf("migrate as a member of %s: %v", dbtest.MigrationsRole, err)
	}
	var owners []string
	rows, err := admin.Query(ctx, "SELECT DISTINCT tableowner::text FROM pg_tables WHERE schemaname = $1", schema.Name)
	if err != nil {
		t.Fatal(err)
	}
	if owners, err = pgx.CollectRows(rows, pgx.RowTo[string]); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(owners, []string{dbtest.MigrationsRole}) {
		t.Errorf("the tables belong to %v, want only %s", owners, dbtest.MigrationsRole)
	}
	server, err := connect(ctx, login(dbtest.AppRole), appRole)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if err := checkDatabase(ctx, server); err != nil {
		t.Errorf("checkDatabase() as a member of %s = %v, want nil", dbtest.AppRole, err)
	}
}

func assertOldDatabaseHint(t *testing.T, err error, codes ...string) {
	t.Helper()
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || !slices.Contains(codes, pgErr.Code) {
		t.Errorf("migrate = %v, want SQLSTATE %s", err, strings.Join(codes, " or "))
	}
	if err == nil || !strings.Contains(err.Error(), migrations.OldDatabaseHint) {
		t.Errorf("migrate = %v, want the hint to recreate the database", err)
	}
}
