package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	"backend/src/migrations/dbtest"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/adapters/postgres"
	"backend/src/procedure/application"
	proceduredomain "backend/src/procedure/domain"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

// TestDevProcedureTypeIsPublishedOnce publishes an older example, as a
// previous binary would have, and then starts twice: the older one stays.
func TestDevProcedureTypeIsPublishedOnce(t *testing.T) {
	ctx := t.Context()
	schema := dbtest.NewSchema(t)
	app := schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted)
	older := devProcedureType()
	older.Definition.Name = "Trámite de ejemplo anterior"
	stored, err := application.NewProcedureTypeService(postgres.NewProcedureVersions(app), devAreaService(), devStaff(), time.Now).Publish(ctx, older)
	if err != nil {
		t.Fatal(err)
	}
	// Each start opens the same database with a new service.
	for range 2 {
		versions := postgres.NewProcedureVersions(app)
		service := application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now)
		if err := publishDevProcedureType(ctx, versions, service); err != nil {
			t.Fatal(err)
		}
	}
	latest, err := postgres.NewProcedureVersions(app).Latest(ctx, devProcedureType().ProcedureType)
	if err != nil || latest.Number != 1 || latest.Digest != stored.Digest {
		t.Fatalf("Latest() = version %d, %v; want the older example as version 1", latest.Number, err)
	}
}

// lostRace is a store where another start publishes rival right before the
// first Add.
type lostRace struct {
	output.ProcedureVersions
	rival input.Publication
	lost  bool
}

func (s *lostRace) Add(ctx context.Context, v version.ProcedureTypeVersion) error {
	if !s.lost {
		s.lost = true
		other := application.NewProcedureTypeService(s.ProcedureVersions, devAreaService(), devStaff(), time.Now)
		if _, err := other.Publish(ctx, s.rival); err != nil {
			return err
		}
	}
	return s.ProcedureVersions.Add(ctx, v)
}

func TestDevProcedureTypeWhenAnotherStartPublishesFirst(t *testing.T) {
	changed := devProcedureType()
	changed.Definition.Name = "Otro trámite de ejemplo"
	for _, tc := range []struct {
		name    string
		rival   input.Publication
		wantErr bool
	}{
		{"same content", devProcedureType(), false},
		{"another content", changed, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			store := &lostRace{ProcedureVersions: memory.NewProcedureVersions(), rival: tc.rival}
			service := application.NewProcedureTypeService(store, devAreaService(), devStaff(), time.Now)
			err := publishDevProcedureType(ctx, store, service)
			if !store.lost {
				t.Fatal("the other start never published")
			}
			if tc.wantErr != (err != nil) {
				t.Fatalf("publishDevProcedureType() = %v, want an error: %v", err, tc.wantErr)
			}
			if tc.wantErr && !errors.Is(err, input.ErrOutdatedBase) {
				t.Errorf("publishDevProcedureType() = %v, want ErrOutdatedBase", err)
			}
			latest, err := store.Latest(ctx, devProcedureType().ProcedureType)
			if err != nil || latest.Number != 1 {
				t.Fatalf("Latest() = version %d, %v; want version 1", latest.Number, err)
			}
		})
	}
}

// The example procedure opens as the server wires it: the applicant acts on
// its open step, and the staff of the area see it without acting.
func TestDevProcedureOpensForTheApplicantAndTheArea(t *testing.T) {
	ctx := t.Context()
	versions := memory.NewProcedureVersions()
	if err := publishDevProcedureType(ctx, versions, application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now)); err != nil {
		t.Fatal(err)
	}
	procedures := memory.NewProcedureRepository()
	if err := addDevProcedure(ctx, versions, procedures, devApplicantID); err != nil {
		t.Fatal(err)
	}
	directory := devPeople()
	areas := authmem.NewAreaRepository(devAreas()...)
	memberships := authapp.NewMembershipService(areas, authmem.NewAdministrators(directory.administrators...), authmem.NewMembershipRepository(directory.memberships...), func() string { return "id" }, time.Now)
	service := application.NewProcedureService(application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now), procedures, memberships, func() string { return "id" }, time.Now)

	for _, tc := range []struct {
		user   string
		canAct bool
	}{
		{devApplicantID, true},
		{"dev-operative", false},
	} {
		view, err := service.Open(ctx, tc.user, devProcedureID)
		if err != nil {
			t.Fatalf("%s opens the example: %v", tc.user, err)
		}
		if view.Procedure.OpenStep != devDocumentsStep || view.CanAct() != tc.canAct {
			t.Errorf("%s sees the open step %s and can act: %t, want documents and %t", tc.user, view.Procedure.OpenStep, view.CanAct(), tc.canAct)
		}
	}
}

func TestDevReviewProcedureIsConfirmedByTheArea(t *testing.T) {
	ctx := t.Context()
	versions := memory.NewProcedureVersions()
	if err := publishDevProcedureType(ctx, versions, application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now)); err != nil {
		t.Fatal(err)
	}
	procedures := memory.NewProcedureRepository()
	if err := addDevProcedure(ctx, versions, procedures, devApplicantID); err != nil {
		t.Fatal(err)
	}
	directory := devPeople()
	areas := authmem.NewAreaRepository(devAreas()...)
	memberships := authapp.NewMembershipService(areas, authmem.NewAdministrators(directory.administrators...), authmem.NewMembershipRepository(directory.memberships...), func() string { return "id" }, time.Now)
	service := application.NewProcedureService(application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now), procedures, memberships, func() string { return "id" }, time.Now)
	confirmation := input.StepConfirmation{Step: devReviewStep, Observation: "Revisado"}

	if _, err := service.Confirm(ctx, devApplicantID, devReviewProcedureID, confirmation); !errors.Is(err, proceduredomain.ErrAccessDenied) {
		t.Errorf("the applicant confirms the review: %v, want ErrAccessDenied", err)
	}
	if _, err := service.Confirm(ctx, "dev-operative", devReviewProcedureID, confirmation); err != nil {
		t.Fatalf("the operative confirms the review: %v", err)
	}
	if got, _ := procedures.Get(ctx, devReviewProcedureID); !got.Closed() {
		t.Errorf("open step = %q, want the example closed", got.OpenStep)
	}
}

// With Google nobody signs in as the test applicant, so the example goes to
// the first account that registers, and only to it.
func TestDevProcedureGoesToTheFirstSignUpWithGoogle(t *testing.T) {
	ctx := t.Context()
	versions := memory.NewProcedureVersions()
	procedureTypes := application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now)
	procedures := memory.NewProcedureRepository()
	users, err := addDevExample(ctx, config{identityProvider: identityGoogle}, slog.New(slog.DiscardHandler), versions, procedureTypes, procedures, authmem.NewUserRepository(devPeople().users...))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := procedures.Get(ctx, devProcedureID); err == nil {
		t.Fatal("the example exists before anyone registers")
	}
	for _, user := range []authdomain.User{
		{ID: "first", Email: "primera@fi.uba.ar"},
		{ID: "again", Email: "primera@fi.uba.ar"},
		{ID: "second", Email: "segunda@fi.uba.ar"},
	} {
		if _, err := users.Register(ctx, user); err != nil {
			t.Fatalf("register %s: %v", user.ID, err)
		}
	}
	got, err := procedures.Get(ctx, devProcedureID)
	if err != nil || got.Applicant != "first" {
		t.Errorf("the example is of %q, %v; want the first account", got.Applicant, err)
	}
}

func TestFailingToGiveTheExampleDoesNotStopTheSignUp(t *testing.T) {
	ctx := t.Context()
	var given []string
	users := &exampleForFirstSignUp{
		UserRepository: authmem.NewUserRepository(),
		logger:         slog.New(slog.DiscardHandler),
		give: func(_ context.Context, applicant string) error {
			if applicant == "first" {
				return errors.New("the example type is not published")
			}
			given = append(given, applicant)
			return nil
		},
	}
	for _, user := range []authdomain.User{
		{ID: "first", Email: "primera@fi.uba.ar"},
		{ID: "second", Email: "segunda@fi.uba.ar"},
		{ID: "third", Email: "tercera@fi.uba.ar"},
	} {
		if registered, err := users.Register(ctx, user); err != nil || registered.ID != user.ID {
			t.Fatalf("register %s: %+v, %v", user.ID, registered, err)
		}
	}
	if len(given) != 1 || given[0] != "second" {
		t.Errorf("the example went to %v, want only the second account", given)
	}
}

// devAreaService reads the areas of the server, so that the example is
// published against the same areas as on a start.
func devAreaService() *authapp.AreaService {
	areas, administrators, memberships := devDirectory()
	return authapp.NewAreaService(areas, administrators, memberships, time.Now)
}

// devStaff holds the memberships of the synthetic people, so that the example
// is published by the same proposer and approver as on a start.
func devStaff() *authapp.MembershipService {
	areas, administrators, memberships := devDirectory()
	return authapp.NewMembershipService(areas, administrators, memberships, newID, time.Now)
}

func devDirectory() (*authmem.AreaRepository, *authmem.Administrators, *authmem.MembershipRepository) {
	areas := authmem.NewAreaRepository(devAreas()...)
	if err := authapp.EnsureSecretariat(context.Background(), areas); err != nil {
		panic(err)
	}
	people := devPeople()
	return areas, authmem.NewAdministrators(people.administrators...), authmem.NewMembershipRepository(people.memberships...)
}

// TestServeLetsTheRequestsInFlightFinish stops the server while a request is
// still running: the request gets its answer before serve returns.
func TestServeLetsTheRequestsInFlightFinish(t *testing.T) {
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(started)
			<-release
			w.WriteHeader(http.StatusNoContent)
		}),
		ReadHeaderTimeout: time.Second,
	}
	stop, cancel := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- serve(stop, server, func() error { return server.Serve(listener) }) }()

	answered := make(chan int, 1)
	go func() {
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+listener.Addr().String(), nil)
		if err != nil {
			t.Error(err)
			answered <- 0
			return
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Error(err)
			answered <- 0
			return
		}
		_ = response.Body.Close()
		answered <- response.StatusCode
	}()
	<-started
	cancel()
	select {
	case err := <-served:
		t.Fatalf("serve() = %v before the request in flight finished", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	if status := <-answered; status != http.StatusNoContent {
		t.Errorf("the request in flight got %d, want %d", status, http.StatusNoContent)
	}
	if err := <-served; err != nil {
		t.Errorf("serve() = %v, want nil", err)
	}
}

// TestStoppedDuringStartup checks that a start interrupted by a signal ends
// without an error, and that one that fails on its own keeps it.
func TestStoppedDuringStartup(t *testing.T) {
	failed := errors.New("the database did not answer")
	if err := stoppedDuringStartup(t.Context(), failed); !errors.Is(err, failed) {
		t.Errorf("without a signal = %v, want %v", err, failed)
	}
	stop, cancel := context.WithCancel(t.Context())
	cancel()
	if err := stoppedDuringStartup(stop, failed); err != nil {
		t.Errorf("after a signal = %v, want nil", err)
	}
}
