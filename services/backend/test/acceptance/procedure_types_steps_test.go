package acceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cucumber/godog"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

const procedureType int64 = 1

// procedureTypeWorld is the state of one scenario of US08.
type procedureTypeWorld struct {
	// areas is what auth keeps, and stepAreas what a version copies of them.
	areas     *authmem.AreaRepository
	stepAreas []domain.Area
	versions  *memory.ProcedureVersions
	authors   domain.Authorship
	now       time.Time

	definition domain.Definition
	revisions  int

	// The result of the last publication.
	published domain.Version
	err       error
	before    int
	problem   domain.Code

	followed     domain.Version
	previousName string
	requested    domain.Version
}

func registerProcedureTypeSteps(sc *godog.ScenarioContext) {
	w := &procedureTypeWorld{
		versions: memory.NewProcedureVersions(),
		now:      time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC),
	}

	sc.Given(`^the staff areas are "([^"]*)" and "([^"]*)"$`, w.givenAreas)
	sc.Given(`^"([^"]*)" proposes and "([^"]*)" approves every publication$`, w.givenAuthors)
	sc.Given(`^version (\d+) of "([^"]*)" is published$`, w.givenPublished)
	sc.Given(`^version (\d+) of "([^"]*)" was proposed by "([^"]*)" and approved by "([^"]*)"$`, w.givenPublishedBy)
	sc.Given(`^a procedure that started with version (\d+) of "([^"]*)"$`, w.givenProcedure)
	sc.Given(`^the stored content of version (\d+) no longer matches its digest$`, w.givenAlteredVersion)

	sc.When(`^a valid "([^"]*)" definition is published from version (\d+)$`, w.publishValid)
	sc.When(`^the same definition is published from version (\d+)$`, w.publishSame)
	sc.When(`^"([^"]*)" proposes and "([^"]*)" approves the same definition$`, w.publishSameBy)
	sc.When(`^the name of a document is corrected and the definition is published from version (\d+)$`, w.publishRevision)
	sc.When(`^a different definition is published from version (\d+)$`, w.publishRevision)
	sc.When(`^"([^"]*)" proposes and "([^"]*)" approves a "([^"]*)" definition$`, w.publishBy)
	sc.When(`^a "([^"]*)" definition with (.+) is published$`, w.publishInvalid)
	sc.When(`^version (\d+) is published$`, w.publishUpTo)
	sc.When(`^the "([^"]*)" area is renamed$`, w.renameArea)
	sc.When(`^a procedure asks for version (\d+)$`, w.askForVersion)

	sc.Then(`^version (\d+) of "([^"]*)" is published$`, w.thenPublished)
	sc.Then(`^it records who proposed it, who approved it and when it was published$`, w.thenRecordsAuthorship)
	sc.Then(`^the latest version is still (\d+)$`, w.thenLatestIs)
	sc.Then(`^it still records "([^"]*)" and "([^"]*)"$`, w.thenStillRecords)
	sc.Then(`^no version is published$`, w.thenNoVersion)
	sc.Then(`^whoever published it is asked to start again from version (\d+)$`, w.thenAskedToStartAgain)
	sc.Then(`^the problem is reported$`, w.thenProblemReported)
	sc.Then(`^the procedure still follows version (\d+)$`, w.thenProcedureFollows)
	sc.Then(`^new procedures follow version (\d+)$`, w.thenLatestIs)
	sc.Then(`^version (\d+) still shows the previous name of the area$`, w.thenShowsPreviousName)
	sc.Then(`^the version is not returned$`, w.thenVersionNotReturned)
	sc.Then(`^nothing is published on top of it$`, w.thenNothingOnTop)
}

func (w *procedureTypeWorld) service() *application.ProcedureTypeService {
	auth := authapp.NewAreaService(w.areas, authmem.NewAdministrators(), authmem.NewMembershipRepository(), authmem.NewDeniedAttemptLog(), func() time.Time { return w.now })
	return application.NewProcedureTypeService(w.versions, auth, func() time.Time { return w.now })
}

func (w *procedureTypeWorld) latest() (int, error) {
	v, err := w.versions.Latest(context.Background(), procedureType)
	if errors.Is(err, output.ErrNoVersion) {
		return 0, nil
	}
	return v.Number, err
}

func (w *procedureTypeWorld) publish(base int, d domain.Definition) error {
	before, err := w.latest()
	if err != nil {
		return err
	}
	w.before = before
	w.definition = d
	w.published, w.err = w.service().Publish(context.Background(), input.Publication{
		ProcedureType: procedureType,
		Base:          base,
		Definition:    d,
		Authorship:    w.authors,
	})
	return nil
}

// validDefinition uses the four actions and every kind of area.
func validDefinition(name string) domain.Definition {
	return domain.Definition{
		Name:        name,
		Description: "Un trámite de prueba",
		Fields:      []domain.Field{{ID: "full_name", Name: "Nombre completo", Type: domain.FieldText}},
		Documents: []domain.Document{
			{ID: "transcript", Name: "Certificado analítico", AcceptedTypes: []string{"application/pdf"}, MaxSize: 10 << 20},
			{ID: "ruling", Name: "Dictamen", AcceptedTypes: []string{"application/pdf"}, MaxSize: 10 << 20},
		},
		Steps: []domain.Step{
			{ID: "request", Name: "Solicitud", Role: domain.RoleApplicant, Action: domain.ActionEnterData, Fields: []string{"full_name"}, Next: "documents"},
			{ID: "documents", Name: "Documentación", Role: domain.RoleApplicant, Action: domain.ActionUploadDocument, Documents: []string{"transcript"}, Next: "review"},
			{
				ID: "review", Name: "Revisión", Role: domain.RoleStaff, Area: "student_affairs", Profile: domain.ProfileStaffOperative, Action: domain.ActionDecide,
				Outcomes: []domain.Outcome{
					{ID: "complete", Name: "Completa", Next: "evaluation"},
					{ID: "incomplete", Name: "Incompleta", Next: "documents"},
				},
			},
			{ID: "evaluation", Name: "Evaluación", Role: domain.RoleStaff, Area: "curricular_commission", Profile: domain.ProfileStaffOperative, Action: domain.ActionUploadDocument, Documents: []string{"ruling"}, Next: "signature"},
			{ID: "signature", Name: "Firma", Role: domain.RoleStaff, Area: "student_affairs", Profile: domain.ProfileStaffApprover, Action: domain.ActionConfirmExternalFact},
		},
	}
}

// revised returns the current definition with a visible change.
func (w *procedureTypeWorld) revised() domain.Definition {
	w.revisions++
	d := w.definition
	d.Documents = slices.Clone(d.Documents)
	d.Documents[0].Name = fmt.Sprintf("Certificado analítico %d", w.revisions)
	return d
}

// publishUpTo publishes revisions until number is the latest version.
func (w *procedureTypeWorld) publishUpTo(number int) error {
	for {
		latest, err := w.latest()
		if err != nil || latest >= number {
			return err
		}
		if err := w.publish(latest, w.revised()); err != nil {
			return err
		}
		if w.err != nil {
			return w.err
		}
	}
}

// givenAreas stores the staff areas in auth. The applicant has no area.
func (w *procedureTypeWorld) givenAreas(office, other string) error {
	w.areas = authmem.NewAreaRepository(
		authdomain.Area{ID: areaID(office), Name: office},
		authdomain.Area{ID: areaID(other), Name: other},
	)
	w.stepAreas = []domain.Area{
		{ID: areaID(office), Name: office},
		{ID: areaID(other), Name: other},
	}
	return nil
}

func areaID(name string) string { return strings.ReplaceAll(name, " ", "_") }

func (w *procedureTypeWorld) givenAuthors(proposer, approver string) error {
	w.authors = domain.Authorship{ProposedBy: proposer, ApprovedBy: approver}
	return nil
}

func (w *procedureTypeWorld) givenPublished(number int, name string) error {
	w.definition = validDefinition(name)
	return w.publishUpTo(number)
}

func (w *procedureTypeWorld) givenPublishedBy(number int, name, proposer, approver string) error {
	if err := w.givenAuthors(proposer, approver); err != nil {
		return err
	}
	return w.givenPublished(number, name)
}

func (w *procedureTypeWorld) givenProcedure(number int, name string) error {
	if err := w.givenPublished(number, name); err != nil {
		return err
	}
	var err error
	w.followed, err = w.service().Version(context.Background(), procedureType, number)
	return err
}

// givenAlteredVersion stores a version whose content is not the one its digest
// covers, as if someone had changed it directly in the store.
func (w *procedureTypeWorld) givenAlteredVersion(number int) error {
	w.definition = validDefinition("subject equivalence")
	encoded, err := domain.NewContent(w.definition, w.stepAreas).Encode()
	if err != nil {
		return err
	}
	return w.versions.Add(context.Background(), domain.Version{
		ProcedureType: procedureType,
		Number:        number,
		Digest:        domain.Digest(encoded),
		Encoded:       append(encoded, ' '),
		Authorship:    w.authors,
		PublishedAt:   w.now,
	})
}

func (w *procedureTypeWorld) publishValid(name string, base int) error {
	return w.publish(base, validDefinition(name))
}

func (w *procedureTypeWorld) publishSame(base int) error {
	return w.publish(base, w.definition)
}

func (w *procedureTypeWorld) publishSameBy(proposer, approver string) error {
	if err := w.givenAuthors(proposer, approver); err != nil {
		return err
	}
	latest, err := w.latest()
	if err != nil {
		return err
	}
	return w.publish(latest, w.definition)
}

func (w *procedureTypeWorld) publishRevision(base int) error {
	return w.publish(base, w.revised())
}

func (w *procedureTypeWorld) publishBy(proposer, approver, name string) error {
	if err := w.givenAuthors(proposer, approver); err != nil {
		return err
	}
	return w.publish(0, validDefinition(name))
}

func (w *procedureTypeWorld) publishInvalid(name, problem string) error {
	d := validDefinition(name)
	const (
		document  = 1
		decision  = 2
		lastStep  = 4
		extraName = "Sin usar"
	)
	switch problem {
	case "a step with an action outside the four":
		d.Steps[0].Action = "approve"
		w.problem = domain.CodeUnknownAction
	case "a step in an unknown area":
		d.Steps[decision].Area = "nowhere"
		w.problem = domain.CodeUnknownArea
	case "a reference to an undeclared document":
		d.Steps[document].Documents = []string{"transcript", "undeclared"}
		w.problem = domain.CodeUnknownDocument
	case "a step unreachable from the first one":
		d.Steps = append(d.Steps, domain.Step{ID: "orphan", Name: extraName, Role: domain.RoleStaff, Area: "student_affairs", Profile: domain.ProfileStaffOperative, Action: domain.ActionConfirmExternalFact})
		w.problem = domain.CodeUnreachableStep
	case "a step that never reaches a closing step":
		d.Steps[lastStep].Next = "review"
		w.problem = domain.CodeStepNeverCloses
	case "a field no step uses":
		d.Fields = append(d.Fields, domain.Field{ID: "unused", Name: extraName, Type: domain.FieldText})
		w.problem = domain.CodeUnusedField
	case "a staff step without a profile":
		d.Steps[decision].Profile = ""
		w.problem = domain.CodeInvalidProfile
	case "an applicant step in a staff area":
		d.Steps[0].Area = "student_affairs"
		w.problem = domain.CodeApplicantWithArea
	default:
		return fmt.Errorf("unknown problem %q", problem)
	}
	return w.publish(0, d)
}

func (w *procedureTypeWorld) renameArea(name string) error {
	a, err := w.areas.Get(context.Background(), areaID(name))
	if err != nil {
		return err
	}
	w.previousName = a.Name
	a.Name += " renamed"
	return w.areas.Save(context.Background(), a)
}

func (w *procedureTypeWorld) askForVersion(number int) error {
	w.requested, w.err = w.service().Version(context.Background(), procedureType, number)
	return nil
}

func (w *procedureTypeWorld) thenPublished(number int, name string) error {
	if w.err != nil {
		return w.err
	}
	if w.published.Number != number {
		return fmt.Errorf("published version %d, want %d", w.published.Number, number)
	}
	c, err := w.published.Content()
	if err != nil {
		return err
	}
	if c.Definition.Name != name {
		return fmt.Errorf("published %q, want %q", c.Definition.Name, name)
	}
	return nil
}

func (w *procedureTypeWorld) thenRecordsAuthorship() error {
	if w.published.Authorship != w.authors {
		return fmt.Errorf("recorded %+v, want %+v", w.published.Authorship, w.authors)
	}
	if !w.published.PublishedAt.Equal(w.now) {
		return fmt.Errorf("published at %s, want %s", w.published.PublishedAt, w.now)
	}
	return nil
}

func (w *procedureTypeWorld) thenLatestIs(number int) error {
	if w.err != nil {
		return w.err
	}
	latest, err := w.latest()
	if err != nil {
		return err
	}
	if latest != number {
		return fmt.Errorf("the latest version is %d, want %d", latest, number)
	}
	return nil
}

func (w *procedureTypeWorld) thenStillRecords(proposer, approver string) error {
	stored, err := w.versions.Latest(context.Background(), procedureType)
	if err != nil {
		return err
	}
	want := domain.Authorship{ProposedBy: proposer, ApprovedBy: approver}
	if stored.Authorship != want || w.published.Authorship != want {
		return fmt.Errorf("stored %+v and returned %+v, want %+v", stored.Authorship, w.published.Authorship, want)
	}
	return nil
}

func (w *procedureTypeWorld) thenNoVersion() error {
	if w.err == nil {
		return errors.New("the publication did not fail")
	}
	latest, err := w.latest()
	if err != nil {
		return err
	}
	if latest != w.before {
		return fmt.Errorf("the latest version went from %d to %d", w.before, latest)
	}
	return nil
}

func (w *procedureTypeWorld) thenAskedToStartAgain(number int) error {
	outdated, ok := errors.AsType[*input.OutdatedBaseError](w.err)
	if !ok {
		return fmt.Errorf("got %w, want ErrOutdatedBase", w.err)
	}
	if outdated.Latest != number {
		return fmt.Errorf("asked to start again from version %d, want %d", outdated.Latest, number)
	}
	return nil
}

func (w *procedureTypeWorld) thenProblemReported() error {
	for _, p := range domain.ProblemsOf(w.err) {
		if p.Code == w.problem {
			return nil
		}
	}
	return fmt.Errorf("got %w, want a problem with code %s", w.err, w.problem)
}

func (w *procedureTypeWorld) thenProcedureFollows(number int) error {
	v, err := w.service().Version(context.Background(), procedureType, w.followed.Number)
	if err != nil {
		return err
	}
	if v.Number != number || v.Digest != w.followed.Digest {
		return fmt.Errorf("follows version %d with digest %s, want version %d with digest %s", v.Number, v.Digest, number, w.followed.Digest)
	}
	return nil
}

func (w *procedureTypeWorld) thenShowsPreviousName(number int) error {
	v, err := w.service().Version(context.Background(), procedureType, number)
	if err != nil {
		return err
	}
	c, err := v.Content()
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(c.Areas, func(a domain.Area) bool { return a.Name == w.previousName }) {
		return fmt.Errorf("version %d shows %+v, want the name %q", number, c.Areas, w.previousName)
	}
	return nil
}

func (w *procedureTypeWorld) thenVersionNotReturned() error {
	if !errors.Is(w.err, domain.ErrAlteredContent) {
		return fmt.Errorf("got %w, want ErrAlteredContent", w.err)
	}
	if w.requested.Number != 0 {
		return fmt.Errorf("returned version %d", w.requested.Number)
	}
	return nil
}

func (w *procedureTypeWorld) thenNothingOnTop() error {
	latest, err := w.latest()
	if err != nil {
		return err
	}
	if err := w.publish(latest, w.revised()); err != nil {
		return err
	}
	if !errors.Is(w.err, domain.ErrAlteredContent) {
		return fmt.Errorf("got %w, want ErrAlteredContent", w.err)
	}
	return w.thenNoVersion()
}
