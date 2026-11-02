package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/src/adapters/dev"
	"backend/src/adapters/google"
	"backend/src/adapters/memory"
	"backend/src/adapters/rest"
	"backend/src/application"
	"backend/src/domain"
	out "backend/src/ports/output"
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
			LoginLinkTTL:  15 * time.Minute,
			AllowedDomain: cfg.allowedDomain,
			LoginLinkURL:  cfg.appURL + "login/link#token=",
		},
		identity,
		memory.NewUserRepository(users...),
		memory.NewSessionStore(),
		memory.NewLoginLinkStore(),
		mailerOf(cfg, logger),
		time.Now,
	)

	secureCookies := strings.HasPrefix(cfg.publicURL, "https://")
	api := rest.NewServer(rest.NewAuthHandler(auth, cfg.appURL, cfg.loginURL(), secureCookies))

	router := http.NewServeMux()
	rest.Mount(router, api, logger)

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
		return dev.NewIdentityProvider(cfg.callbackURL(), devUsers()[0].Email), nil
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

// errNoMailer answers a login link request outside development: there is no
// mail adapter yet, and whether login links stay at all is pending with the
// Secretaría.
var errNoMailer = errors.New("login links are not available outside development")

type noMailer struct{}

func (noMailer) SendLoginLink(context.Context, string, string) error { return errNoMailer }

func mailerOf(cfg config, logger *slog.Logger) out.Mailer {
	if cfg.development() {
		return dev.NewMailer(logger)
	}
	return noMailer{}
}

// devUsers are synthetic accounts, one per role, for local development.
func devUsers() []domain.User {
	return []domain.User{
		{ID: "dev-student", Email: "alumno@fi.uba.ar", Name: "Alumno de prueba", Role: domain.RoleApplicant},
		{ID: "dev-operative", Email: "operativo@fi.uba.ar", Name: "Operativo de prueba", Role: domain.RoleStaffOperative},
		{ID: "dev-approver", Email: "aprobacion@fi.uba.ar", Name: "Aprobación de prueba", Role: domain.RoleStaffApprover},
	}
}
