package acceptance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/draft"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

// draftWorld is the state of one scenario of US72. Every scenario works on a
// single draft, of procedure type pt.
type draftWorld struct {
	areas       *authmem.AreaRepository
	memberships *authmem.MembershipRepository
	versions    output.ProcedureTypeVersionRepository
	drafts      input.ProcedureTypeDrafts
	types       *application.ProcedureTypeService

	pt       int64
	view     input.DraftView
	first    version.ProcedureTypeVersion
	approved version.ProcedureTypeVersion
	err      error
	// before is the latest version of pt when the draft was approved.
	before int
}

func registerProcedureTypeDraftSteps(sc *godog.ScenarioContext) {
	w := &draftWorld{versions: memory.NewProcedureTypeVersionRepository()}

	sc.Given(`^the staff areas for the drafts are "([^"]*)" and "([^"]*)"$`, w.givenAreas)
	sc.Given(`^"([^"]*)" can propose procedure types, "([^"]*)" can approve them and "([^"]*)" can do both$`, w.givenPeople)
	sc.Given(`^"([^"]*)" created the draft of a new procedure type "([^"]*)"$`, w.givenCreated)
	sc.Given(`^"([^"]*)" created the draft of a new procedure type with a step in an unknown area$`, w.givenCreatedInvalid)
	sc.Given(`^"([^"]*)" proposed the draft of a new procedure type "([^"]*)"$`, w.givenProposedNew)
	sc.Given(`^version 1 of "([^"]*)" is published without a draft$`, w.givenFirstVersion)
	sc.Given(`^"([^"]*)" proposed a draft of "([^"]*)" with another document name$`, w.givenProposedRevision)
	sc.Given(`^version 2 of "([^"]*)" is published by other means$`, w.givenOtherVersion)

	sc.When(`^"([^"]*)" proposes the draft$`, w.propose)
	sc.When(`^"([^"]*)" approves the draft$`, w.approve)
	sc.When(`^"([^"]*)" rejects the draft because "([^"]*)"$`, w.reject)
	sc.When(`^"([^"]*)" opens a draft of "([^"]*)"$`, w.open)
	sc.When(`^"([^"]*)" edits the draft$`, w.edit)

	sc.Then(`^version 1 of the new procedure type is published, proposed by "([^"]*)" and approved by "([^"]*)"$`, w.thenNewTypePublished)
	sc.Then(`^the procedure type has no draft left$`, w.thenNoDraft)
	sc.Then(`^the draft has the definition of version 1$`, w.thenFromFirstVersion)
	sc.Then(`^version 2 of "([^"]*)" is published from the draft$`, w.thenSecondVersion)
	sc.Then(`^version 1 keeps its content$`, w.thenFirstUnchanged)
	sc.Then(`^the change is refused because the draft is proposed$`, w.thenRefusedProposed)
	sc.Then(`^the draft can be edited again$`, w.thenEditable)
	sc.Then(`^the draft shows that "([^"]*)" rejected it because "([^"]*)"$`, w.thenShowsRejection)
	sc.Then(`^the draft is not proposed$`, w.thenNotProposed)
	sc.Then(`^the problem points at the step in the unknown area$`, w.thenProblemAtStep)
	sc.Then(`^no version is published from the draft$`, w.thenNoVersionFromDraft)
	sc.Then(`^the approval is refused because another version was published since$`, w.thenOutdated)
}

// givenAreas adds the Secretariat, whose permissions the drafts ask for
// (ADR-020), to both areas.
func (w *draftWorld) givenAreas(office, other string) error {
	w.areas = authmem.NewAreaRepository(
		authdomain.Area{ID: areaID(office), Name: office, Permissions: []authdomain.Permission{{ID: "review", Name: "Revisar"}, {ID: "sign", Name: "Firmar"}}},
		authdomain.Area{ID: areaID(other), Name: other, Permissions: []authdomain.Permission{{ID: "evaluate", Name: "Evaluar"}}},
	)
	return authapp.EnsureSecretariat(context.Background(), w.areas)
}

func (w *draftWorld) givenPeople(proposer, approver, both string) error {
	edit, publish := authinput.EditProcedureTypes, authinput.PublishProcedureTypes
	w.memberships = authmem.NewMembershipRepository(
		authdomain.Membership{ID: "m1", UserID: proposer, Area: authinput.SecretariatArea, Permissions: []string{edit}},
		authdomain.Membership{ID: "m2", UserID: approver, Area: authinput.SecretariatArea, Permissions: []string{publish}},
		authdomain.Membership{ID: "m3", UserID: both, Area: authinput.SecretariatArea, Permissions: []string{edit, publish}},
		// The authors of the versions published without a draft.
		authdomain.Membership{ID: "m4", UserID: "carla", Area: authinput.SecretariatArea, Permissions: []string{edit}},
		authdomain.Membership{ID: "m5", UserID: "dario", Area: authinput.SecretariatArea, Permissions: []string{publish}},
	)
	now := func() time.Time { return time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC) }
	administrators := authmem.NewAdministratorRepository()
	areas := authapp.NewAreaService(w.areas, administrators, w.memberships, now)
	staff := authapp.NewMembershipService(w.areas, administrators, w.memberships, func() string { return "id" }, now)
	w.types = application.NewProcedureTypeService(w.versions, areas, staff, now)
	w.drafts = application.NewProcedureTypeDraftService(memory.NewProcedureTypeDraftRepository(), w.types, staff, now)
	return nil
}

func (w *draftWorld) create(person string, d definition.ProcedureTypeDefinition) error {
	view, err := w.drafts.Create(context.Background(), person, d)
	if err != nil {
		return err
	}
	w.pt, w.view = view.Draft.ProcedureTypeID, view
	return nil
}

func (w *draftWorld) givenCreated(person, name string) error {
	return w.create(person, validDefinition(name))
}

func (w *draftWorld) givenCreatedInvalid(person string) error {
	d := validDefinition("subject equivalence")
	d.Steps[2].Area = "rectorate"
	return w.create(person, d)
}

func (w *draftWorld) givenProposedNew(person, name string) error {
	if err := w.givenCreated(person, name); err != nil {
		return err
	}
	return w.mustPropose(person)
}

func (w *draftWorld) mustPropose(person string) error {
	view, err := w.drafts.Propose(context.Background(), person, w.pt, w.view.Draft.Revision)
	if err != nil {
		return err
	}
	w.view = view
	return nil
}

func (w *draftWorld) publishDirectly(base int, d definition.ProcedureTypeDefinition) (version.ProcedureTypeVersion, error) {
	return w.types.Publish(context.Background(), input.Publication{
		ProcedureType: procedureType,
		Base:          base,
		Definition:    d,
		Authorship:    version.Authorship{ProposedBy: "carla", ApprovedBy: "dario"},
	})
}

func (w *draftWorld) givenFirstVersion(name string) error {
	v, err := w.publishDirectly(0, validDefinition(name))
	w.first = v
	return err
}

func (w *draftWorld) givenProposedRevision(person, name string) error {
	if err := w.open(person, name); err != nil {
		return err
	}
	if w.err != nil {
		return w.err
	}
	d := w.view.Draft.Definition
	d.Documents = append([]definition.Document(nil), d.Documents...)
	d.Documents[0].Name = "Certificado analítico legalizado"
	view, err := w.drafts.Edit(context.Background(), person, w.pt, w.view.Draft.Revision, d)
	if err != nil {
		return err
	}
	w.view = view
	return w.mustPropose(person)
}

func (w *draftWorld) givenOtherVersion(name string) error {
	d := validDefinition(name)
	d.Description = "Publicada por otro lado"
	_, err := w.publishDirectly(1, d)
	return err
}

func (w *draftWorld) propose(person string) error {
	view, err := w.drafts.Propose(context.Background(), person, w.pt, w.view.Draft.Revision)
	w.err = err
	if err == nil {
		w.view = view
	}
	return nil
}

func (w *draftWorld) approve(person string) error {
	before, err := w.latest()
	if err != nil {
		return err
	}
	w.before = before
	w.approved, w.err = w.drafts.Approve(context.Background(), person, w.pt, w.view.Draft.Revision)
	return nil
}

func (w *draftWorld) latest() (int, error) {
	v, err := w.versions.Latest(context.Background(), w.pt)
	if errors.Is(err, output.ErrVersionNotFound) {
		return 0, nil
	}
	return v.Number, err
}

func (w *draftWorld) reject(person, reason string) error {
	view, err := w.drafts.Reject(context.Background(), person, w.pt, w.view.Draft.Revision, reason)
	w.err = err
	if err == nil {
		w.view = view
	}
	return nil
}

func (w *draftWorld) open(person, _ string) error {
	view, err := w.drafts.Open(context.Background(), person, procedureType)
	w.err = err
	if err == nil {
		w.pt, w.view = procedureType, view
	}
	return nil
}

func (w *draftWorld) edit(person string) error {
	d := w.view.Draft.Definition
	d.Description = "Otra descripción"
	view, err := w.drafts.Edit(context.Background(), person, w.pt, w.view.Draft.Revision, d)
	w.err = err
	if err == nil {
		w.view = view
	}
	return nil
}

func (w *draftWorld) thenNewTypePublished(proposer, approver string) error {
	if w.err != nil {
		return w.err
	}
	v := w.approved
	if v.ProcedureTypeID != w.pt || v.Number != 1 || v.Authorship.ProposedBy != proposer || v.Authorship.ApprovedBy != approver || v.PublishedAt.IsZero() {
		return fmt.Errorf("published %+v, want version 1 of procedure type %d by %s and %s", v, w.pt, proposer, approver)
	}
	return nil
}

func (w *draftWorld) thenNoDraft() error {
	_, err := w.drafts.Get(context.Background(), "ana", w.pt)
	if !errors.Is(err, output.ErrDraftNotFound) {
		return fmt.Errorf("the draft is still there: %w", err)
	}
	return nil
}

func (w *draftWorld) thenFromFirstVersion() error {
	if w.err != nil {
		return w.err
	}
	content, err := w.first.DecodeContent()
	if err != nil {
		return err
	}
	d := w.view.Draft
	if d.Base != 1 || d.Status != draft.StatusEditing || d.Definition.Name != content.Definition.Name || len(d.Definition.Steps) != len(content.Definition.Steps) {
		return fmt.Errorf("draft %+v, want the definition of version 1", d)
	}
	return nil
}

func (w *draftWorld) thenSecondVersion(string) error {
	if w.err != nil {
		return w.err
	}
	if w.approved.ProcedureTypeID != procedureType || w.approved.Number != 2 {
		return fmt.Errorf("published %+v, want version 2", w.approved)
	}
	return nil
}

func (w *draftWorld) thenFirstUnchanged() error {
	v, err := w.versions.Get(context.Background(), procedureType, 1)
	if err != nil {
		return err
	}
	if v.Digest != w.first.Digest || string(v.EncodedContent) != string(w.first.EncodedContent) {
		return errors.New("version 1 changed")
	}
	return nil
}

func (w *draftWorld) thenRefusedProposed() error {
	if !errors.Is(w.err, draft.ErrDraftProposed) {
		return fmt.Errorf("edit = %w, want ErrDraftProposed", w.err)
	}
	return nil
}

func (w *draftWorld) thenEditable() error {
	if w.err != nil {
		return w.err
	}
	if w.view.Draft.Status != draft.StatusEditing {
		return fmt.Errorf("status %s, want editing", w.view.Draft.Status)
	}
	if err := w.edit("ana"); err != nil {
		return err
	}
	return w.err
}

func (w *draftWorld) thenShowsRejection(person, reason string) error {
	view, err := w.drafts.Get(context.Background(), "ana", w.pt)
	if err != nil {
		return err
	}
	r := view.Draft.Rejection
	if r.By != person || r.Reason != reason || r.At.IsZero() {
		return fmt.Errorf("rejection %+v, want by %s because %q", r, person, reason)
	}
	return nil
}

func (w *draftWorld) thenNotProposed() error {
	if w.err == nil {
		return errors.New("the draft was proposed")
	}
	view, err := w.drafts.Get(context.Background(), "ana", w.pt)
	if err != nil {
		return err
	}
	if view.Draft.Status != draft.StatusEditing {
		return fmt.Errorf("status %s, want editing", view.Draft.Status)
	}
	return nil
}

func (w *draftWorld) thenProblemAtStep() error {
	invalid, ok := errors.AsType[*definition.ValidationError](w.err)
	if !ok {
		return fmt.Errorf("propose = %w, want a ValidationError", w.err)
	}
	for _, p := range invalid.Problems {
		if p.Code == definition.CodeUnknownArea && len(p.Path) > 0 && p.Path[len(p.Path)-1].ID == "review" && p.Value == "rectorate" {
			return nil
		}
	}
	return fmt.Errorf("problems %v do not point at the review step", invalid.Problems)
}

func (w *draftWorld) thenNoVersionFromDraft() error {
	if w.err == nil {
		return fmt.Errorf("published %+v", w.approved)
	}
	latest, err := w.latest()
	if err != nil {
		return err
	}
	if latest != w.before {
		return fmt.Errorf("latest version is %d, it was %d", latest, w.before)
	}
	return nil
}

func (w *draftWorld) thenOutdated() error {
	if !errors.Is(w.err, input.ErrOutdatedBase) {
		return fmt.Errorf("approve = %w, want ErrOutdatedBase", w.err)
	}
	return nil
}
