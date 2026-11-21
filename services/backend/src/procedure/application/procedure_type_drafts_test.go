package application

import (
	"errors"
	"testing"
	"time"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/draft"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

// newDraftService has ana, who proposes, beto, who approves, carla, who
// proposes only for procedure type 1, dani, who is in no area, and eva, who
// proposes and approves, all in the Secretariat.
func newDraftService(t *testing.T) *ProcedureTypeDraftService {
	t.Helper()
	areas := authmem.NewAreaRepository(office)
	if err := authapp.EnsureSecretariat(t.Context(), areas); err != nil {
		t.Fatal(err)
	}
	edit, publish := authinput.EditProcedureTypes, authinput.PublishProcedureTypes
	memberships := authmem.NewMembershipRepository(
		authdomain.Membership{ID: "m1", UserID: "ana", Area: authinput.SecretariatArea, Permissions: []string{edit}},
		authdomain.Membership{ID: "m2", UserID: "beto", Area: authinput.SecretariatArea, Permissions: []string{publish}},
		authdomain.Membership{ID: "m3", UserID: "carla", Area: authinput.SecretariatArea, Permissions: []string{edit}, Scope: []int64{1}},
		authdomain.Membership{ID: "m4", UserID: "eva", Area: authinput.SecretariatArea, Permissions: []string{edit, publish}},
	)
	administrators := authmem.NewAdministratorRepository()
	staff := authapp.NewMembershipService(areas, administrators, memberships, authmem.NewUserRepository(), func() string { return "id" }, time.Now)
	types := NewProcedureTypeService(memory.NewProcedureTypeVersionRepository(), authapp.NewAreaService(areas, administrators, memberships, time.Now), staff, func() time.Time { return publishedAt })
	return NewProcedureTypeDraftService(memory.NewProcedureTypeDraftRepository(), types, staff, func() time.Time { return publishedAt })
}

func TestDraftOfANewProcedureTypeIsPublishedWhenApproved(t *testing.T) {
	ctx := t.Context()
	s := newDraftService(t)
	publish(t, s, 1, testDefinition())

	created, err := s.Create(ctx, "ana", definition.ProcedureTypeDefinition{Name: "Nuevo"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Draft.ProcedureTypeID != 2 || created.Draft.Base != 0 || len(created.Problems) == 0 {
		t.Errorf("Create() = %+v, want procedure type 2 from no version, with problems", created)
	}
	if _, err := s.Propose(ctx, "ana", 2, 1); !errors.Is(err, definition.ErrInvalidDefinition) {
		t.Errorf("Propose() of an invalid draft = %v, want ErrInvalidDefinition", err)
	}

	edited, err := s.Edit(ctx, "ana", 2, 1, testDefinition())
	if err != nil {
		t.Fatal(err)
	}
	if len(edited.Problems) != 0 {
		t.Errorf("Edit() problems = %v, want none", edited.Problems)
	}
	if _, err := s.Propose(ctx, "ana", 2, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Approve(ctx, "ana", 2, 3); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("Approve() by whoever proposes = %v, want ErrAccessDenied", err)
	}
	v, err := s.Approve(ctx, "beto", 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if v.ProcedureTypeID != 2 || v.Number != 1 || v.Authorship.ProposedBy != "ana" || v.Authorship.ApprovedBy != "beto" {
		t.Errorf("Approve() = %+v", v)
	}
	if _, err := s.Get(ctx, "ana", 2); !errors.Is(err, output.ErrDraftNotFound) {
		t.Errorf("Get() after Approve() = %v, want ErrDraftNotFound", err)
	}
}

func TestNewVersionStartsFromTheLatestAndChecksItsBase(t *testing.T) {
	ctx := t.Context()
	s := newDraftService(t)
	publish(t, s, 1, testDefinition())
	opened, err := s.Open(ctx, "ana", 1)
	if err != nil {
		t.Fatal(err)
	}
	if opened.Draft.Base != 1 || opened.Draft.Definition.Name != testDefinition().Name {
		t.Errorf("Open() = %+v, want the definition of version 1", opened.Draft)
	}
	if _, err := s.Open(ctx, "ana", 1); !errors.Is(err, output.ErrDraftExists) {
		t.Errorf("Open() again = %v, want ErrDraftExists", err)
	}
	if _, err := s.Open(ctx, "ana", 9); !errors.Is(err, output.ErrVersionNotFound) {
		t.Errorf("Open() of an unknown procedure type = %v, want ErrVersionNotFound", err)
	}

	changed := testDefinition()
	changed.Description = "Otra descripción"
	if _, err := s.Edit(ctx, "ana", 1, 1, changed); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Propose(ctx, "ana", 1, 2); err != nil {
		t.Fatal(err)
	}
	// Someone publishes version 2 before the draft is approved.
	other := testDefinition()
	other.Description = "De otro lado"
	if _, err := s.types.Publish(ctx, publication(1, 1, other)); err != nil {
		t.Fatal(err)
	}
	var outdated *input.OutdatedBaseError
	if _, err := s.Approve(ctx, "beto", 1, 3); !errors.As(err, &outdated) || outdated.Latest != 2 {
		t.Errorf("Approve() on an outdated base = %v, want OutdatedBaseError with latest 2", err)
	}
}

func TestRejectionGoesBackToEditing(t *testing.T) {
	ctx := t.Context()
	s := newDraftService(t)
	if _, err := s.Create(ctx, "ana", testDefinition()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Propose(ctx, "ana", 1, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Edit(ctx, "ana", 1, 2, testDefinition()); !errors.Is(err, draft.ErrDraftProposed) {
		t.Errorf("Edit() while proposed = %v, want ErrDraftProposed", err)
	}
	if err := s.Discard(ctx, "ana", 1, 2); !errors.Is(err, draft.ErrDraftProposed) {
		t.Errorf("Discard() while proposed = %v, want ErrDraftProposed", err)
	}
	rejected, err := s.Reject(ctx, "beto", 1, 2, "Falta el plazo")
	if err != nil {
		t.Fatal(err)
	}
	if rejected.Draft.Status != draft.StatusEditing || rejected.Draft.Rejection.Reason != "Falta el plazo" {
		t.Errorf("Reject() = %+v", rejected.Draft)
	}
	if _, err := s.Edit(ctx, "ana", 1, 2, testDefinition()); !errors.Is(err, output.ErrDraftChanged) {
		t.Errorf("Edit() on an old revision = %v, want ErrDraftChanged", err)
	}
	if err := s.Discard(ctx, "ana", 1, 3); err != nil {
		t.Errorf("Discard() = %v", err)
	}
}

func TestDraftPermissions(t *testing.T) {
	ctx := t.Context()
	s := newDraftService(t)
	if _, err := s.Create(ctx, "carla", testDefinition()); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("Create() by a scoped member = %v, want ErrAccessDenied", err)
	}
	if _, err := s.Create(ctx, "ana", testDefinition()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(ctx, "ana", testDefinition()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Edit(ctx, "carla", 1, 1, testDefinition()); err != nil {
		t.Errorf("Edit() of a procedure type in scope = %v", err)
	}
	if _, err := s.Edit(ctx, "carla", 2, 1, testDefinition()); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("Edit() out of scope = %v, want ErrAccessDenied", err)
	}
	if _, err := s.Propose(ctx, "beto", 2, 1); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("Propose() by whoever approves = %v, want ErrAccessDenied", err)
	}
	if _, err := s.Get(ctx, "dani", 1); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("Get() by someone in no area = %v, want ErrAccessDenied", err)
	}
	if _, err := s.Propose(ctx, "eva", 2, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Approve(ctx, "eva", 2, 2); !errors.Is(err, draft.ErrSelfApproval) {
		t.Errorf("Approve() of one's own proposal = %v, want ErrSelfApproval", err)
	}
	for user, want := range map[string]int{"ana": 2, "beto": 2, "carla": 1, "dani": 0} {
		views, err := s.List(ctx, user)
		if err != nil || len(views) != want {
			t.Errorf("List(%s) = %d drafts, %v, want %d", user, len(views), err, want)
		}
	}
}

func publish(t *testing.T, s *ProcedureTypeDraftService, procedureType int64, d definition.ProcedureTypeDefinition) {
	t.Helper()
	if _, err := s.types.Publish(t.Context(), publication(procedureType, 0, d)); err != nil {
		t.Fatal(err)
	}
}
