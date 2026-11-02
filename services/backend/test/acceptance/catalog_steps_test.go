package acceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/cucumber/godog"

	"backend/src/adapters/memory"
	"backend/src/application"
	"backend/src/domain/procedure"
	"backend/src/ports/input"
)

// catalogWorld is the state of one scenario of US13.
type catalogWorld struct {
	service *application.ProcedureTypeService
	types   map[string]int64
	latest  map[string]procedure.Definition

	entries []input.CatalogEntry
	err     error
	publish error
}

func registerCatalogSteps(sc *godog.ScenarioContext) {
	versions := &tamperableVersions{ProcedureVersions: memory.NewProcedureVersions()}
	areas := memory.NewAreaCatalog(
		procedure.Area{ID: "titular", Name: "Titular", Kind: procedure.AreaApplicant},
		procedure.Area{ID: "secretaria", Name: "Secretaría", Kind: procedure.AreaOffice},
	)
	w := &catalogWorld{
		service: application.NewProcedureTypeService(versions, areas, time.Now),
		types:   map[string]int64{},
		latest:  map[string]procedure.Definition{},
	}

	sc.Given(`^a published "([^"]*)" procedure type described as "([^"]*)"$`, w.givenPublished)
	sc.Given(`^the "([^"]*)" field is also mandatory in a later applicant step$`, w.givenMandatoryLater)
	sc.Given(`^the catalog was opened$`, w.open)
	sc.Given(`^a new version of "([^"]*)" described as "([^"]*)"$`, w.givenNewVersion)
	sc.Given(`^the stored content of the latest version no longer matches its digest$`, func() { versions.altered = true })

	sc.When(`^the catalog is opened$`, w.open)
	sc.When(`^a "([^"]*)" procedure type is published$`, w.publishNew)
	sc.When(`^a definition with (.+) is published$`, w.publishBroken)

	sc.Then(`^it lists (\d+) procedure types$`, w.thenLists)
	sc.Then(`^"([^"]*)" is described as "([^"]*)"$`, w.thenDescribed)
	sc.Then(`^"([^"]*)" shows the "([^"]*)" field as (mandatory|optional)$`, w.thenField)
	sc.Then(`^"([^"]*)" shows the "([^"]*)" document as (mandatory|optional)$`, w.thenDocument)
	sc.Then(`^"([^"]*)" does not show the "([^"]*)" field$`, w.thenNoField)
	sc.Then(`^"([^"]*)" does not show the "([^"]*)" document$`, w.thenNoDocument)
	sc.Then(`^the "([^"]*)" document accepts "([^"]*)" up to (\d+) bytes$`, w.thenAccepts)
	sc.Then(`^"([^"]*)" shows the deadline "([^"]*)"$`, w.thenDeadline)
	sc.Then(`^"([^"]*)" shows the admissibility condition "([^"]*)"$`, w.thenCondition)
	sc.Then(`^the catalog is not shown$`, w.thenNotShown)
	sc.Then(`^the publication is rejected with the code "([^"]*)"$`, w.thenRejected)
}

// tamperableVersions alters what the store returns, as a change made behind
// SIGTA's back would.
type tamperableVersions struct {
	*memory.ProcedureVersions
	altered bool
}

func (s *tamperableVersions) LatestOfEach(ctx context.Context) ([]procedure.Version, error) {
	versions, err := s.ProcedureVersions.LatestOfEach(ctx)
	if err == nil && s.altered {
		versions[0].Encoded = append(versions[0].Encoded, ' ')
	}
	return versions, err
}

func catalogDefinition(name, description string) procedure.Definition {
	return procedure.Definition{
		Name:          name,
		Description:   description,
		Deadlines:     []string{"Hasta 30 días antes de cursar"},
		Admissibility: []string{"Ser estudiante regular"},
		Fields: []procedure.Field{
			{ID: "full_name", Name: "Nombre completo", Type: procedure.FieldText},
			{ID: "reason", Name: "Motivo", Type: procedure.FieldLongText},
			{ID: "file_number", Name: "Número de expediente", Type: procedure.FieldText},
		},
		Documents: []procedure.Document{
			{ID: "transcript", Name: "Certificado analítico", AcceptedTypes: []string{"application/pdf"}, MaxSize: 10 << 20},
			{ID: "program", Name: "Programa", AcceptedTypes: []string{"application/pdf"}, MaxSize: 10 << 20},
			{ID: "ruling", Name: "Dictamen", AcceptedTypes: []string{"application/pdf"}, MaxSize: 10 << 20},
		},
		Steps: []procedure.Step{
			{ID: "request", Name: "Solicitud", Area: "titular", Action: procedure.ActionEnterData, Fields: []string{"full_name"}, OptionalFields: []string{"reason"}, Next: "documents"},
			{ID: "documents", Name: "Documentación", Area: "titular", Action: procedure.ActionUploadDocument, Documents: []string{"transcript"}, OptionalDocuments: []string{"program"}, Next: "ruling"},
			{ID: "ruling", Name: "Dictamen", Area: "secretaria", Profile: procedure.ProfileStaffOperative, Action: procedure.ActionUploadDocument, Documents: []string{"ruling"}, Next: "record"},
			{ID: "record", Name: "Registro", Area: "secretaria", Profile: procedure.ProfileStaffOperative, Action: procedure.ActionConfirmExternalFact, Fields: []string{"file_number"}},
		},
	}
}

func (w *catalogWorld) publishDefinition(name string, d procedure.Definition) error {
	id, known := w.types[name]
	if !known {
		id = int64(len(w.types) + 1)
		w.types[name] = id
	}
	base := 0
	if known {
		v, err := w.latestNumber(id)
		if err != nil {
			return err
		}
		base = v
	}
	_, err := w.service.Publish(context.Background(), input.Publication{
		ProcedureType: id,
		Base:          base,
		Definition:    d,
		Authorship:    procedure.Authorship{ProposedBy: "ana", ApprovedBy: "beto"},
	})
	w.latest[name] = d
	return err
}

// latestNumber returns the number of the latest version shown by the catalog.
func (w *catalogWorld) latestNumber(id int64) (int, error) {
	entries, err := w.service.Catalog(context.Background())
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		if e.ProcedureType == id {
			return e.Version, nil
		}
	}
	return 0, nil
}

func (w *catalogWorld) givenPublished(name, description string) error {
	return w.publishDefinition(name, catalogDefinition(name, description))
}

func (w *catalogWorld) givenNewVersion(name, description string) error {
	d := w.latest[name]
	d.Description = description
	return w.publishDefinition(name, d)
}

func (w *catalogWorld) givenMandatoryLater(field string) error {
	for name, d := range w.latest {
		d.Steps = slices.Clone(d.Steps)
		d.Steps[0].Next = "more"
		d.Steps = slices.Insert(d.Steps, 1, procedure.Step{
			ID: "more", Name: "Ampliación", Area: "titular", Action: procedure.ActionEnterData, Fields: []string{field}, Next: "documents",
		})
		if err := w.publishDefinition(name, d); err != nil {
			return err
		}
	}
	return nil
}

func (w *catalogWorld) open() {
	w.entries, w.err = w.service.Catalog(context.Background())
}

func (w *catalogWorld) publishNew(name string) error {
	return w.publishDefinition(name, catalogDefinition(name, "Un trámite nuevo"))
}

func (w *catalogWorld) publishBroken(problem string) error {
	d := catalogDefinition("broken", "Un trámite")
	switch problem {
	case "no description":
		d.Description = ""
	case "a document without formats":
		d.Documents[0].AcceptedTypes = nil
	case "a document with a wrong format":
		d.Documents[0].AcceptedTypes = []string{"pdf"}
	case "a document without a size limit":
		d.Documents[0].MaxSize = 0
	case "a blank deadline":
		d.Deadlines = []string{" "}
	default:
		return fmt.Errorf("unknown problem %q", problem)
	}
	w.publish = w.publishDefinition("broken", d)
	return nil
}

func (w *catalogWorld) entry(name string) (input.CatalogEntry, error) {
	if w.err != nil {
		return input.CatalogEntry{}, w.err
	}
	for _, e := range w.entries {
		if e.Name == name {
			return e, nil
		}
	}
	return input.CatalogEntry{}, fmt.Errorf("%q is not in the catalog", name)
}

func (w *catalogWorld) thenLists(count int) error {
	if w.err != nil {
		return w.err
	}
	if len(w.entries) != count {
		return fmt.Errorf("the catalog lists %d procedure types, want %d", len(w.entries), count)
	}
	return nil
}

func (w *catalogWorld) thenDescribed(name, description string) error {
	e, err := w.entry(name)
	if err != nil {
		return err
	}
	if e.Description != description {
		return fmt.Errorf("description = %q, want %q", e.Description, description)
	}
	return nil
}

func (w *catalogWorld) thenField(name, id, kind string) error {
	e, err := w.entry(name)
	if err != nil {
		return err
	}
	for _, f := range e.Fields {
		if f.Field.ID == id {
			if f.Required != (kind == "mandatory") {
				return fmt.Errorf("field %q required = %v, want %s", id, f.Required, kind)
			}
			return nil
		}
	}
	return fmt.Errorf("field %q is not shown", id)
}

func (w *catalogWorld) document(name, id string) (procedure.DocumentRequirement, bool, error) {
	e, err := w.entry(name)
	if err != nil {
		return procedure.DocumentRequirement{}, false, err
	}
	for _, d := range e.Documents {
		if d.Document.ID == id {
			return d, true, nil
		}
	}
	return procedure.DocumentRequirement{}, false, nil
}

func (w *catalogWorld) thenDocument(name, id, kind string) error {
	d, found, err := w.document(name, id)
	switch {
	case err != nil:
		return err
	case !found:
		return fmt.Errorf("document %q is not shown", id)
	case d.Required != (kind == "mandatory"):
		return fmt.Errorf("document %q required = %v, want %s", id, d.Required, kind)
	}
	return nil
}

func (w *catalogWorld) thenNoField(name, id string) error {
	e, err := w.entry(name)
	if err != nil {
		return err
	}
	for _, f := range e.Fields {
		if f.Field.ID == id {
			return fmt.Errorf("field %q is shown", id)
		}
	}
	return nil
}

func (w *catalogWorld) thenNoDocument(name, id string) error {
	_, found, err := w.document(name, id)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("document %q is shown", id)
	}
	return nil
}

func (w *catalogWorld) thenAccepts(id, mediaType string, size int64) error {
	if w.err != nil {
		return w.err
	}
	for _, e := range w.entries {
		d, found, err := w.document(e.Name, id)
		if err != nil {
			return err
		}
		if found {
			if !slices.Contains(d.Document.AcceptedTypes, mediaType) || d.Document.MaxSize != size {
				return fmt.Errorf("document %q = %+v", id, d.Document)
			}
			return nil
		}
	}
	return fmt.Errorf("document %q is not shown", id)
}

func (w *catalogWorld) thenDeadline(name, text string) error {
	e, err := w.entry(name)
	if err != nil {
		return err
	}
	if !slices.Contains(e.Deadlines, text) {
		return fmt.Errorf("deadlines = %q", e.Deadlines)
	}
	return nil
}

func (w *catalogWorld) thenCondition(name, text string) error {
	e, err := w.entry(name)
	if err != nil {
		return err
	}
	if !slices.Contains(e.Admissibility, text) {
		return fmt.Errorf("conditions = %q", e.Admissibility)
	}
	return nil
}

func (w *catalogWorld) thenNotShown() error {
	if !errors.Is(w.err, procedure.ErrAlteredContent) {
		return fmt.Errorf("Catalog() error = %w, want %w", w.err, procedure.ErrAlteredContent)
	}
	return nil
}

func (w *catalogWorld) thenRejected(code string) error {
	if !errors.Is(w.publish, procedure.ErrInvalidDefinition) {
		return fmt.Errorf("Publish() error = %w, want an invalid definition", w.publish)
	}
	for _, p := range procedure.ProblemsOf(w.publish) {
		if string(p.Code) == code {
			return nil
		}
	}
	return fmt.Errorf("problems = %v, want %q", procedure.ProblemsOf(w.publish), code)
}
