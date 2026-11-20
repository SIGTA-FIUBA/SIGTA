package main

import (
	"errors"
	"log/slog"
	"strings"
	"testing"

	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	"backend/src/migrations"
	"backend/src/migrations/dbtest"
)

func TestGrantAdminNeedsAnEmailAndItsURL(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	withURL := env(map[string]string{"SIGTA_MIGRATIONS_DATABASE_URL": "postgres://backend_migrations@db/sigta"})
	for _, args := range [][]string{nil, {" "}, {"ana@fi.uba.ar", "beto@fi.uba.ar"}, {"--by=beto@fi.uba.ar", "ana@fi.uba.ar"}, {"--help"}} {
		if err := grantAdminCommand(withURL, args, logger); err == nil || !strings.Contains(err.Error(), "usage") {
			t.Errorf("grantAdminCommand(%q) = %v, want the usage", args, err)
		}
	}
	err := grantAdminCommand(env(map[string]string{"SIGTA_DATABASE_URL": "postgres://backend@db/sigta"}), []string{"ana@fi.uba.ar"}, logger)
	if err == nil || !strings.Contains(err.Error(), "SIGTA_MIGRATIONS_DATABASE_URL") {
		t.Errorf("grantAdminCommand() = %v, want an error about SIGTA_MIGRATIONS_DATABASE_URL", err)
	}
}

// The administrator that grant-admin adds is one for the server, which
// connects as the application role. It adds only the first.
func TestGrantAdminMakesAnAdministratorForTheServer(t *testing.T) {
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)
	schema := dbtest.NewSchema(t)
	cfg := config{env: envProduction, allowedDomain: "fi.uba.ar", databaseURL: schema.MemberURL(t, dbtest.AppRole)}
	s, closeStores, err := openStores(ctx, cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(closeStores)
	ana, err := s.users.Register(ctx, authdomain.User{ID: "U-1", Email: "ana@fi.uba.ar", Name: "Ana"})
	if err != nil {
		t.Fatal(err)
	}
	beto, err := s.users.Register(ctx, authdomain.User{ID: "U-2", Email: "beto@fi.uba.ar", Name: "Beto"})
	if err != nil {
		t.Fatal(err)
	}
	owner := env(map[string]string{"SIGTA_MIGRATIONS_DATABASE_URL": schema.MemberURL(t, dbtest.MigrationsRole)})

	if err := grantAdminCommand(owner, []string{"carla@fi.uba.ar"}, logger); !errors.Is(err, authapp.ErrUserNotSignedIn) {
		t.Errorf("grantAdminCommand() of an email nobody signed in with = %v, want ErrUserNotSignedIn", err)
	}
	if err := grantAdminCommand(owner, []string{"ANA@fi.uba.ar"}, logger); err != nil {
		t.Fatalf("grantAdminCommand() = %v", err)
	}
	if err := grantAdminCommand(owner, []string{"beto@fi.uba.ar"}, logger); !errors.Is(err, authapp.ErrAdministratorExists) {
		t.Errorf("grantAdminCommand() with an administrator = %v, want ErrAdministratorExists", err)
	}
	for id, want := range map[string]bool{ana.ID: true, beto.ID: false} {
		if got, err := s.administrators.IsAdministrator(ctx, id); err != nil || got != want {
			t.Errorf("IsAdministrator(%s) = %v, %v; want %v", id, got, err, want)
		}
	}
}

func TestGrantAdminNeedsAMigratedSchema(t *testing.T) {
	schema := dbtest.NewEmptySchema(t)
	owner := env(map[string]string{"SIGTA_MIGRATIONS_DATABASE_URL": schema.MemberURL(t, dbtest.MigrationsRole)})
	if err := grantAdminCommand(owner, []string{"ana@fi.uba.ar"}, slog.New(slog.DiscardHandler)); !errors.Is(err, migrations.ErrPending) {
		t.Errorf("grantAdminCommand() before migrating = %v, want ErrPending", err)
	}
}

func TestCommandsCheckTheirArguments(t *testing.T) {
	for _, args := range [][]string{{"grant-admin"}, {"migrate", "grant-admin", "ana@fi.uba.ar"}} {
		if err := command(args, slog.New(slog.DiscardHandler)); err == nil || !strings.Contains(err.Error(), "usage: backend "+args[0]) {
			t.Errorf("command(%q) = %v, want its usage", args, err)
		}
	}
}

func TestAnUnknownCommandDoesNotStartTheServer(t *testing.T) {
	if err := command([]string{"grant-admins", "ana@fi.uba.ar"}, slog.New(slog.DiscardHandler)); err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("command(grant-admins) = %v, want an unknown command", err)
	}
}
