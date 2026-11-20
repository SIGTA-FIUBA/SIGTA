package postgres_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"backend/src/auth/adapters/postgres"
	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
	"backend/src/auth/ports/output/outputtest"
	"backend/src/migrations/dbtest"
)

func TestAdministratorRepository(t *testing.T) {
	outputtest.TestAdministratorRepository(t, func(t *testing.T) output.FirstAdministratorRepository {
		return postgres.NewAdministratorRepository(dbtest.NewSchema(t).Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted))
	})
}

// The server grants and ends administrators as the application role.
func TestAdministratorPeriodRepository(t *testing.T) {
	outputtest.TestAdministratorPeriodRepository(t, func(t *testing.T) output.AdministratorPeriodRepository {
		return postgres.NewAdministratorRepository(dbtest.NewSchema(t).Pool(t, dbtest.AppRole, dbtest.ReadCommitted))
	})
}

// The application role only ends an open period: it cannot reopen one nor
// move when it ended.
func TestApplicationRoleOnlyEndsOpenPeriods(t *testing.T) {
	schema := dbtest.NewSchema(t)
	owner := schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	if _, err := owner.Exec(t.Context(), `INSERT INTO administrators (id, user_id, valid_from, valid_until, grant_source, granted_by, ended_by)
		VALUES ('A-1', 'U-1', now() - interval '2 days', now() - interval '1 day', 'administrator', 'U-3', 'U-3')`); err != nil {
		t.Fatal(err)
	}
	app := schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
	for _, statement := range []string{
		"UPDATE administrators SET valid_until = NULL, ended_by = NULL WHERE id = 'A-1'",
		"UPDATE administrators SET valid_until = now() WHERE id = 'A-1'",
	} {
		if tag, err := app.Exec(t.Context(), statement); err != nil || tag.RowsAffected() != 0 {
			t.Errorf("%s as the application role = %v, %v; want no row changed", statement, tag, err)
		}
	}
}

// The first administrator is the initial grant of the system: nobody typed
// who granted it.
func TestGrantFirstIsTheInitialGrant(t *testing.T) {
	ctx := t.Context()
	pool := dbtest.NewSchema(t).Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	at := time.Date(2026, 11, 20, 13, 0, 0, 0, time.UTC)
	if _, err := postgres.NewAdministratorRepository(pool).GrantFirst(ctx, domain.Administrator{ID: "A-1", UserID: "U-1", From: at}); err != nil {
		t.Fatal(err)
	}
	var from time.Time
	var source string
	var by *string
	if err := pool.QueryRow(ctx, "SELECT valid_from, grant_source, granted_by FROM administrators WHERE id = 'A-1' AND user_id = 'U-1' AND valid_until IS NULL").Scan(&from, &source, &by); err != nil {
		t.Fatal(err)
	}
	if !from.Equal(at) || source != "bootstrap" || by != nil {
		t.Errorf("administrator U-1 from %v, %s, granted by %v; want %v, bootstrap, by nobody", from, source, by, at)
	}
}

// An ended administrator keeps its row, with who ended it, but administers
// nothing: the first administrator is the first one without an end.
func TestAnEndedAdministratorDoesNotCount(t *testing.T) {
	ctx := t.Context()
	pool := dbtest.NewSchema(t).Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	if _, err := pool.Exec(ctx, `INSERT INTO administrators (id, user_id, valid_from, valid_until, grant_source, granted_by, ended_by)
		VALUES ('A-1', 'U-1', now() - interval '2 days', now() - interval '1 day', 'administrator', 'U-3', 'U-3')`); err != nil {
		t.Fatal(err)
	}
	r := postgres.NewAdministratorRepository(pool)
	if got, err := r.IsAdministrator(ctx, "U-1"); err != nil || got {
		t.Errorf("IsAdministrator() of an ended administrator = %v, %v; want false", got, err)
	}
	if granted, err := r.GrantFirst(ctx, domain.Administrator{ID: "A-2", UserID: "U-2", From: time.Now()}); err != nil || !granted {
		t.Errorf("GrantFirst() with only an ended administrator = %v, %v; want true", granted, err)
	}
}

// Only the initial grant has no author, an end always has one, and a period
// ends after it starts.
func TestEveryAdministratorButTheInitialHasWhoGrantedIt(t *testing.T) {
	pool := dbtest.NewSchema(t).Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	for _, values := range []string{
		"('A-1', 'U-1', now(), NULL, 'administrator', NULL, NULL)",
		"('A-1', 'U-1', now(), NULL, 'bootstrap', 'U-2', NULL)",
		"('A-1', 'U-1', now(), now(), 'administrator', 'U-2', NULL)",
		"('A-1', 'U-1', now(), NULL, 'administrator', 'U-2', 'U-2')",
		"('A-1', 'U-1', now(), NULL, 'console', NULL, NULL)",
		"('A-1', 'U-1', now(), NULL, 'console', 'U-2', NULL)",
		"('A-1', 'U-1', now(), now() - interval '1 second', 'administrator', 'U-2', 'U-2')",
	} {
		_, err := pool.Exec(t.Context(), "INSERT INTO administrators (id, user_id, valid_from, valid_until, grant_source, granted_by, ended_by) VALUES "+values)
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != "23514" {
			t.Errorf("INSERT %s = %v, want a check violation", values, err)
		}
	}
}

// A person has at most one period without an end.
func TestAPersonHasOneOpenPeriod(t *testing.T) {
	pool := dbtest.NewSchema(t).Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	if _, err := pool.Exec(t.Context(), "INSERT INTO administrators (id, user_id, valid_from, grant_source) VALUES ('A-1', 'U-1', now(), 'bootstrap')"); err != nil {
		t.Fatal(err)
	}
	_, err := pool.Exec(t.Context(), "INSERT INTO administrators (id, user_id, valid_from, grant_source, granted_by) VALUES ('A-2', 'U-1', now(), 'administrator', 'U-2')")
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != "23505" {
		t.Errorf("a second open period = %v, want a unique violation", err)
	}
}

// A period that has not started yet administers nothing.
func TestAFutureAdministratorDoesNotCount(t *testing.T) {
	pool := dbtest.NewSchema(t).Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	if _, err := pool.Exec(t.Context(), "INSERT INTO administrators (id, user_id, valid_from, grant_source) VALUES ('A-1', 'U-1', now() + interval '1 day', 'bootstrap')"); err != nil {
		t.Fatal(err)
	}
	if got, err := postgres.NewAdministratorRepository(pool).IsAdministrator(t.Context(), "U-1"); err != nil || got {
		t.Errorf("IsAdministrator() before the period starts = %v, %v; want false", got, err)
	}
}

// The server connects as the application role, which grants administrators
// only as an administrator: the initial grant stays with grant-admin.
func TestApplicationRoleCannotGrantAdministrators(t *testing.T) {
	schema := dbtest.NewSchema(t)
	app := postgres.NewAdministratorRepository(schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted))
	_, err := app.GrantFirst(t.Context(), domain.Administrator{ID: "A-1", UserID: "U-1", From: time.Now()})
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != "42501" {
		t.Errorf("GrantFirst() as the application role = %v, want insufficient privilege", err)
	}
	if got, err := app.IsAdministrator(t.Context(), "U-1"); err != nil || got {
		t.Errorf("IsAdministrator() as the application role = %v, %v; want false", got, err)
	}
}
