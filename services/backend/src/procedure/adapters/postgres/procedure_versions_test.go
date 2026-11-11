package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	"backend/src/migrations/dbtest"
	"backend/src/procedure/adapters/postgres"
	"backend/src/procedure/application"
	"backend/src/procedure/domain/definition/definitiontest"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
	"backend/src/procedure/ports/output/outputtest"
)

func TestProcedureVersions(t *testing.T) {
	for _, isolation := range []dbtest.Isolation{dbtest.ReadCommitted, dbtest.Serializable} {
		t.Run(string(isolation), func(t *testing.T) {
			outputtest.TestProcedureVersions(t, func(t *testing.T) output.ProcedureVersions {
				return postgres.NewProcedureVersions(dbtest.NewSchema(t).Pool(t, dbtest.AppRole, isolation))
			})
		})
	}
}

// TestPublicationThatLosesTheRace forces each way PostgreSQL reports that a
// rival published the same number: a unique violation when Add waits on the
// rival and it commits, no row inserted when the rival committed first, and a
// serialization failure when Add waits on the rival at SERIALIZABLE. The last
// one is forced with version 1, whose procedure type both of them insert,
// because ON CONFLICT DO NOTHING fails that way on a row it cannot see.
func TestPublicationThatLosesTheRace(t *testing.T) {
	for _, tc := range []struct {
		name              string
		isolation         dbtest.Isolation
		number            int
		rivalCommitsFirst bool
		code              string
	}{
		{"waits on the rival", dbtest.ReadCommitted, 2, false, "23505"},
		{"the rival already committed", dbtest.ReadCommitted, 2, true, ""},
		{"waits on the rival for the first version", dbtest.ReadCommitted, 1, false, ""},
		{"waits on the rival at serializable", dbtest.Serializable, 1, false, "40001"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			schema := dbtest.NewSchema(t)
			pool := schema.Pool(t, dbtest.AppRole, tc.isolation)
			store := postgres.NewProcedureVersions(pool)
			if tc.number > 1 {
				if err := store.Add(ctx, outputtest.Version(1, 1, "Primera")); err != nil {
					t.Fatal(err)
				}
			}

			rival, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = rival.Rollback(ctx) }()
			won := outputtest.Version(1, tc.number, "La del rival")
			if err := insertAsRival(ctx, rival, won); err != nil {
				t.Fatal(err)
			}

			loser := outputtest.Version(1, tc.number, "La que pierde")
			var lost error
			if tc.rivalCommitsFirst {
				if err := rival.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				lost = store.Add(ctx, loser)
			} else {
				result := make(chan error, 1)
				go func() { result <- store.Add(context.WithoutCancel(ctx), loser) }()
				schema.WaitForLocks(t, 1)
				if err := rival.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				lost = <-result
			}

			if !errors.Is(lost, output.ErrVersionExists) {
				t.Fatalf("Add() = %v, want ErrVersionExists", lost)
			}
			pgErr, isPgErr := errors.AsType[*pgconn.PgError](lost)
			switch {
			case tc.code == "" && isPgErr:
				t.Errorf("Add() = %v, want no error from PostgreSQL", lost)
			case tc.code != "" && (!isPgErr || pgErr.Code != tc.code):
				t.Errorf("Add() = %v, want SQLSTATE %s", lost, tc.code)
			}
			got, err := store.Latest(ctx, 1)
			if err != nil || got.Digest != won.Digest {
				t.Errorf("Latest() = %s, %v, want the version of the rival, %s", got.Digest, err, won.Digest)
			}
		})
	}
}

// insertAsRival inserts v in tx as another publication would, without the
// checks of Add, along with its procedure type if v is its first version.
func insertAsRival(ctx context.Context, tx pgx.Tx, v version.ProcedureTypeVersion) error {
	if v.Number == 1 {
		if _, err := tx.Exec(ctx, "INSERT INTO procedure_types (id) VALUES ($1)", v.ProcedureTypeID); err != nil {
			return err
		}
	}
	_, err := tx.Exec(ctx, "INSERT INTO procedure_type_versions (procedure_type_id, number, digest, content, proposed_by, approved_by, published_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		v.ProcedureTypeID, v.Number, v.Digest, v.EncodedContent, v.Authorship.ProposedBy, v.Authorship.ApprovedBy, v.PublishedAt)
	return err
}

func TestDigestFromSQL(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewSchema(t)
	pool := schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
	store := postgres.NewProcedureVersions(pool)
	v := outputtest.Version(1, 1, "Equivalencia de materias: año, acción, 日本語, 🎓")
	if err := store.Add(ctx, v); err != nil {
		t.Fatal(err)
	}
	var fromSQL string
	if err := pool.QueryRow(ctx, "SELECT 'sha256:' || encode(sha256(content), 'hex') FROM procedure_type_versions WHERE procedure_type_id = 1 AND number = 1").Scan(&fromSQL); err != nil {
		t.Fatal(err)
	}
	if fromSQL != v.Digest {
		t.Errorf("digest from SQL = %s, want %s", fromSQL, v.Digest)
	}
}

// TestServiceOnAnAlteredVersion alters a row as the owner of the table, the
// only role that can, and expects the service to refuse it.
func TestServiceOnAnAlteredVersion(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewSchema(t)
	service := application.NewProcedureTypeService(postgres.NewProcedureVersions(schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted)), areaService(), everyoneHolds{}, time.Now)
	d := definitiontest.Definition()
	authors := version.Authorship{ProposedBy: "ana", ApprovedBy: "beto"}
	published, err := service.Publish(ctx, input.Publication{ProcedureType: 1, Definition: d, Authorship: authors})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := service.Pinned(ctx, published.Pin()); err != nil || got.Digest != published.Digest {
		t.Fatalf("Pinned() = %s, %v, want %s", got.Digest, err, published.Digest)
	}

	owner := schema.Pool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
	if _, err := owner.Exec(ctx, "UPDATE procedure_type_versions SET content = content || ' '::bytea"); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Version(ctx, 1, 1); !errors.Is(err, version.ErrAlteredContent) {
		t.Errorf("Version() = %v, want ErrAlteredContent", err)
	}
	if _, err := service.Pinned(ctx, published.Pin()); !errors.Is(err, version.ErrAlteredContent) {
		t.Errorf("Pinned() = %v, want ErrAlteredContent", err)
	}
	d.Name = "Otro nombre"
	if _, err := service.Publish(ctx, input.Publication{ProcedureType: 1, Base: 1, Definition: d, Authorship: authors}); !errors.Is(err, version.ErrAlteredContent) {
		t.Errorf("Publish() on top of it = %v, want ErrAlteredContent", err)
	}
}

// TestRivalOfAnotherProcedureType runs, against each attempt of Add, a rival
// that publishes a version of another procedure type at SERIALIZABLE: it
// reads its latest number before the attempt inserts and inserts before the
// attempt commits. Their predicate locks cover the same pages, so the rival
// commits and the attempt fails with a serialization failure although its
// number is still free. Add tries again, and if every attempt fails it keeps
// the 40001 and does not say that the version exists.
func TestRivalOfAnotherProcedureType(t *testing.T) {
	// Two retries, so that a couple of publications of other procedure types
	// do not fail Add.
	if postgres.Attempts < 3 {
		t.Fatalf("Add tries %d times, want at least 3", postgres.Attempts)
	}
	for _, tc := range []struct {
		name   string
		rivals int
	}{
		{"in the first attempt", 1},
		{"in every attempt", postgres.Attempts},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			pool := dbtest.NewSchema(t).Pool(t, dbtest.AppRole, dbtest.Serializable)
			store := postgres.NewProcedureVersions(pool)

			rivals, attempts := 0, 0
			var rival pgx.Tx
			startRival := func() {
				var err error
				if rival, err = pool.Begin(ctx); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = rival.Rollback(context.WithoutCancel(ctx)) })
				if _, err := rival.Exec(ctx, "SELECT coalesce(max(number), 0) FROM procedure_type_versions WHERE procedure_type_id = $1", rivals+2); err != nil {
					t.Fatal(err)
				}
			}
			startRival()
			postgres.SetBeforeCommit(store, func(ctx context.Context) {
				attempts++
				if rivals == tc.rivals {
					return
				}
				if err := insertAsRival(ctx, rival, outputtest.Version(int64(rivals+2), 1, "Otro tipo")); err != nil {
					t.Errorf("rival: %v", err)
				}
				if err := rival.Commit(ctx); err != nil {
					t.Errorf("rival: %v", err)
				}
				rivals++
				if rivals < tc.rivals {
					startRival()
				}
			})

			err := store.Add(ctx, outputtest.Version(1, 1, "Primera"))
			if want := min(tc.rivals+1, postgres.Attempts); rivals != tc.rivals || attempts != want {
				t.Errorf("%d rivals published and %d attempts, want %d and %d", rivals, attempts, tc.rivals, want)
			}
			_, getErr := store.Get(ctx, 1, 1)
			if tc.rivals < postgres.Attempts {
				if err != nil || getErr != nil {
					t.Errorf("Add() = %v and then Get() = %v, want both nil", err, getErr)
				}
				return
			}
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != "40001" || errors.Is(err, output.ErrVersionExists) {
				t.Errorf("Add() = %v, want SQLSTATE 40001 and not ErrVersionExists", err)
			}
			if !errors.Is(getErr, output.ErrNoVersion) {
				t.Errorf("Get() = %v, want ErrNoVersion", getErr)
			}
		})
	}
}

// TestApproverThatTheDatabaseFolds checks that the constraint on the authors
// is reported as an invalid authorship. The constraint uses lower(), which
// depends on the locale of the database, and Authorship.Validate uses
// strings.EqualFold, so they can disagree outside ASCII.
func TestApproverThatTheDatabaseFolds(t *testing.T) {
	ctx := t.Context()
	pool := dbtest.NewSchema(t).Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
	store := postgres.NewProcedureVersions(pool)

	v := outputtest.Version(1, 1, "Primera")
	v.Authorship = version.Authorship{ProposedBy: "ana", ApprovedBy: "ANA"}
	err := store.Add(ctx, v)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); !errors.Is(err, version.ErrInvalidAuthorship) || !ok || pgErr.Code != "23514" {
		t.Errorf("Add() = %v, want ErrInvalidAuthorship with SQLSTATE 23514", err)
	}

	authors := version.Authorship{ProposedBy: "İnés", ApprovedBy: "inés"}
	var folds bool
	if err := pool.QueryRow(ctx, "SELECT lower($1) = lower($2)", authors.ProposedBy, authors.ApprovedBy).Scan(&folds); err != nil {
		t.Fatal(err)
	}
	if !folds {
		t.Skip("the locale of the database does not fold İ into i")
	}
	if err := authors.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil: the case needs names that only the database takes for the same person", err)
	}
	service := application.NewProcedureTypeService(store, areaService(), everyoneHolds{}, time.Now)
	if _, err := service.Publish(ctx, input.Publication{ProcedureType: 1, Definition: definitiontest.Definition(), Authorship: authors}); !errors.Is(err, version.ErrInvalidAuthorship) {
		t.Errorf("Publish() = %v, want ErrInvalidAuthorship", err)
	}
}

// areaService offers the areas of definitiontest from auth, where the server
// reads them (ADR-021).
// everyoneHolds lets anyone publish: these tests are about storing versions,
// not about who may publish them.
type everyoneHolds struct{ authinput.StaffAuthorizer }

func (everyoneHolds) Holds(context.Context, string, string, int64, []string) (bool, error) {
	return true, nil
}

func areaService() *authapp.AreaService {
	var areas []authdomain.Area
	for _, a := range definitiontest.Areas() {
		area := authdomain.Area{ID: a.ID, Name: a.Name}
		for _, p := range a.Permissions {
			area.Permissions = append(area.Permissions, authdomain.Permission{ID: p.ID, Name: p.Name})
		}
		areas = append(areas, area)
	}
	return authapp.NewAreaService(authmem.NewAreaRepository(areas...), authmem.NewAdministrators(), authmem.NewMembershipRepository(), time.Now)
}
