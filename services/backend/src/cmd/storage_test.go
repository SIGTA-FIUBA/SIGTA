package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	"backend/src/migrations"
	"backend/src/migrations/dbtest"
	"backend/src/procedure/adapters/disk"
	proceduredomain "backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

// TestCheckDatabase refuses a schema that is not migrated, and then the role
// that owns the tables. The cases of the role are in the migrations package.
func TestCheckDatabase(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewEmptySchema(t)
	app := schema.MemberPool(t, dbtest.AppRole, dbtest.ReadCommitted)
	owner := schema.MemberPool(t, dbtest.MigrationsRole, dbtest.ReadCommitted)
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
	migrator, err := connect(ctx, schema.MemberURL(t, dbtest.MigrationsRole), migrationsRole)
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
	server, err := connect(ctx, schema.MemberURL(t, dbtest.AppRole), appRole)
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

type startedOn struct {
	stores stores
	started
}

// startOn starts as run does on the database of schema, and closes the stores
// when the test ends.
func startOn(t *testing.T, schema *dbtest.Schema, files output.Files) startedOn {
	t.Helper()
	cfg := config{env: envProduction, allowedDomain: "fi.uba.ar", databaseURL: schema.MemberURL(t, dbtest.AppRole)}
	logger := slog.New(slog.DiscardHandler)
	s, closeStores, err := openStores(t.Context(), cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closeStores)
	launched, err := start(t.Context(), cfg, s, files, nil, logger)
	if err != nil {
		t.Fatal(err)
	}
	return startedOn{s, launched}
}

// TestAStartAddsOnlyTheSecretariat starts on a database just migrated: nobody
// exists until they sign in with Google, so the only rows are the Secretariat
// and its platform permissions.
func TestAStartAddsOnlyTheSecretariat(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewSchema(t)
	files, err := disk.NewFiles(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	startOn(t, schema, files)

	admin := schema.Admin(t)
	rows, err := admin.Query(ctx, "SELECT tablename::text FROM pg_tables WHERE schemaname = $1 AND tablename <> 'schema_migrations'", schema.Name)
	if err != nil {
		t.Fatal(err)
	}
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	secretariat := map[string]string{
		"areas":            "id = 'secretariat' AND system",
		"area_permissions": "area_id = 'secretariat' AND system",
	}
	for _, table := range tables {
		others := "true"
		if condition, ok := secretariat[table]; ok {
			others = "NOT (" + condition + ")"
		}
		var count int
		if err := admin.QueryRow(ctx, fmt.Sprintf("SELECT count(*) FROM %s WHERE %s", pgx.Identifier{schema.Name, table}.Sanitize(), others)).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("%s has %d rows besides the Secretariat after a start, want none", table, count)
		}
	}
	var areas int
	if err := admin.QueryRow(ctx, fmt.Sprintf("SELECT count(*) FROM %s", pgx.Identifier{schema.Name, "areas"}.Sanitize())).Scan(&areas); err != nil || areas != 1 {
		t.Errorf("areas after a start = %d, %v; want the Secretariat", areas, err)
	}
}

// TestTheStoresOutliveTheServer starts twice on the same database, as a
// restart does: the second start finds the session, the memberships, the
// procedure and the document of the first.
func TestTheStoresOutliveTheServer(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewSchema(t)
	files, err := disk.NewFiles(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	first := startOn(t, schema, files)
	for _, m := range []authdomain.Membership{
		{ID: "m-proposer", UserID: "proposer", Area: authinput.SecretariatArea, Permissions: []string{authinput.EditProcedureTypes}, From: time.Now().Add(-time.Hour)},
		{ID: "m-approver", UserID: "approver", Area: authinput.SecretariatArea, Permissions: []string{authinput.PublishProcedureTypes}, From: time.Now().Add(-time.Hour)},
	} {
		if err := first.stores.memberships.Save(ctx, m); err != nil {
			t.Fatal(err)
		}
	}
	if err := first.stores.sessions.Save(ctx, authdomain.Session{TokenHash: "hash", UserID: "applicant", LastSeenAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	published, err := first.procedureTypes.Publish(ctx, input.Publication{
		ProcedureType: 1,
		Authorship:    version.Authorship{ProposedBy: "proposer", ApprovedBy: "approver"},
		Definition: definition.ProcedureTypeDefinition{
			Name:        "Constancia",
			Description: "Un trámite inventado para la prueba.",
			Documents:   []definition.Document{{ID: "statement", Name: "Constancia", AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: 1 << 20}},
			Steps:       []definition.Step{{ID: "documents", Name: "Documentación", Role: definition.RoleApplicant, Action: definition.ActionUploadDocument, Documents: []string{"statement"}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	startedAt := time.Now().UTC().Truncate(time.Microsecond)
	if err := first.stores.procedures.Add(ctx, output.Start{Procedure: proceduredomain.Procedure{
		ID: "procedure", ProcedureType: 1, Version: published.Number, VersionDigest: published.Digest,
		Applicant: "applicant", OpenStep: "documents", StartedAt: startedAt, MovedAt: startedAt,
	}}); err != nil {
		t.Fatal(err)
	}
	attached, err := first.documents.Attach(ctx, "applicant", input.Upload{ProcedureID: "procedure", Requirement: "statement", FileName: "constancia.pdf", Content: strings.NewReader("%PDF-1.7\n")})
	if err != nil {
		t.Fatal(err)
	}

	second := startOn(t, schema, files)
	if session, err := second.stores.sessions.Get(ctx, "hash"); err != nil || session.UserID != "applicant" {
		t.Errorf("session after the restart = %+v, %v; want the one of applicant", session, err)
	}
	if again, err := second.stores.memberships.ListByUser(ctx, "proposer"); err != nil || len(again) != 1 {
		t.Errorf("memberships of proposer after the restart = %+v, %v; want the one of the first start", again, err)
	}
	all, err := second.stores.procedures.List(ctx)
	if err != nil || len(all) != 1 {
		t.Errorf("procedures after the restart = %d, %v; want the one of the first start", len(all), err)
	}
	document, file, err := second.documents.Open(ctx, "applicant", "procedure", attached.ID)
	if err != nil {
		t.Fatalf("open the document after the restart: %v", err)
	}
	defer func() { _ = file.Close() }()
	if content, err := io.ReadAll(file); err != nil || string(content) != "%PDF-1.7\n" || document.FileName != "constancia.pdf" {
		t.Errorf("document after the restart = %s with %q, %v; want constancia.pdf with its content", document.FileName, content, err)
	}
}
