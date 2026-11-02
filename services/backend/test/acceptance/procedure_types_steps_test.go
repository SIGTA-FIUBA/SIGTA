package acceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"backend/src/adapters/memory"
	"backend/src/application"
	"backend/src/domain/procedure"
	"backend/src/ports/input"
	"backend/src/ports/output"
)

const procedureType int64 = 1

// procedureTypeWorld is the state of one scenario of US08.
type procedureTypeWorld struct {
	areas    []procedure.Area
	versions *memory.ProcedureVersions
	authors  procedure.Authorship
	now      time.Time

	definition procedure.Definition
	revisions  int

	// The outcome of the last publication.
	published procedure.Version
	err       error
	before    int
	problem   procedure.Code

	followed     procedure.Version
	previousName string
	requested    procedure.Version
}

func registerProcedureTypeSteps(sc *godog.ScenarioContext) {
	w := &procedureTypeWorld{
		versions: memory.NewProcedureVersions(),
		now:      time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC),
	}

	sc.Given(`^the areas "([^"]*)", "([^"]*)" and "([^"]*)"$`, w.givenAreas)
	sc.Given(`^"([^"]*)" proposes and "([^"]*)" approves every publication$`, w.givenAuthors)
	sc.Given(`^version (\d+) of "([^"]*)" is published$`, w.givenPublished)
	sc.Given(`^version (\d+) of "([^"]*)" was proposed by "([^"]*)" and approved by "([^"]*)"$`, w.givenPublishedBy)
	sc.Given(`^version (\d+) of "([^"]*)" was published from version \d+$`, w.givenPublished)
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
	return application.NewProcedureTypeService(w.versions, memory.NewAreaCatalog(w.areas...), func() time.Time { return w.now })
}

func (w *procedureTypeWorld) latest() (int, error) {
	v, err := w.versions.Latest(context.Background(), procedureType)
	if errors.Is(err, output.ErrNoVersion) {
		return 0, nil
	}
	return v.Number, err
}

func (w *procedureTypeWorld) publish(base int, d procedure.Definition) error {
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

// valid is a definition that uses the four actions and every kind of area.
func (w *procedureTypeWorld) valid(name string) procedure.Definition {
	return procedure.Definition{
		Name:      name,
		Fields:    []procedure.Field{{ID: "full_name", Name: "Nombre completo", Type: procedure.FieldText}},
		Documents: []procedure.Document{{ID: "transcript", Name: "Certificado analítico"}, {ID: "ruling", Name: "Dictamen"}},
		Steps: []procedure.Step{
			{ID: "request", Name: "Solicitud", Area: "applicant", Action: procedure.ActionEnterData, Fields: []string{"full_name"}, Next: "documents"},
			{ID: "documents", Name: "Documentación", Area: "applicant", Action: procedure.ActionUploadDocument, Documents: []string{"transcript"}, Next: "review"},
			{
				ID: "review", Name: "Revisión", Area: "student_affairs", Profile: procedure.ProfileStaffOperative, Action: procedure.ActionDecide,
				Outcomes: []procedure.Outcome{
					{ID: "complete", Name: "Completa", Next: "evaluation"},
					{ID: "incomplete", Name: "Incompleta", Next: "documents"},
				},
			},
			{ID: "evaluation", Name: "Evaluación", Area: "curricular_commission", Profile: procedure.ProfileStaffOperative, Action: procedure.ActionUploadDocument, Documents: []string{"ruling"}, Next: "signature"},
			{ID: "signature", Name: "Firma", Area: "student_affairs", Profile: procedure.ProfileStaffApprover, Action: procedure.ActionConfirmExternalFact},
		},
	}
}

// revised returns the current definition with a visible change.
func (w *procedureTypeWorld) revised() procedure.Definition {
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

func (w *procedureTypeWorld) givenAreas(applicant, office, external string) error {
	w.areas = []procedure.Area{
		{ID: areaID(applicant), Name: applicant, Kind: procedure.AreaApplicant},
		{ID: areaID(office), Name: office, Kind: procedure.AreaOffice},
		{ID: areaID(external), Name: external, Kind: procedure.AreaExternal},
	}
	return nil
}

func areaID(name string) string { return strings.ReplaceAll(name, " ", "_") }

func (w *procedureTypeWorld) givenAuthors(proposer, approver string) error {
	w.authors = procedure.Authorship{ProposedBy: proposer, ApprovedBy: approver}
	return nil
}

func (w *procedureTypeWorld) givenPublished(number int, name string) error {
	w.definition = w.valid(name)
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
	w.definition = w.valid("subject equivalence")
	encoded, err := procedure.NewContent(w.definition, w.areas).Encode()
	if err != nil {
		return err
	}
	return w.versions.Add(context.Background(), procedure.Version{
		ProcedureType: procedureType,
		Number:        number,
		Digest:        procedure.Digest(encoded),
		Encoded:       append(encoded, ' '),
		Authorship:    w.authors,
		PublishedAt:   w.now,
	})
}

func (w *procedureTypeWorld) publishValid(name string, base int) error {
	return w.publish(base, w.valid(name))
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
	return w.publish(0, w.valid(name))
}

func (w *procedureTypeWorld) publishInvalid(name, problem string) error {
	d := w.valid(name)
	const (
		document  = 1
		decision  = 2
		lastStep  = 4
		extraName = "Sin usar"
	)
	switch problem {
	case "a step with an action outside the four":
		d.Steps[0].Action = "approve"
		w.problem = procedure.CodeUnknownAction
	case "a step in an unknown area":
		d.Steps[decision].Area = "nowhere"
		w.problem = procedure.CodeUnknownArea
	case "a reference to an undeclared document":
		d.Steps[document].Documents = []string{"transcript", "undeclared"}
		w.problem = procedure.CodeUnknownDocument
	case "a step unreachable from the first one":
		d.Steps = append(d.Steps, procedure.Step{ID: "orphan", Name: extraName, Area: "student_affairs", Profile: procedure.ProfileStaffOperative, Action: procedure.ActionConfirmExternalFact})
		w.problem = procedure.CodeUnreachableStep
	case "a step that never reaches a closing step":
		d.Steps[lastStep].Next = "review"
		w.problem = procedure.CodeStepNeverCloses
	case "a field no step uses":
		d.Fields = append(d.Fields, procedure.Field{ID: "unused", Name: extraName, Type: procedure.FieldText})
		w.problem = procedure.CodeUnusedField
	case "an office step without a profile":
		d.Steps[decision].Profile = ""
		w.problem = procedure.CodeInvalidProfile
	default:
		return fmt.Errorf("unknown problem %q", problem)
	}
	return w.publish(0, d)
}

func (w *procedureTypeWorld) renameArea(name string) error {
	for i, a := range w.areas {
		if a.Name == name {
			w.previousName = a.Name
			w.areas[i].Name = a.Name + " renamed"
			return nil
		}
	}
	return fmt.Errorf("no area named %q", name)
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
	want := procedure.Authorship{ProposedBy: proposer, ApprovedBy: approver}
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
	if !errors.Is(w.err, input.ErrOutdatedBase) {
		return fmt.Errorf("got %w, want ErrOutdatedBase", w.err)
	}
	if want := fmt.Sprintf("the latest is %d", number); !strings.Contains(w.err.Error(), want) {
		return fmt.Errorf("%w does not say %q", w.err, want)
	}
	return nil
}

func (w *procedureTypeWorld) thenProblemReported() error {
	for _, p := range procedure.ProblemsOf(w.err) {
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
	if !slices.ContainsFunc(c.Areas, func(a procedure.Area) bool { return a.Name == w.previousName }) {
		return fmt.Errorf("version %d shows %+v, want the name %q", number, c.Areas, w.previousName)
	}
	return nil
}

func (w *procedureTypeWorld) thenVersionNotReturned() error {
	if !errors.Is(w.err, procedure.ErrAlteredContent) {
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
	if !errors.Is(w.err, procedure.ErrAlteredContent) {
		return fmt.Errorf("got %w, want ErrAlteredContent", w.err)
	}
	return w.thenNoVersion()
}
