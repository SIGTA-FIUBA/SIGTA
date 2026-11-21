package migrations

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OldDatabaseHint explains the errors of a database created before the
// migrations role existed: deploy/postgres/sql/roles.sql runs only when the
// volume of the database is created.
const OldDatabaseHint = "if the database was created before the migrations role existed, recreate it with docker compose down --volumes"

// ErrPrivileged means the role of the server can change what the schema
// promises to keep: a published version, an event, a document, a value
// entered, the version a procedure follows, who started it or a revision of a
// role.
var ErrPrivileged = errors.New("the role can change rows that must not change")

// immutable lists the tables, and their columns, whose stored values the
// application role must never change. A table without columns is append-only:
// none of its columns. No migration grants UPDATE on them, nor DELETE or
// TRUNCATE on the table.
var immutable = []struct {
	table   string
	columns []string
}{
	{table: "procedure_type_versions"},
	{table: "procedure_events"},
	{table: "documents"},
	{table: "procedure_field_values"},
	{table: "procedure_correction_answers"},
	{table: "area_role_revisions"},
	{table: "procedures", columns: []string{"procedure_type_id", "version_number", "version_digest", "applicant", "started_at"}},
}

// CheckApplicationRole fails with ErrPrivileged if the role of pool, the
// login it connected with, or a role either of them can become, can change an
// immutable value: if it is a superuser, can create roles, owns the table or
// its schema, or has UPDATE on one of the columns, DELETE or TRUNCATE. The
// login counts because SET ROLE does not drop its privileges: RESET ROLE
// gives them back.
func CheckApplicationRole(ctx context.Context, pool *pgxpool.Pool) error {
	for _, t := range immutable {
		var role string
		err := pool.QueryRow(ctx, `SELECT m.rolname
			FROM pg_roles m, pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE c.oid = $1::regclass
				AND (pg_has_role(current_user, m.oid, 'MEMBER') OR pg_has_role(session_user, m.oid, 'MEMBER')) AND (
				m.rolsuper OR m.rolcreaterole OR m.oid IN (c.relowner, n.nspowner)
				OR has_table_privilege(m.oid, c.oid, 'DELETE')
				OR has_table_privilege(m.oid, c.oid, 'TRUNCATE')
				OR coalesce(cardinality($2::text[]), 0) = 0 AND has_any_column_privilege(m.oid, c.oid, 'UPDATE')
				OR EXISTS (SELECT FROM unnest($2::text[]) col WHERE has_column_privilege(m.oid, c.oid, col, 'UPDATE')))
			ORDER BY m.rolname LIMIT 1`, t.table, t.columns).Scan(&role)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			continue
		case err != nil:
			return fmt.Errorf("check the privileges on %s: %w", t.table, err)
		}
		return fmt.Errorf("%w: %s can change or delete the rows of %s", ErrPrivileged, role, t.table)
	}
	return nil
}
