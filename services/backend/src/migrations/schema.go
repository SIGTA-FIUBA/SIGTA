package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OldDatabaseHint explains the errors of a database created before the
// migrations role existed: deploy/postgres/sql/roles.sql runs only when the
// volume of the database is created.
const OldDatabaseHint = "if the database was created before the migrations role existed, recreate it with docker compose down --volumes"

// appendOnlyTables are the tables whose rows the application role must never
// change or delete. No migration grants it UPDATE, DELETE or TRUNCATE on them.
var appendOnlyTables = []string{"procedure_type_versions", "procedure_events"}

// CheckApplicationRole fails if the role of pool can change or delete the
// rows of an append-only table: if it owns the table, as the migrations role
// or a superuser does, or has UPDATE on any of its columns, DELETE or
// TRUNCATE.
func CheckApplicationRole(ctx context.Context, pool *pgxpool.Pool) error {
	for _, table := range appendOnlyTables {
		var role string
		var canChange bool
		err := pool.QueryRow(ctx, `SELECT current_user,
			pg_has_role(current_user, relowner, 'USAGE')
			OR has_any_column_privilege(current_user, oid, 'UPDATE')
			OR has_table_privilege(current_user, oid, 'DELETE')
			OR has_table_privilege(current_user, oid, 'TRUNCATE')
			FROM pg_class WHERE oid = $1::regclass`, table).Scan(&role, &canChange)
		if err != nil {
			return fmt.Errorf("check the privileges on %s: %w", table, err)
		}
		if canChange {
			return fmt.Errorf("SIGTA_DATABASE_URL connects as %s, which can change or delete the rows of %s: the server connects as the application role (backend) and only the migrate command uses the role that owns the tables (backend_migrations)", role, table)
		}
	}
	return nil
}
