package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"backend/src/adapters/dev"
	"backend/src/adapters/filesystem"
	"backend/src/adapters/memory"
	"backend/src/adapters/rest"
	"backend/src/application"
	"backend/src/domain"
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

	store, err := filesystem.NewDocumentStore(envOr("SIGTA_DOCUMENTS_DIR", "data/documents"))
	if err != nil {
		logger.Error("cannot open the document storage", "error", err)
		os.Exit(1)
	}

	// The procedures are a stand-in until the procedure model (SIG-12) and
	// authorization (SIG-14) exist.
	procedures := memory.NewProcedures(devProcedures()...)
	documents := application.NewDocumentService(
		application.DocumentConfig{UnlinkedRule: pdfUpTo10MB},
		procedures,
		procedures,
		memory.NewDocumentRepository(),
		store,
		time.Now,
	)

	greeter := &application.GreeterService{}
	api := rest.NewServer(
		rest.NewGreeterHandler(greeter),
		rest.NewAuthHandler(auth, appURL, false),
		rest.NewDocumentHandler(auth, documents),
	)

	router := http.NewServeMux()
	rest.HandlerFromMux(rest.NewStrictHandler(api, nil), router)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
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

var pdfUpTo10MB = domain.DocumentRule{AcceptedTypes: []string{"application/pdf"}, MaxSize: 10 << 20}

// devProcedures are synthetic procedures for local development.
func devProcedures() []memory.Procedure {
	return []memory.Procedure{{
		ID:           "dev-procedure",
		ApplicantID:  "dev-student",
		Requirements: map[string]domain.DocumentRule{"transcript": pdfUpTo10MB},
	}}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
