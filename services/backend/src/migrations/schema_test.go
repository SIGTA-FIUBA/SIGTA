package migrations_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/migrations"
	"backend/src/migrations/dbtest"
)

const (
	insufficientPrivilege = "42501"
	foreignKeyViolation   = "23503"
	checkViolation        = "23514"
)

// grant is what the application role may do with a relation: the privileges
// on all of it, and the columns it may update.
type grant struct {
	table     []string
	updatable []string
}

// TestOwnerAndPrivilegesOfEachRelation checks every relation of the schema
// against its owner and what the application role may do. A new table fails
// here until it is listed, so its privileges are a decision and not an
// accident. An append-only table gets SELECT and INSERT and no column to
// update; any other is listed in granted.
func TestOwnerAndPrivilegesOfEachRelation(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewSchema(t)
	app := schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
	addProcedureWithEvent(t, app)

	granted := map[string]grant{
		"schema_migrations":               {table: []string{"SELECT"}},
		"procedure_types":                 {table: []string{"SELECT", "INSERT"}},
		"current_procedure_type_versions": {table: []string{"SELECT"}},
		"procedures":                      {table: []string{"SELECT", "INSERT"}},
	}
	for _, table := range migrations.AppendOnlyTables {
		if _, ok := granted[table]; ok {
			t.Errorf("%s is append-only and also in granted: list it in one place", table)
		}
		granted[table] = grant{table: []string{"SELECT", "INSERT"}}
	}

	relations := relationsOf(t, schema)
	if len(relations) == 0 {
		t.Fatal("the migrations created no table")
	}
	for name, r := range relations {
		if r.owner != dbtest.MigrationsRole {
			t.Errorf("%s belongs to %s, want %s", name, r.owner, dbtest.MigrationsRole)
		}
		if _, ok := granted[name]; !ok {
			t.Errorf("%s is not in this test: add the privileges it should have", name)
		}
	}

	for name, g := range granted {
		t.Run(name, func(t *testing.T) {
			r, ok := relations[name]
			if !ok {
				t.Fatalf("%s is not in the schema", name)
			}
			for _, privilege := range []string{"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"} {
				var got bool
				if err := app.QueryRow(ctx, "SELECT has_table_privilege($1, $2, $3)", dbtest.AppRole, name, privilege).Scan(&got); err != nil {
					t.Fatal(err)
				}
				if want := slices.Contains(g.table, privilege); got != want {
					t.Errorf("%s on %s: %t, want %t", privilege, name, got, want)
				}
			}
			// has_table_privilege misses a grant on some columns only.
			columns := columnsOf(t, app, name)
			for _, column := range columns {
				for _, privilege := range []string{"UPDATE", "REFERENCES"} {
					var got bool
					if err := app.QueryRow(ctx, "SELECT has_column_privilege($1, $2, $3, $4)", dbtest.AppRole, name, column, privilege).Scan(&got); err != nil {
						t.Fatal(err)
					}
					want := slices.Contains(g.table, privilege) || privilege == "UPDATE" && slices.Contains(g.updatable, column)
					if got != want {
						t.Errorf("%s on %s.%s: %t, want %t", privilege, name, column, got, want)
					}
				}
			}
			for _, column := range g.updatable {
				if !slices.Contains(columns, column) {
					t.Errorf("updatable column %s.%s does not exist", name, column)
				}
			}

			statements := []string{"DROP TABLE " + name, "ALTER TABLE " + name + " ADD COLUMN extra text"}
			if r.view {
				statements = []string{"DROP VIEW " + name, "CREATE OR REPLACE VIEW " + name + " AS SELECT 1"}
			} else {
				statements = append(statements, "DELETE FROM "+name, "TRUNCATE "+name)
				for _, column := range columns {
					if !slices.Contains(g.updatable, column) {
						statements = append(statements, "UPDATE "+name+" SET "+column+" = "+column)
					}
				}
			}
			for _, statement := range statements {
				_, err := app.Exec(ctx, statement)
				if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != insufficientPrivilege {
					t.Errorf("%s as the application role = %v, want insufficient privilege", statement, err)
				}
			}
		})
	}

	// The version a procedure follows, said outright so that renaming a
	// column does not drop it from the loop above unnoticed.
	for _, statement := range []string{
		"UPDATE procedures SET version_number = 2",
		"UPDATE procedures SET version_digest = 'sha256:0000'",
		"UPDATE procedures SET id = 'p2'",
	} {
		_, err := app.Exec(ctx, statement)
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != insufficientPrivilege {
			t.Errorf("%s as the application role = %v, want insufficient privilege", statement, err)
		}
	}

	var canCreate bool
	if err := app.QueryRow(ctx, "SELECT has_schema_privilege($1, $2, 'CREATE')", dbtest.AppRole, schema.Name).Scan(&canCreate); err != nil {
		t.Fatal(err)
	}
	if canCreate {
		t.Error("the application role can create tables in the schema")
	}
	if _, err := app.Exec(ctx, "CREATE TABLE extra (id integer)"); err == nil {
		t.Error("the application role created a table")
	}
}

func addProcedureWithEvent(t *testing.T, app *pgxpool.Pool) {
	t.Helper()
	for _, statement := range []string{
		"INSERT INTO procedure_types (id) VALUES (1)",
		"INSERT INTO procedure_type_versions (procedure_type_id, number, digest, content, proposed_by, approved_by, published_at) SELECT 1, 1, 'sha256:' || encode(sha256(c), 'hex'), c, 'ana', 'beto', now() FROM (SELECT '{}'::bytea AS c) AS s",
		"INSERT INTO procedures (id, procedure_type_id, version_number, version_digest) SELECT 'p1', 1, 1, digest FROM procedure_type_versions",
		"INSERT INTO procedure_events (procedure_id, number) VALUES ('p1', 1)",
	} {
		if _, err := app.Exec(t.Context(), statement); err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}
}

type relation struct {
	owner string
	view  bool
}

// relationsOf returns each table, view and sequence of the schema.
func relationsOf(t *testing.T, schema *dbtest.Schema) map[string]relation {
	t.Helper()
	rows, err := schema.Admin(t).Query(t.Context(),
		"SELECT relname, pg_get_userbyid(relowner), relkind = 'v' FROM pg_class WHERE relnamespace = $1::regnamespace AND relkind IN ('r', 'v', 'm', 'S', 'p', 'f')", schema.Name)
	if err != nil {
		t.Fatal(err)
	}
	relations := map[string]relation{}
	var name string
	var r relation
	if _, err := pgx.ForEachRow(rows, []any{&name, &r.owner, &r.view}, func() error {
		relations[name] = r
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return relations
}

func columnsOf(t *testing.T, db *pgxpool.Pool, table string) []string {
	t.Helper()
	rows, err := db.Query(t.Context(),
		"SELECT attname FROM pg_attribute WHERE attrelid = $1::regclass AND attnum > 0 AND NOT attisdropped ORDER BY attnum", table)
	if err != nil {
		t.Fatal(err)
	}
	columns, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	if len(columns) == 0 {
		t.Fatalf("%s has no columns", table)
	}
	return columns
}

func TestCheckApplicationRole(t *testing.T) {
	ctx := t.Context()
	for _, tc := range []struct {
		name  string
		grant string
		role  string
	}{
		{"the owner of the tables", "", dbtest.MigrationsRole},
		{"a superuser", "", ""},
		{"UPDATE on a single column", "GRANT UPDATE (approved_by) ON procedure_type_versions TO " + dbtest.AppRole, dbtest.AppRole},
		{"DELETE", "GRANT DELETE ON procedure_events TO " + dbtest.AppRole, dbtest.AppRole},
		{"TRUNCATE", "GRANT TRUNCATE ON procedure_type_versions TO " + dbtest.AppRole, dbtest.AppRole},
	} {
		t.Run(tc.name, func(t *testing.T) {
			schema := dbtest.NewSchema(t)
			if err := migrations.CheckApplicationRole(ctx, schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted)); err != nil {
				t.Fatalf("CheckApplicationRole() as %s = %v, want nil", dbtest.AppRole, err)
			}
			if tc.grant != "" {
				if _, err := schema.Admin(t).Exec(ctx, tc.grant); err != nil {
					t.Fatal(err)
				}
			}
			err := migrations.CheckApplicationRole(ctx, schema.Pool(t, tc.role, dbtest.ReadCommitted))
			if err == nil || !strings.Contains(err.Error(), "application role") || !strings.Contains(err.Error(), "migrate command") {
				t.Errorf("CheckApplicationRole() = %v, want an error that names the role of the server and the one of migrate", err)
			}
		})
	}
}

func TestNoForeignKeyCascades(t *testing.T) {
	schema := dbtest.NewSchema(t)
	rows, err := schema.Admin(t).Query(t.Context(),
		"SELECT conname FROM pg_constraint WHERE connamespace = $1::regnamespace AND contype = 'f' AND (confdeltype <> 'a' OR confupdtype <> 'a')", schema.Name)
	if err != nil {
		t.Fatal(err)
	}
	cascades, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	if len(cascades) > 0 {
		t.Errorf("foreign keys that change or delete rows on their own: %v, want none", cascades)
	}
}

func TestConstraints(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewSchema(t)
	app := schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
	addProcedureWithEvent(t, app)

	for _, tc := range []struct {
		name, statement, code string
	}{
		{"a procedure with a digest that is not the one of its version",
			"INSERT INTO procedures (id, procedure_type_id, version_number, version_digest) VALUES ('p2', 1, 1, 'sha256:0000')", foreignKeyViolation},
		{"a procedure of a version that does not exist",
			"INSERT INTO procedures (id, procedure_type_id, version_number, version_digest) SELECT 'p2', 1, 2, digest FROM procedure_type_versions", foreignKeyViolation},
		{"an event of a procedure that does not exist",
			"INSERT INTO procedure_events (procedure_id, number) VALUES ('p2', 1)", foreignKeyViolation},
		{"an event number 0",
			"INSERT INTO procedure_events (procedure_id, number) VALUES ('p1', 0)", checkViolation},
		{"a version of a procedure type that does not exist",
			"INSERT INTO procedure_type_versions (procedure_type_id, number, digest, content, proposed_by, approved_by, published_at) VALUES (2, 1, 'sha256:0000', '{}', 'ana', 'beto', now())", foreignKeyViolation},
		{"a version number 0",
			"INSERT INTO procedure_type_versions (procedure_type_id, number, digest, content, proposed_by, approved_by, published_at) VALUES (1, 0, 'sha256:0000', '{}', 'ana', 'beto', now())", checkViolation},
		{"a version approved by whoever proposed it, in other case",
			"INSERT INTO procedure_type_versions (procedure_type_id, number, digest, content, proposed_by, approved_by, published_at) VALUES (1, 2, 'sha256:0000', '{}', 'ana', 'ANA', now())", checkViolation},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := app.Exec(ctx, tc.statement)
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != tc.code {
				t.Errorf("Exec() = %v, want SQLSTATE %s", err, tc.code)
			}
		})
	}

	// The foreign key also stops the owner from deleting a version that a
	// procedure follows with a plain DELETE. It does not stop the owner for
	// good: it can drop the constraint or TRUNCATE ... CASCADE.
	owner := schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	_, err := owner.Exec(ctx, "DELETE FROM procedure_type_versions")
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != foreignKeyViolation {
		t.Errorf("deleting a version that a procedure follows = %v, want a foreign key violation", err)
	}
}

// TestRolesOfTheDatabase checks what deploy/postgres/sql/roles.sql leaves in
// the database of SIGTA_TEST_DATABASE_URL. A cluster made with initdb has no
// backend schema, so the test runs only where roles.sql ran, as in the
// continuous integration. A database created before the migrations role
// existed fails here.
func TestRolesOfTheDatabase(t *testing.T) {
	admin, err := pgxpool.New(t.Context(), dbtest.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	var owner string
	var appCanUse, appCanCreate bool
	err = admin.QueryRow(t.Context(),
		"SELECT pg_get_userbyid(nspowner), has_schema_privilege($1, oid, 'USAGE'), has_schema_privilege($1, oid, 'CREATE') FROM pg_namespace WHERE nspname = 'backend'",
		dbtest.AppRole).Scan(&owner, &appCanUse, &appCanCreate)
	if errors.Is(err, pgx.ErrNoRows) && !dbtest.InCI() {
		t.Skip("the database has no backend schema: deploy/postgres/sql/roles.sql did not run on it")
	}
	if err != nil {
		t.Fatal(err)
	}
	if owner != dbtest.MigrationsRole {
		t.Errorf("the backend schema belongs to %s, want %s; %s", owner, dbtest.MigrationsRole, migrations.OldDatabaseHint)
	}
	if !appCanUse || appCanCreate {
		t.Errorf("the application role can use the backend schema: %t, create in it: %t; want true and false", appCanUse, appCanCreate)
	}
	for _, role := range []string{dbtest.AppRole, dbtest.MigrationsRole} {
		var canLogin, isSuperuser, canCreateInPublic bool
		err := admin.QueryRow(t.Context(),
			"SELECT rolcanlogin, rolsuper, has_schema_privilege(oid, 'public', 'CREATE') FROM pg_roles WHERE rolname = $1",
			role).Scan(&canLogin, &isSuperuser, &canCreateInPublic)
		if err != nil {
			t.Fatalf("role %s: %v; %s", role, err, migrations.OldDatabaseHint)
		}
		if !canLogin || isSuperuser {
			t.Errorf("role %s can log in: %t, is superuser: %t; want true and false; %s", role, canLogin, isSuperuser, migrations.OldDatabaseHint)
		}
		// PostgreSQL 14 and older let every role create in public.
		if canCreateInPublic {
			t.Errorf("role %s can create in the public schema, want false", role)
		}
	}
}
