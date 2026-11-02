package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/src/adapters/dev"
	"backend/src/adapters/memory"
	"backend/src/adapters/rest"
	"backend/src/application"
	"backend/src/domain"
	"backend/src/domain/procedure"
	"backend/src/ports/input"
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

	secureCookies := strings.HasPrefix(publicURL, "https://")
	procedureTypes := application.NewProcedureTypeService(memory.NewProcedureVersions(), memory.NewAreaCatalog(devAreas()...), time.Now)
	// The check above already stops production, but the example must not
	// reach the public catalog when that check goes away with SIG-61.
	if os.Getenv("SIGTA_ENV") == "development" {
		if _, err := procedureTypes.Publish(context.Background(), devProcedureType()); err != nil {
			logger.Error("publish the example procedure type", "error", err)
			os.Exit(1)
		}
	}

	api := rest.NewServer(rest.NewAuthHandler(auth, appURL, secureCookies), rest.NewCatalogHandler(procedureTypes))

	router := http.NewServeMux()
	rest.Mount(router, api, logger)

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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
