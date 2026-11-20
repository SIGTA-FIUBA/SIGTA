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
	"sync"
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
	authoutput "backend/src/auth/ports/output"
	proceduremem "backend/src/procedure/adapters/memory"
	procedurerest "backend/src/procedure/adapters/rest"
	procedureapp "backend/src/procedure/application"
	proceduredomain "backend/src/procedure/domain"
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

	administrators := memory.NewAdministratorRepository(directory.administrators...)
	membershipRepository := memory.NewMembershipRepository(directory.memberships...)
	areaService := application.NewAreaService(areaRepository, administrators, membershipRepository, time.Now)
	memberships := application.NewMembershipService(areaRepository, administrators, membershipRepository, newID, time.Now)

	// Bounds the start against the database, so one that does not answer
	// stops it.
	startup, cancel := context.WithTimeout(stop, 30*time.Second)
	defer cancel()
	versions, closeVersions, err := openProcedureTypeVersions(startup, cfg, logger)
	if err != nil {
		return stoppedDuringStartup(stop, err)
	}
	defer closeVersions()
	procedureTypes := procedureapp.NewProcedureTypeService(versions, areaService, memberships, time.Now)
	// The procedures and what is recorded about their documents stay in
	// memory until they are stored in PostgreSQL.
	procedureRepository := proceduremem.NewProcedureRepository()
	users := authoutput.UserRepository(memory.NewUserRepository(directory.users...))
	if cfg.development() {
		users, err = addDevExample(startup, cfg, logger, versions, procedureTypes, procedureRepository, users)
		if err != nil {
			return stoppedDuringStartup(stop, err)
		}
		if err := addDevInbox(startup, versions, procedureRepository, time.Now()); err != nil {
			return stoppedDuringStartup(stop, fmt.Errorf("add the procedures of the inbox: %w", err))
		}
	}

	auth := application.NewAuthenticatorService(
		application.AuthConfig{
			IdleTimeout:   30 * time.Minute,
			AllowedDomain: cfg.allowedDomain,
		},
		identity,
		users,
		memory.NewSessionRepository(),
		time.Now,
	)

	authHandler := authrest.NewAuthHandler(auth, memberships, cfg.appURL, cfg.loginURL(), cfg.secureCookies())
	files, err := openFiles(cfg, logger)
	if err != nil {
		return err
	}
	documentRepository := proceduremem.NewDocumentRepository()
	procedures := procedureapp.NewProcedureService(procedureTypes, procedureRepository, documentRepository, memberships, newID, time.Now)
	documents := procedureapp.NewDocumentService(procedures, documentRepository, files, newID, time.Now)
	people := application.NewPeopleService(users)
	// Once, at startup: the server assumes it is the only instance.
	removed, err := documents.RemoveUnreferencedFiles(startup)
	if err != nil {
		return fmt.Errorf("remove the files that no document names: %w", err)
	}
	if removed > 0 {
		logger.Info("removed the files that no document names", "files", removed)
	}
	app := api.NewServer(
		authHandler,
		procedurerest.NewCatalogHandler(procedureTypes),
		authrest.NewAdministrationHandler(auth, areaService, memberships),
		procedurerest.NewDocumentHandler(auth, people, documents),
		procedurerest.NewProcedureHandler(auth, people, procedureapp.NewProcedureDetailService(procedures, documentRepository), procedureapp.NewSubmissionService(procedures, documentRepository, files, newID, time.Now), procedures),
		procedurerest.NewInboxHandler(auth, people, procedureapp.NewInboxService(procedures)),
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           api.Mount(http.NewServeMux(), app, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       api.ReadTimeout,
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

func identityProviderOf(cfg config, logger *slog.Logger) (authoutput.IdentityProvider, error) {
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

const (
	devApplicantID    = "dev-applicant"
	devApplicantEmail = "titular@fi.uba.ar"
)

// devPeople are synthetic accounts for local development: an applicant, a
// member who reviews and edits procedure types, one who also approves and
// publishes them, a member of the committee and an administrator.
func devPeople() people {
	return people{
		users: []domain.User{
			{ID: devApplicantID, Email: devApplicantEmail, Name: "Titular de prueba"},
			{ID: "dev-applicant-01", Email: "lucia.ejemplo@fi.uba.ar", Name: "Lucía Ejemplo"},
			{ID: "dev-applicant-02", Email: "martin.prueba@fi.uba.ar", Name: "Martín Prueba"},
			{ID: "dev-applicant-03", Email: "sofia.inventada@fi.uba.ar", Name: "Sofía Inventada"},
			{ID: "dev-applicant-04", Email: "tomas.ficticio@fi.uba.ar", Name: "Tomás Ficticio"},
			{ID: "dev-applicant-05", Email: "valentina.muestra@fi.uba.ar", Name: "Valentina Muestra"},
			{ID: "dev-applicant-06", Email: "joaquin.ejemplo@fi.uba.ar", Name: "Joaquín Ejemplo"},
			{ID: "dev-operative", Email: "operativo@fi.uba.ar", Name: "Operativo de prueba"},
			{ID: "dev-approver", Email: "aprobacion@fi.uba.ar", Name: "Aprobación de prueba"},
			{ID: "dev-committee", Email: "comision@fi.uba.ar", Name: "Comisión de prueba"},
			{ID: "dev-admin", Email: "admin@fi.uba.ar", Name: "Administración de prueba"},
		},
		administrators: []string{"dev-admin"},
		memberships: []domain.Membership{
			{ID: "dev-operative-1", UserID: "dev-operative", Area: "office", Permissions: []string{"review"}, GrantedBy: "dev-admin"},
			{ID: "dev-approver-1", UserID: "dev-approver", Area: "office", Permissions: []string{"review", "approve"}, GrantedBy: "dev-admin"},
			{ID: "dev-committee-1", UserID: "dev-committee", Area: devCommitteeArea, Permissions: []string{"evaluate"}, GrantedBy: "dev-admin"},
			{ID: "dev-operative-2", UserID: "dev-operative", Area: authinput.SecretariatArea, Permissions: []string{authinput.EditProcedureTypes}, GrantedBy: "dev-admin"},
			{ID: "dev-approver-2", UserID: "dev-approver", Area: authinput.SecretariatArea, Permissions: []string{authinput.PublishProcedureTypes}, GrantedBy: "dev-admin"},
		},
	}
}

// devAreas are the areas of staff until they are stored in PostgreSQL, besides
// the Secretariat, which EnsureSecretariat adds.
func devAreas() []domain.Area {
	return []domain.Area{
		{ID: "office", Name: "Área de prueba", Permissions: []domain.Permission{
			{ID: "review", Name: "Revisar"},
			{ID: "approve", Name: "Aprobar"},
		}},
		{ID: devCommitteeArea, Name: "Comisión de prueba", Permissions: []domain.Permission{{ID: "evaluate", Name: "Evaluar"}}},
	}
}

// publishIfMissing publishes an example only if its procedure type has no
// version. A database that has one keeps it even if the example changed
// since, so editing the example does not add a version on every start: a new
// flow goes in a new procedure type, as devReferralProcedureType. If another
// start publishes first, publishing again finds the same content and returns
// the stored version; another content is an error.
func publishIfMissing(ctx context.Context, versions output.ProcedureTypeVersionRepository, procedureTypes input.ProcedureTypes, publication input.Publication) error {
	_, err := versions.Latest(ctx, publication.ProcedureType)
	if !errors.Is(err, output.ErrVersionNotFound) {
		return err
	}
	_, err = procedureTypes.Publish(ctx, publication)
	if errors.Is(err, input.ErrOutdatedBase) {
		_, err = procedureTypes.Publish(ctx, publication)
	}
	return err
}

// addDevExample publishes the example procedure type and leaves a procedure of
// it for an applicant. With the identity of development the procedure is of
// the test applicant. With Google nobody signs in as that applicant, so it is
// of the first account that registers, which the returned users hand it to.
func addDevExample(ctx context.Context, cfg config, logger *slog.Logger, versions output.ProcedureTypeVersionRepository, procedureTypes input.ProcedureTypes, procedures output.ProcedureRepository, users authoutput.UserRepository) (authoutput.UserRepository, error) {
	if err := publishIfMissing(ctx, versions, procedureTypes, devProcedureType()); err != nil {
		return nil, fmt.Errorf("publish the example procedure type: %w", err)
	}
	if err := publishIfMissing(ctx, versions, procedureTypes, devReferralProcedureType()); err != nil {
		return nil, fmt.Errorf("publish the example procedure type with a referral: %w", err)
	}
	if err := publishIfMissing(ctx, versions, procedureTypes, devRulingProcedureType()); err != nil {
		return nil, fmt.Errorf("publish the example procedure type with a ruling: %w", err)
	}
	if cfg.identityProvider != identityDev {
		return &exampleForFirstSignUp{UserRepository: users, logger: logger, give: func(ctx context.Context, applicant string) error {
			return addDevProcedure(ctx, versions, procedures, applicant)
		}}, nil
	}
	if err := addDevProcedure(ctx, versions, procedures, devApplicantID); err != nil {
		return nil, fmt.Errorf("add the example procedure: %w", err)
	}
	return users, nil
}

// exampleForFirstSignUp gives the example procedure to the first account that
// registers. If giving it fails, the account still signs in, without it, and
// the example is left for the next account.
type exampleForFirstSignUp struct {
	authoutput.UserRepository
	give   func(ctx context.Context, applicant string) error
	logger *slog.Logger

	mu    sync.Mutex
	given bool
}

func (r *exampleForFirstSignUp) Register(ctx context.Context, user domain.User) (domain.User, error) {
	registered, err := r.UserRepository.Register(ctx, user)
	if err != nil || registered.ID != user.ID {
		return registered, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.given {
		return registered, nil
	}
	if err := r.give(ctx, registered.ID); err != nil {
		r.logger.Warn("could not give the example procedure", "user", registered.ID, "error", err)
		return registered, nil
	}
	r.given = true
	return registered, nil
}

// devProcedureID names the example procedure, so that its detail is always
// at the same address in local development. It waits on devDocumentsStep.
// devReviewProcedureID waits on devReviewStep, for staff to confirm.
const (
	devProcedureID       = "dev-example"
	devDocumentsStep     = "documents"
	devReviewProcedureID = "dev-review"
	devReviewStep        = "review"
)

// addDevProcedure leaves two procedures of the example type, one waiting on
// the documents of the applicant and one on the review of staff, so that
// there is something to show and to confirm until procedures can be started
// and moved from the interface.
func addDevProcedure(ctx context.Context, versions output.ProcedureTypeVersionRepository, procedures output.ProcedureRepository, applicant string) error {
	latest, err := versions.Latest(ctx, devProcedureType().ProcedureType)
	if err != nil {
		return err
	}
	now := time.Now()
	for id, step := range map[string]string{devProcedureID: devDocumentsStep, devReviewProcedureID: devReviewStep} {
		p := proceduredomain.Procedure{
			ID:            id,
			ProcedureType: latest.ProcedureTypeID,
			Version:       latest.Number,
			VersionDigest: latest.Digest,
			Applicant:     applicant,
			OpenStep:      step,
			StartedAt:     now,
			MovedAt:       now,
		}
		if err := procedures.Add(ctx, p, nil, devValues(p)...); err != nil {
			return err
		}
	}
	return nil
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
				{ID: "request", Name: "Solicitud", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Fields: []string{"reason"}, OptionalFields: []string{"comment"}, Next: devDocumentsStep},
				{ID: devDocumentsStep, Name: "Documentación", Role: definition.RoleApplicant, Action: definition.ActionUploadDocument, Documents: []string{"statement"}, Next: "review"},
				{ID: devReviewStep, Name: "Revisión", Role: definition.RoleStaff, Area: "office", Permissions: []string{"review"}, Action: definition.ActionConfirmExternalFact},
			},
		},
	}
}

const (
	devCommitteeArea = "committee"
	devEvaluateStep  = "evaluate"
)

// devReferralProcedureType is a second made-up procedure type whose review is
// a decision that refers it to the committee, and whose last step returns it.
func devReferralProcedureType() input.Publication {
	example := devProcedureType()
	example.ProcedureType = 2
	d := &example.Definition
	d.Name = "Trámite de ejemplo con derivación"
	d.Description = "Un trámite inventado para probar la derivación y el retorno en desarrollo."
	d.Steps = []definition.Step{
		d.Steps[0],
		d.Steps[1],
		{ID: devReviewStep, Name: "Revisión", Role: definition.RoleStaff, Area: "office", Permissions: []string{"review"}, Action: definition.ActionDecide, Outcomes: []definition.Outcome{
			{ID: "accept", Name: "Aceptar y derivar a la comisión", Next: devEvaluateStep},
			{ID: "send_back", Name: "Devolver al titular", Next: devDocumentsStep, SendBack: true},
		}},
		{ID: devEvaluateStep, Name: "Dictamen de la comisión", Role: definition.RoleStaff, Area: devCommitteeArea, Permissions: []string{"evaluate"}, Action: definition.ActionConfirmExternalFact, Next: "communicate"},
		{ID: "communicate", Name: "Comunicación del dictamen", Role: definition.RoleStaff, Area: "office", Permissions: []string{"review"}, Action: definition.ActionConfirmExternalFact},
	}
	return example
}

// devRulingProcedureType is the example with a referral whose last step asks
// the staff to attach the ruling.
func devRulingProcedureType() input.Publication {
	example := devReferralProcedureType()
	example.ProcedureType = 3
	d := &example.Definition
	d.Name = "Trámite de ejemplo con dictamen"
	d.Description = "Un trámite inventado para probar adjuntar el dictamen en desarrollo."
	d.Documents = append(d.Documents, definition.Document{ID: "ruling", Name: "Dictamen", AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: 10 << 20})
	last := &d.Steps[len(d.Steps)-1]
	last.Action, last.Documents = definition.ActionUploadDocument, []string{"ruling"}
	return example
}

// devValues are what the applicant of an example entered on the request step.
func devValues(p proceduredomain.Procedure) []proceduredomain.FieldValue {
	value := func(field, text string) proceduredomain.FieldValue {
		return proceduredomain.FieldValue{Procedure: p.ID, Field: field, Step: "request", Text: text, EnteredBy: p.Applicant, EnteredAt: p.StartedAt}
	}
	return []proceduredomain.FieldValue{
		value("reason", "Motivo de ejemplo: pido el trámite para probar SIGTA.\nSegunda línea del motivo."),
		value("comment", "Comentario de ejemplo."),
	}
}
