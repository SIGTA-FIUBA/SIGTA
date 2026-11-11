package migrations_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgconn"

	"backend/src/migrations"
	"backend/src/migrations/dbtest"
)

func TestApplyAndCheck(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewEmptySchema(t)
	owner := schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	app := schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
	if err := migrations.Check(ctx, app); !errors.Is(err, migrations.ErrPending) {
		t.Fatalf("Check() of an empty schema = %v, want ErrPending", err)
	}
	_, err := migrations.Apply(ctx, app)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != insufficientPrivilege {
		t.Fatalf("Apply() as the application role = %v, want insufficient privilege", err)
	}
	for i, want := range []int{embedded(t), 0} {
		if applied, err := migrations.Apply(ctx, owner); err != nil || applied != want {
			t.Fatalf("Apply() number %d = %d, %v, want %d migrations applied", i+1, applied, err, want)
		}
	}
	if err := migrations.Check(ctx, app); err != nil {
		t.Fatalf("Check() as the application role = %v", err)
	}
}

// embedded returns how many migrations the binary has.
func embedded(t *testing.T) int {
	t.Helper()
	all, err := migrations.Load(migrations.Files)
	if err != nil {
		t.Fatal(err)
	}
	return len(all)
}

// TestConcurrentApply blocks two runs on the lock of the migrations and then
// lets them go: the first one migrates and the second finds nothing to do,
// even if the default isolation of the database is stricter.
func TestConcurrentApply(t *testing.T) {
	for _, isolation := range []dbtest.Isolation{dbtest.ReadCommitted, dbtest.Serializable} {
		t.Run(string(isolation), func(t *testing.T) {
			ctx := t.Context()
			schema := dbtest.NewEmptySchema(t)
			owner := schema.Pool(t, dbtest.MigrationsRole, isolation)

			holder, err := schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted).Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = holder.Rollback(ctx) }()
			if _, err := holder.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", migrations.LockKey); err != nil {
				t.Fatal(err)
			}

			results := make(chan error, 2)
			for range 2 {
				go func() {
					_, err := migrations.Apply(context.WithoutCancel(ctx), owner)
					results <- err
				}()
			}
			schema.WaitForLocks(t, 2)
			if err := holder.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			for range 2 {
				if err := <-results; err != nil {
					t.Errorf("Apply() = %v", err)
				}
			}

			var applied int
			if err := owner.QueryRow(ctx, "SELECT count(*) FROM schema_migrations").Scan(&applied); err != nil {
				t.Fatal(err)
			}
			if want := embedded(t); applied != want {
				t.Errorf("schema_migrations has %d rows, want %d", applied, want)
			}
		})
	}
}

func TestCheckRefusesAnotherSchema(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change string
		want   error
	}{
		{"a migration that was not applied", "DELETE FROM schema_migrations WHERE name = '0001_procedures.sql'", migrations.ErrPending},
		{"a migration edited after it was applied", "UPDATE schema_migrations SET checksum = 'edited'", migrations.ErrChanged},
		{"a migration this binary does not have", "INSERT INTO schema_migrations (name, checksum) VALUES ('9999_newer.sql', 'x')", migrations.ErrUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			schema := dbtest.NewSchema(t)
			owner := schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
			if _, err := owner.Exec(ctx, tc.change); err != nil {
				t.Fatal(err)
			}
			err := migrations.Check(ctx, schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted))
			if !errors.Is(err, tc.want) {
				t.Errorf("Check() = %v, want %v", err, tc.want)
			}
			if !errors.Is(tc.want, migrations.ErrPending) {
				if errors.Is(err, migrations.ErrPending) {
					t.Errorf("Check() = %v, want an error that is not ErrPending", err)
				}
				if _, err := migrations.Apply(ctx, owner); !errors.Is(err, tc.want) {
					t.Errorf("Apply() = %v, want %v", err, tc.want)
				}
			}
		})
	}
}

func TestMigrationsThatControlTheTransaction(t *testing.T) {
	for _, tc := range []struct {
		name, sql string
		valid     bool
	}{
		{"a commit", "CREATE TABLE a (id integer);\nCOMMIT;\nCREATE TABLE b (id integer);", false},
		{"a begin in lower case", "begin;\nCREATE TABLE a (id integer);", false},
		{"a rollback", "ROLLBACK;", false},
		{"a start transaction", "START  TRANSACTION;", false},
		{"an index built concurrently", "CREATE INDEX CONCURRENTLY a_id ON a (id);", false},
		{"the words in comments and strings", "-- no COMMIT here\n/* nor a ROLLBACK */\nCOMMENT ON TABLE a IS 'begin, commit';\nCREATE TABLE \"commit\" (id integer);", true},
		{"the body of a function", "CREATE FUNCTION f() RETURNS trigger AS $body$ BEGIN RETURN NEW; END $body$ LANGUAGE plpgsql;\nSELECT $$commit$$;", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := migrations.Load(fstest.MapFS{"0001_a.sql": {Data: []byte(tc.sql)}})
			if tc.valid && err != nil {
				t.Errorf("Load() = %v, want nil", err)
			}
			if !tc.valid && !errors.Is(err, migrations.ErrTransactionControl) {
				t.Errorf("Load() = %v, want ErrTransactionControl", err)
			}
		})
	}
}
