package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/src/api"
	"backend/src/auth/adapters/google"
	authrest "backend/src/auth/adapters/rest"
	"backend/src/auth/application"
	authoutput "backend/src/auth/ports/output"
	procedurerest "backend/src/procedure/adapters/rest"
	procedureapp "backend/src/procedure/application"
	"backend/src/procedure/ports/output"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := command(os.Args[1:], logger); err != nil {
		name := "server"
		if len(os.Args) > 1 {
			name = os.Args[1]
		}
		logger.Error(name+" failed", "error", err)
		os.Exit(1)
	}
}

// command runs the server, or the command that args names. A command it
// does not know fails, so that a mistyped one does not start a server.
func command(args []string, logger *slog.Logger) error {
	if len(args) == 0 {
		return run(logger)
	}
	switch args[0] {
	case "migrate":
		if len(args) > 1 {
			return errors.New("usage: backend migrate")
		}
		return migrateCommand(os.Getenv, logger)
	case "grant-admin":
		return grantAdminCommand(os.Getenv, args[1:], logger)
	}
	return fmt.Errorf("unknown command %q: usage: backend [migrate | grant-admin <email>]", args[0])
}

func run(logger *slog.Logger) error {
	// Taken before anything else, so that a SIGTERM while the server waits for
	// the database also runs the deferred functions instead of killing it.
	stop, cancelStop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelStop()
	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		return err
	}
	identity, err := googleIdentityProvider(cfg, logger)
	if err != nil {
		return err
	}

	// Bounds the start against the database, so one that does not answer
	// stops it.
	startup, cancel := context.WithTimeout(stop, 30*time.Second)
	defer cancel()
	stores, closeStores, err := openStores(startup, cfg, logger)
	if err != nil {
		return stoppedDuringStartup(stop, err)
	}
	defer closeStores()
	files, err := openFiles(cfg, logger)
	if err != nil {
		return err
	}
	started, err := start(startup, cfg, stores, files, identity, logger)
	if err != nil {
		return stoppedDuringStartup(stop, err)
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           api.Mount(http.NewServeMux(), started.server, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       api.ReadTimeout,
	}
	logger.Info("listening", "env", cfg.env)
	return serve(stop, server, server.ListenAndServe)
}

type started struct {
	server         api.Server
	procedureTypes *procedureapp.ProcedureTypeService
	documents      *procedureapp.DocumentService
}

// start wires the services over the stores and removes the files that no
// document names. It adds nothing: the Secretariat comes with the stores
// (openStores), the first administrator from the grant-admin command, and
// people, memberships and procedure types from those who use SIGTA.
func start(ctx context.Context, cfg config, stores stores, files output.Files, identity authoutput.IdentityProvider, logger *slog.Logger) (started, error) {
	areaService := application.NewAreaService(stores.areas, stores.administrators, stores.memberships, now)
	memberships := application.NewMembershipService(stores.areas, stores.administrators, stores.memberships, stores.users, newID, now)

	procedureTypes := procedureapp.NewProcedureTypeService(stores.versions, areaService, memberships, now)
	drafts := procedureapp.NewProcedureTypeDraftService(stores.drafts, procedureTypes, memberships, now)

	auth := application.NewAuthenticatorService(
		application.AuthConfig{
			IdleTimeout:   30 * time.Minute,
			AllowedDomain: cfg.allowedDomain,
		},
		identity,
		stores.users,
		stores.sessions,
		now,
	)

	authHandler := authrest.NewAuthHandler(auth, memberships, cfg.appURL, cfg.loginURL(), cfg.secureCookies())
	procedures := procedureapp.NewProcedureService(procedureTypes, stores.procedures, stores.documents, memberships, newID, now)
	documents := procedureapp.NewDocumentService(procedures, stores.documents, files, newID, now)
	people := application.NewPeopleService(stores.users, stores.administrators)
	// Once, at startup: the server assumes it is the only instance.
	removed, err := documents.RemoveUnreferencedFiles(ctx)
	if err != nil {
		return started{}, fmt.Errorf("remove the files that no document names: %w", err)
	}
	if removed > 0 {
		logger.Info("removed the files that no document names", "files", removed)
	}
	server := api.NewServer(
		authHandler,
		procedurerest.NewCatalogHandler(procedureTypes),
		authrest.NewAdministrationHandler(auth, areaService, memberships, application.NewAdministratorService(stores.administrators, stores.users, newID, now), people),
		procedurerest.NewDocumentHandler(auth, people, documents),
		procedurerest.NewProcedureHandler(auth, people, procedureapp.NewProcedureDetailService(procedures, stores.documents), procedureapp.NewSubmissionService(procedures, files, newID, now), procedures),
		procedurerest.NewInboxHandler(auth, people, procedureapp.NewInboxService(procedures)),
		procedurerest.NewProcedureTypeDraftHandler(auth, drafts),
	)
	return started{server, procedureTypes, documents}, nil
}

// stoppedDuringStartup returns nil instead of err if the server was asked to
// stop while it started, since then err is only the interrupted wait.
func stoppedDuringStartup(stop context.Context, err error) error {
	if stop.Err() != nil {
		return nil
	}
	return err
}

// shutdownTimeout is how long the requests in flight have to finish once the
// server is asked to stop. Docker waits 10 seconds before it kills it.
const shutdownTimeout = 8 * time.Second

// serve runs listen until stop is done, and then lets the requests in flight
// finish before it returns, so that they do not lose the database that the
// deferred functions of run close.
func serve(stop context.Context, server *http.Server, listen func() error) error {
	failed := make(chan error, 1)
	go func() { failed <- listen() }()
	select {
	case err := <-failed:
		return err
	case <-stop.Done():
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shut down: %w", err)
	}
	if err := <-failed; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func newID() string {
	return rand.Text()
}

func googleIdentityProvider(cfg config, logger *slog.Logger) (authoutput.IdentityProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return google.NewIdentityProvider(ctx, google.Config{
		Issuer:       google.Issuer,
		ClientID:     cfg.googleClient.id,
		ClientSecret: cfg.googleClient.secret,
		RedirectURL:  cfg.callbackURL(),
		HostedDomain: cfg.allowedDomain,
	}, logger)
}

// now is the time as PostgreSQL keeps it, in UTC and to the microsecond, so
// that what a request answers right after saving says the same as a later
// read.
func now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}
