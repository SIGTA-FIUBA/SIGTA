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

			type result struct {
				applied int
				err     error
			}
			results := make(chan result, 2)
			for range 2 {
				go func() {
					applied, err := migrations.Apply(context.WithoutCancel(ctx), owner)
					results <- result{applied, err}
				}()
			}
			schema.WaitForLocks(t, 2)
			if err := holder.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			first, second := <-results, <-results
			if first.err != nil || second.err != nil {
				t.Fatalf("Apply() = %v and %v", first.err, second.err)
			}
			if total, want := first.applied+second.applied, embedded(t); total != want || min(first.applied, second.applied) != 0 {
				t.Errorf("the runs applied %d and %d migrations, want %d and 0", first.applied, second.applied, want)
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

func TestMigrationsThatEndTheTransaction(t *testing.T) {
	for _, tc := range []struct {
		name, sql string
		want      error
	}{
		{"a commit", "CREATE TABLE a (id integer);\nCOMMIT;\nCREATE TABLE b (id integer);", migrations.ErrTransactionControl},
		{"a commit that opens another transaction", "CREATE TABLE a (id integer);\nCOMMIT;\nBEGIN;", migrations.ErrTransactionControl},
		{"a rollback", "ROLLBACK;", migrations.ErrTransactionControl},
		{"the words in comments, strings and a function", "-- no COMMIT here\nCREATE TABLE \"commit\" (id integer);\nCOMMENT ON TABLE \"commit\" IS 'begin, rollback';\n" +
			"CREATE FUNCTION f() RETURNS trigger AS $body$ BEGIN RETURN NEW; END $body$ LANGUAGE plpgsql;", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			owner := dbtest.NewEmptySchema(t).Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
			_, err := migrations.ApplyFrom(t.Context(), owner, fstest.MapFS{"0001_a.sql": {Data: []byte(tc.sql)}})
			if !errors.Is(err, tc.want) {
				t.Errorf("Apply() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestMigrationOutOfOrder(t *testing.T) {
	ctx := t.Context()
	owner := dbtest.NewEmptySchema(t).Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	if _, err := migrations.ApplyFrom(ctx, owner, fstest.MapFS{"0002_b.sql": {}}); err != nil {
		t.Fatal(err)
	}
	_, err := migrations.ApplyFrom(ctx, owner, fstest.MapFS{"0001_a.sql": {}, "0002_b.sql": {}})
	if !errors.Is(err, migrations.ErrOutOfOrder) {
		t.Errorf("Apply() of a migration numbered before an applied one = %v, want ErrOutOfOrder", err)
	}
}

func TestMigrationNames(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files []string
		valid bool
	}{
		{"numbered and described", []string{"0001_procedures.sql", "0002_add_areas_2.sql"}, true},
		{"without a description", []string{"0001.sql"}, false},
		{"with three digits", []string{"001_procedures.sql"}, false},
		{"with upper case", []string{"0001_Procedures.sql"}, false},
		{"two with the same number", []string{"0002_areas.sql", "0002_documents.sql"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{}
			for _, file := range tc.files {
				fsys[file] = &fstest.MapFile{}
			}
			_, err := migrations.Load(fsys)
			if tc.valid && err != nil {
				t.Errorf("Load() = %v, want nil", err)
			}
			if !tc.valid && !errors.Is(err, migrations.ErrInvalidName) {
				t.Errorf("Load() = %v, want ErrInvalidName", err)
			}
		})
	}
}
