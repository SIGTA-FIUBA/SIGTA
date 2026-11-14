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
	"backend/src/auth/adapters/dev"
	"backend/src/auth/adapters/google"
	"backend/src/auth/adapters/memory"
	authrest "backend/src/auth/adapters/rest"
	"backend/src/auth/application"
	"backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	out "backend/src/auth/ports/output"
	proceduremem "backend/src/procedure/adapters/memory"
	procedurerest "backend/src/procedure/adapters/rest"
	procedureapp "backend/src/procedure/application"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := migrateCommand(os.Getenv, logger); err != nil {
			logger.Error("migrations failed", "error", err)
			os.Exit(1)
		}
		return
	}
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
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
	ctx := context.Background()
	areaRepository := memory.NewAreaRepository(devAreas()...)
	if err := application.EnsureSecretariat(ctx, areaRepository); err != nil {
		return err
	}
	areas, err := areaRepository.List(ctx)
	if err != nil {
		return err
	}
	directory, err := peopleOf(cfg, areas)
	if err != nil {
		return err
	}
	identity, err := identityProviderOf(cfg, logger)
	if err != nil {
		return err
	}

	auth := application.NewAuthService(
		application.AuthConfig{
			IdleTimeout:   30 * time.Minute,
			AllowedDomain: cfg.allowedDomain,
		},
		identity,
		memory.NewUserRepository(directory.users...),
		memory.NewSessionStore(),
		time.Now,
	)

	administrators := memory.NewAdministrators(directory.administrators...)
	membershipRepository := memory.NewMembershipRepository(directory.memberships...)
	areaService := application.NewAreaService(areaRepository, administrators, membershipRepository, time.Now)
	memberships := application.NewMembershipService(areaRepository, administrators, membershipRepository, newID, time.Now)

	authHandler := authrest.NewAuthHandler(auth, memberships, cfg.appURL, cfg.loginURL(), cfg.secureCookies())
	// Bounds the start against the database, so one that does not answer
	// stops it.
	startup, cancel := context.WithTimeout(stop, 30*time.Second)
	defer cancel()
	versions, closeVersions, err := openProcedureVersions(startup, cfg, logger)
	if err != nil {
		return stoppedDuringStartup(stop, err)
	}
	defer closeVersions()
	procedureTypes := procedureapp.NewProcedureTypeService(versions, areaService, memberships, time.Now)
	if cfg.development() {
		if err := publishDevProcedureType(startup, versions, procedureTypes); err != nil {
			return stoppedDuringStartup(stop, fmt.Errorf("publish the example procedure type: %w", err))
		}
	}
	files, err := openFiles(cfg, logger)
	if err != nil {
		return err
	}
	// The procedures and what is recorded about their documents stay in
	// memory until they are stored in PostgreSQL.
	procedures := procedureapp.NewProcedureService(procedureTypes, proceduremem.NewProcedureRepository(), memberships, newID)
	documents := procedureapp.NewDocumentService(procedures, proceduremem.NewDocumentRepository(), files, newID, time.Now)
	app := api.NewServer(
		authHandler,
		procedurerest.NewCatalogHandler(procedureTypes),
		authrest.NewAdministrationHandler(auth, areaService, memberships),
		procedurerest.NewDocumentHandler(auth, documents),
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           api.Mount(http.NewServeMux(), app, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("listening", "env", cfg.env, "identity_provider", cfg.identityProvider, "users", len(directory.users))
	return serve(stop, server, server.ListenAndServe)
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

func peopleOf(cfg config, areas []domain.Area) (people, error) {
	if cfg.usersFile != "" {
		return loadUsers(cfg.usersFile, cfg.allowedDomain, areas)
	}
	// loadConfig asks for a users file everywhere but development.
	return devPeople(), nil
}

func newID() string {
	return rand.Text()
}

func identityProviderOf(cfg config, logger *slog.Logger) (out.IdentityProvider, error) {
	if cfg.identityProvider == identityDev {
		return dev.NewIdentityProvider(cfg.callbackURL(), devApplicantEmail), nil
	}
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

const devApplicantEmail = "titular@fi.uba.ar"

// devPeople are synthetic accounts for local development: an applicant, a
// member who reviews and edits procedure types, one who also approves and
// publishes them, and an administrator.
func devPeople() people {
	return people{
		users: []domain.User{
			{ID: "dev-applicant", Email: devApplicantEmail, Name: "Titular de prueba"},
			{ID: "dev-operative", Email: "operativo@fi.uba.ar", Name: "Operativo de prueba"},
			{ID: "dev-approver", Email: "aprobacion@fi.uba.ar", Name: "Aprobación de prueba"},
			{ID: "dev-admin", Email: "admin@fi.uba.ar", Name: "Administración de prueba"},
		},
		administrators: []string{"dev-admin"},
		memberships: []domain.Membership{
			{ID: "dev-operative-1", UserID: "dev-operative", Area: "office", Permissions: []string{"review"}, GrantedBy: "dev-admin"},
			{ID: "dev-approver-1", UserID: "dev-approver", Area: "office", Permissions: []string{"review", "approve"}, GrantedBy: "dev-admin"},
			{ID: "dev-operative-2", UserID: "dev-operative", Area: authinput.SecretariatArea, Permissions: []string{authinput.EditProcedureTypes}, GrantedBy: "dev-admin"},
			{ID: "dev-approver-2", UserID: "dev-approver", Area: authinput.SecretariatArea, Permissions: []string{authinput.PublishProcedureTypes}, GrantedBy: "dev-admin"},
		},
	}
}

// devAreas are the areas of staff until they are stored in PostgreSQL, besides
// the Secretariat, which EnsureSecretariat adds.
func devAreas() []domain.Area {
	return []domain.Area{{ID: "office", Name: "Área de prueba", Permissions: []domain.Permission{
		{ID: "review", Name: "Revisar"},
		{ID: "approve", Name: "Aprobar"},
	}}}
}

// publishDevProcedureType publishes the example only if its procedure type
// has no version. A database that has one keeps it even if the example
// changed since, so editing the example does not add a version on every
// start. If another start publishes first, publishing again finds the same
// content and returns the stored version; another content is an error.
func publishDevProcedureType(ctx context.Context, versions output.ProcedureVersions, procedureTypes input.ProcedureTypes) error {
	publication := devProcedureType()
	_, err := versions.Latest(ctx, publication.ProcedureType)
	if !errors.Is(err, output.ErrNoVersion) {
		return err
	}
	_, err = procedureTypes.Publish(ctx, publication)
	if errors.Is(err, input.ErrOutdatedBase) {
		_, err = procedureTypes.Publish(ctx, publication)
	}
	return err
}

// devProcedureType is a made-up procedure type, so that the catalog has
// something to show in local development. It describes no real process.
func devProcedureType() input.Publication {
	return input.Publication{
		ProcedureType: 1,
		Authorship:    version.Authorship{ProposedBy: "dev-operative", ApprovedBy: "dev-approver"},
		Definition: definition.ProcedureTypeDefinition{
			Name:          "Trámite de ejemplo",
			Description:   "Un trámite inventado para probar SIGTA en desarrollo.",
			Deadlines:     []string{"Plazo de ejemplo: hasta 30 días antes del inicio de la cursada."},
			Admissibility: []string{"Condición de ejemplo: tener una cuenta de prueba."},
			Fields: []definition.Field{
				{ID: "reason", Name: "Motivo", Type: definition.FieldLongText},
				{ID: "comment", Name: "Comentario", Type: definition.FieldLongText},
			},
			Documents: []definition.Document{
				{ID: "statement", Name: "Constancia de ejemplo", AcceptedTypes: []string{"application/pdf", "image/png"}, MaxSizeBytes: 10 << 20},
			},
			Steps: []definition.Step{
				{ID: "request", Name: "Solicitud", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Fields: []string{"reason"}, OptionalFields: []string{"comment"}, Next: "documents"},
				{ID: "documents", Name: "Documentación", Role: definition.RoleApplicant, Action: definition.ActionUploadDocument, Documents: []string{"statement"}, Next: "review"},
				{ID: "review", Name: "Revisión", Role: definition.RoleStaff, Area: "office", Permissions: []string{"review"}, Action: definition.ActionConfirmExternalFact},
			},
		},
	}
}
