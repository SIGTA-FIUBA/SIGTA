package acceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/cucumber/godog"

	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain/catalog"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
)

// catalogWorld is the state of one scenario of US13.
type catalogWorld struct {
	service *application.ProcedureTypeService
	types   map[string]int64
	latest  map[string]definition.ProcedureTypeDefinition

	entries []input.CatalogEntry
	err     error
	publish error
}

func registerCatalogSteps(sc *godog.ScenarioContext) {
	versions := &tamperableVersions{ProcedureVersions: memory.NewProcedureVersions()}
	areas := memory.NewAreaCatalog(
		definition.Area{ID: "titular", Name: "Titular", Kind: definition.AreaApplicant},
		definition.Area{ID: "secretaria", Name: "Secretaría", Kind: definition.AreaOffice},
	)
	world := &catalogWorld{
		service: application.NewProcedureTypeService(versions, areas, time.Now),
		types:   map[string]int64{},
		latest:  map[string]definition.ProcedureTypeDefinition{},
	}

	sc.Given(`^a published "([^"]*)" procedure type described as "([^"]*)"$`, world.givenPublished)
	sc.Given(`^the "([^"]*)" field is also mandatory in a later applicant step$`, world.givenMandatoryLater)
	sc.Given(`^the catalog was opened$`, world.open)
	sc.Given(`^a new version of "([^"]*)" described as "([^"]*)"$`, world.givenNewVersion)
	sc.Given(`^the stored content of the latest version no longer matches its digest$`, func() { versions.altered = true })

	sc.When(`^the catalog is opened$`, world.open)
	sc.When(`^a "([^"]*)" procedure type is published$`, world.publishNew)
	sc.When(`^a definition with (.+) is published$`, world.publishBroken)

	sc.Then(`^it lists (\d+) procedure types?$`, world.thenLists)
	sc.Then(`^"([^"]*)" is described as "([^"]*)"$`, world.thenDescribed)
	sc.Then(`^"([^"]*)" shows the "([^"]*)" field as (mandatory|optional)$`, world.thenField)
	sc.Then(`^"([^"]*)" shows the "([^"]*)" document as (mandatory|optional)$`, world.thenDocument)
	sc.Then(`^"([^"]*)" does not show the "([^"]*)" field$`, world.thenNoField)
	sc.Then(`^"([^"]*)" does not show the "([^"]*)" document$`, world.thenNoDocument)
	sc.Then(`^the "([^"]*)" document accepts "([^"]*)" up to (\d+) bytes$`, world.thenAccepts)
	sc.Then(`^"([^"]*)" shows the deadline "([^"]*)"$`, world.thenDeadline)
	sc.Then(`^"([^"]*)" shows the admissibility condition "([^"]*)"$`, world.thenCondition)
	sc.Then(`^the catalog is not shown$`, world.thenNotShown)
	sc.Then(`^the publication is rejected with the code "([^"]*)"$`, world.thenRejected)
}

// tamperableVersions alters what the store returns, as a change made behind
// SIGTA's back would.
type tamperableVersions struct {
	*memory.ProcedureVersions
	altered bool
}

func (s *tamperableVersions) LatestOfEach(ctx context.Context) ([]version.ProcedureTypeVersion, error) {
	versions, err := s.ProcedureVersions.LatestOfEach(ctx)
	if err == nil && s.altered {
		versions[0].EncodedContent = append(versions[0].EncodedContent, ' ')
	}
	return versions, err
}

func catalogDefinition(name, description string) definition.ProcedureTypeDefinition {
	return definition.ProcedureTypeDefinition{
		Name:          name,
		Description:   description,
		Deadlines:     []string{"Hasta 30 días antes de cursar"},
		Admissibility: []string{"Ser estudiante regular"},
		Fields: []definition.Field{
			{ID: "full_name", Name: "Nombre completo", Type: definition.FieldText},
			{ID: "reason", Name: "Motivo", Type: definition.FieldLongText},
			{ID: "file_number", Name: "Número de expediente", Type: definition.FieldText},
		},
		Documents: []definition.Document{
			{ID: "transcript", Name: "Certificado analítico", AcceptedTypes: []string{"application/pdf"}, MaxSize: 10 << 20},
			{ID: "program", Name: "Programa", AcceptedTypes: []string{"application/pdf"}, MaxSize: 10 << 20},
			{ID: "ruling", Name: "Dictamen", AcceptedTypes: []string{"application/pdf"}, MaxSize: 10 << 20},
		},
		Steps: []definition.Step{
			{ID: "request", Name: "Solicitud", Area: "titular", Action: definition.ActionEnterData, Fields: []string{"full_name"}, OptionalFields: []string{"reason"}, Next: "documents"},
			{ID: "documents", Name: "Documentación", Area: "titular", Action: definition.ActionUploadDocument, Documents: []string{"transcript"}, OptionalDocuments: []string{"program"}, Next: "ruling"},
			{ID: "ruling", Name: "Dictamen", Area: "secretaria", Profile: definition.ProfileStaffOperative, Action: definition.ActionUploadDocument, Documents: []string{"ruling"}, Next: "record"},
			{ID: "record", Name: "Registro", Area: "secretaria", Profile: definition.ProfileStaffOperative, Action: definition.ActionConfirmExternalFact, Fields: []string{"file_number"}},
		},
	}
}

func (w *catalogWorld) publishDefinition(name string, d definition.ProcedureTypeDefinition) error {
	base := 0
	typeID, known := w.types[name]
	if known {
		number, err := w.latestNumber(typeID)
		if err != nil {
			return err
		}
		base = number
	} else {
		typeID = int64(len(w.types) + 1)
		w.types[name] = typeID
	}
	_, err := w.service.Publish(context.Background(), input.Publication{
		ProcedureType: typeID,
		Base:          base,
		Definition:    d,
		Authorship:    version.Authorship{ProposedBy: "ana", ApprovedBy: "beto"},
	})
	if err == nil {
		w.latest[name] = d
	}
	return err
}

// latestNumber returns the number of the latest version shown by the catalog.
func (w *catalogWorld) latestNumber(id int64) (int, error) {
	entries, err := w.service.Catalog(context.Background())
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		if entry.ProcedureType == id {
			return entry.Version, nil
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
		d.Steps = slices.Insert(d.Steps, 1, definition.Step{
			ID: "more", Name: "Ampliación", Area: "titular", Action: definition.ActionEnterData, Fields: []string{field}, Next: "documents",
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
	for _, entry := range w.entries {
		if entry.Name == name {
			return entry, nil
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
	entry, err := w.entry(name)
	if err != nil {
		return err
	}
	if entry.Description != description {
		return fmt.Errorf("description = %q, want %q", entry.Description, description)
	}
	return nil
}

func (w *catalogWorld) thenField(name, id, kind string) error {
	entry, err := w.entry(name)
	if err != nil {
		return err
	}
	for _, field := range entry.Fields {
		if field.Field.ID == id {
			if field.Required != (kind == "mandatory") {
				return fmt.Errorf("field %q required = %v, want %s", id, field.Required, kind)
			}
			return nil
		}
	}
	return fmt.Errorf("field %q is not shown", id)
}

func (w *catalogWorld) document(name, id string) (catalog.DocumentRequirement, bool, error) {
	entry, err := w.entry(name)
	if err != nil {
		return catalog.DocumentRequirement{}, false, err
	}
	for _, requirement := range entry.Documents {
		if requirement.Document.ID == id {
			return requirement, true, nil
		}
	}
	return catalog.DocumentRequirement{}, false, nil
}

func (w *catalogWorld) thenDocument(name, id, kind string) error {
	requirement, found, err := w.document(name, id)
	switch {
	case err != nil:
		return err
	case !found:
		return fmt.Errorf("document %q is not shown", id)
	case requirement.Required != (kind == "mandatory"):
		return fmt.Errorf("document %q required = %v, want %s", id, requirement.Required, kind)
	}
	return nil
}

func (w *catalogWorld) thenNoField(name, id string) error {
	entry, err := w.entry(name)
	if err != nil {
		return err
	}
	for _, field := range entry.Fields {
		if field.Field.ID == id {
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
	for _, entry := range w.entries {
		for _, requirement := range entry.Documents {
			if requirement.Document.ID != id {
				continue
			}
			if !slices.Contains(requirement.Document.AcceptedTypes, mediaType) || requirement.Document.MaxSize != size {
				return fmt.Errorf("document %q = %+v", id, requirement.Document)
			}
			return nil
		}
	}
	return fmt.Errorf("document %q is not shown", id)
}

func (w *catalogWorld) thenDeadline(name, text string) error {
	entry, err := w.entry(name)
	if err != nil {
		return err
	}
	if !slices.Contains(entry.Deadlines, text) {
		return fmt.Errorf("deadlines = %q", entry.Deadlines)
	}
	return nil
}

func (w *catalogWorld) thenCondition(name, text string) error {
	entry, err := w.entry(name)
	if err != nil {
		return err
	}
	if !slices.Contains(entry.Admissibility, text) {
		return fmt.Errorf("conditions = %q", entry.Admissibility)
	}
	return nil
}

func (w *catalogWorld) thenNotShown() error {
	if !errors.Is(w.err, version.ErrAlteredContent) {
		return fmt.Errorf("Catalog() error = %w, want %w", w.err, version.ErrAlteredContent)
	}
	return nil
}

func (w *catalogWorld) thenRejected(code string) error {
	if !errors.Is(w.publish, definition.ErrInvalidDefinition) {
		return fmt.Errorf("Publish() error = %w, want an invalid definition", w.publish)
	}
	for _, problem := range definition.ValidationProblemsOf(w.publish) {
		if string(problem.Code) == code {
			return nil
		}
	}
	return fmt.Errorf("problems = %v, want %q", definition.ValidationProblemsOf(w.publish), code)
}
