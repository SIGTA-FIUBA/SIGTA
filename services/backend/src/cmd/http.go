package main

import (
	"context"
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
	procedureapp "backend/src/procedure/application"
	procedure "backend/src/procedure/domain"
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
	app := api.NewServer(authHandler, procedurerest.NewCatalogHandler(procedureTypes))

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

func devAreas() []procedure.Area {
	return []procedure.Area{
		{ID: "applicant", Name: "Titular", Kind: procedure.AreaApplicant},
		{ID: "office", Name: "Área de prueba", Kind: procedure.AreaOffice},
	}
}

// devProcedureType is a made-up procedure type, so that the catalog has
// something to show in local development. It describes no real process.
func devProcedureType() input.Publication {
	return input.Publication{
		ProcedureType: 1,
		Authorship:    procedure.Authorship{ProposedBy: "dev-operative", ApprovedBy: "dev-approver"},
		Definition: procedure.Definition{
			Name:          "Trámite de ejemplo",
			Description:   "Un trámite inventado para probar SIGTA en desarrollo.",
			Deadlines:     []string{"Plazo de ejemplo: hasta 30 días antes del inicio de la cursada."},
			Admissibility: []string{"Condición de ejemplo: tener una cuenta de prueba."},
			Fields: []procedure.Field{
				{ID: "reason", Name: "Motivo", Type: procedure.FieldLongText},
				{ID: "comment", Name: "Comentario", Type: procedure.FieldLongText},
			},
			Documents: []procedure.Document{
				{ID: "statement", Name: "Constancia de ejemplo", AcceptedTypes: []string{"application/pdf", "image/png"}, MaxSize: 10 << 20},
			},
			Steps: []procedure.Step{
				{ID: "request", Name: "Solicitud", Area: "applicant", Action: procedure.ActionEnterData, Fields: []string{"reason"}, OptionalFields: []string{"comment"}, Next: "documents"},
				{ID: "documents", Name: "Documentación", Area: "applicant", Action: procedure.ActionUploadDocument, Documents: []string{"statement"}, Next: "review"},
				{ID: "review", Name: "Revisión", Area: "office", Profile: procedure.ProfileStaffOperative, Action: procedure.ActionConfirmExternalFact},
			},
		},
	}
}
