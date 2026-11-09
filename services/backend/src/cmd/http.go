package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"backend/src/api"
	"backend/src/auth/adapters/dev"
	"backend/src/auth/adapters/google"
	"backend/src/auth/adapters/memory"
	authrest "backend/src/auth/adapters/rest"
	"backend/src/auth/application"
	"backend/src/auth/domain"
	out "backend/src/auth/ports/output"
	proceduremem "backend/src/procedure/adapters/memory"
	procedurerest "backend/src/procedure/adapters/rest"
	"backend/src/procedure/adapters/roleaccess"
	procedureapp "backend/src/procedure/application"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := loadConfig(os.Getenv)
	if err != nil {
		return err
	}
	users, err := usersOf(cfg)
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
		memory.NewUserRepository(users...),
		memory.NewSessionStore(),
		time.Now,
	)

	authHandler := authrest.NewAuthHandler(auth, cfg.appURL, cfg.loginURL(), cfg.secureCookies())
	procedureTypes := procedureapp.NewProcedureTypeService(proceduremem.NewProcedureVersions(), proceduremem.NewAreaCatalog(devAreas()...), time.Now)
	if cfg.development() {
		if _, err := procedureTypes.Publish(context.Background(), devProcedureType()); err != nil {
			return fmt.Errorf("publish the example procedure type: %w", err)
		}
	}
	summaries := proceduremem.NewProcedureSummaries()
	if cfg.development() {
		for _, summary := range devProcedures(time.Now()) {
			summaries.Add(summary)
		}
	}
	inboxHandler := procedurerest.NewInboxHandler(procedureapp.NewInboxService(roleaccess.Permissions{}, summaries, time.Now), personOf(auth))
	app := api.NewServer(authHandler, procedurerest.NewCatalogHandler(procedureTypes), inboxHandler)

	router := http.NewServeMux()
	api.Mount(router, app, logger)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("listening", "env", cfg.env, "identity_provider", cfg.identityProvider, "users", len(users))
	return server.ListenAndServe()
}

// personOf lets the procedure module ask who has a session, without
// depending on the auth module.
func personOf(auth *application.AuthService) procedurerest.Authenticate {
	return func(ctx context.Context, sessionToken string) (inbox.Person, error) {
		user, err := auth.Authenticate(ctx, sessionToken)
		if errors.Is(err, domain.ErrSessionNotFound) || errors.Is(err, domain.ErrSessionExpired) || errors.Is(err, domain.ErrUnknownUser) {
			return inbox.Person{}, procedurerest.ErrNoSession
		}
		if err != nil {
			return inbox.Person{}, err
		}
		return inbox.Person{ID: user.ID, Role: string(user.Role)}, nil
	}
}

func usersOf(cfg config) ([]domain.User, error) {
	if cfg.usersFile != "" {
		return loadUsers(cfg.usersFile, cfg.allowedDomain)
	}
	// loadConfig asks for a users file everywhere but development.
	return devUsers(), nil
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

// devApplicantEmail is who the development identity provider logs in by
// default.
const devApplicantEmail = "titular@fi.uba.ar"

// devUsers are synthetic accounts, one per role, for local development.
func devUsers() []domain.User {
	return []domain.User{
		{ID: "dev-applicant", Email: devApplicantEmail, Name: "Titular de prueba", Role: domain.RoleApplicant},
		{ID: "dev-operative", Email: "operativo@fi.uba.ar", Name: "Operativo de prueba", Role: domain.RoleStaffOperative},
		{ID: "dev-approver", Email: "aprobacion@fi.uba.ar", Name: "Aprobación de prueba", Role: domain.RoleStaffApprover},
	}
}

func devAreas() []definition.Area {
	return []definition.Area{
		{ID: "applicant", Name: "Titular", Kind: definition.AreaApplicant},
		{ID: "office", Name: "Área de prueba", Kind: definition.AreaOffice},
	}
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
				{ID: "statement", Name: "Constancia de ejemplo", AcceptedTypes: []string{"application/pdf", "image/png"}, MaxSize: 10 << 20},
			},
			Steps: []definition.Step{
				{ID: "request", Name: "Solicitud", Area: "applicant", Action: definition.ActionEnterData, Fields: []string{"reason"}, OptionalFields: []string{"comment"}, Next: "documents"},
				{ID: "documents", Name: "Documentación", Area: "applicant", Action: definition.ActionUploadDocument, Documents: []string{"statement"}, Next: "review"},
				{ID: "review", Name: "Revisión", Area: "office", Profile: definition.ProfileStaffOperative, Action: definition.ActionConfirmExternalFact},
			},
		},
	}
}
