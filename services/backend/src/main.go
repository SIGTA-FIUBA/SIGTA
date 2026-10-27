package main

import (
	"application"
	"dev"
	"domain"
	"log/slog"
	"memory"
	"net/http"
	"os"
	"rest"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Only the development adapters exist so far; the institutional OpenID
	// Connect adapter comes with SIG-61.
	if os.Getenv("SIGTA_ENV") != "development" {
		logger.Error("no production identity provider yet; run with SIGTA_ENV=development")
		os.Exit(1)
	}

	publicURL := envOr("SIGTA_PUBLIC_URL", "http://localhost:8080")
	appURL := envOr("SIGTA_APP_URL", "http://localhost:5173/")

	auth := application.NewAuthService(
		application.AuthConfig{
			IdleTimeout:   30 * time.Minute,
			LoginLinkTTL:  15 * time.Minute,
			AllowedDomain: envOr("SIGTA_ALLOWED_DOMAIN", "fi.uba.ar"),
			LoginLinkURL:  appURL + "login/link#token=",
		},
		dev.NewIdentityProvider(publicURL+"/auth/callback"),
		memory.NewUserRepository(devUsers()...),
		memory.NewSessionStore(),
		memory.NewLoginLinkStore(),
		dev.NewMailer(logger),
		time.Now,
	)

	api := rest.NewServer(rest.NewAuthHandler(auth, appURL, false))

	router := http.NewServeMux()
	rest.HandlerFromMux(rest.NewStrictHandler(api, nil), router)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// devUsers are synthetic accounts, one per role, for local development.
func devUsers() []domain.User {
	return []domain.User{
		{ID: "dev-student", Email: "alumno@fi.uba.ar", Name: "Alumno de prueba", Role: domain.RoleApplicant},
		{ID: "dev-operative", Email: "operativo@fi.uba.ar", Name: "Operativo de prueba", Role: domain.RoleStaffOperative},
		{ID: "dev-approver", Email: "aprobacion@fi.uba.ar", Name: "Aprobación de prueba", Role: domain.RoleStaffApprover},
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
