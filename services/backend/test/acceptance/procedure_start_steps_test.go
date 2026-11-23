package acceptance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"iter"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/disk"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/submission"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

const (
	submittedType   = 1
	unpublishedType = 99
)

// submissionWorld builds the procedure type from three tables, fields,
// documents and steps, and publishes it with the last one.
type submissionWorld struct {
	areas       *authapp.AreaService
	memberships *authapp.MembershipService
	people      *authmem.UserRepository
	types       *application.ProcedureTypeService
	procedures  *memory.ProcedureRepository
	documents   *memory.DocumentRepository
	service     *application.SubmissionService
	dir         string

	def    definition.ProcedureTypeDefinition
	known  map[string]bool
	nextID int

	values  map[string]submission.Value
	files   map[string][]byte
	started input.Started
	err     error
}

func registerSubmissionSteps(sc *godog.ScenarioContext) {
	w := &submissionWorld{}
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return ctx, w.setUp()
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		if w.dir == "" {
			return ctx, err
		}
		return ctx, errors.Join(err, os.RemoveAll(w.dir))
	})

	sc.Given(`^a published procedure type with these fields:$`, w.givenFields)
	sc.Given(`^these documents:$`, w.givenDocuments)
	sc.Given(`^these steps:$`, w.givenSteps)
	sc.Given(`^"([^"]*)" is staff of "([^"]*)" with the "([^"]*)" permission$`, w.givenStaff)

	sc.When(`^"([^"]*)" submits the procedure with every mandatory requirement$`, func(user string) error {
		return w.submit(user, submittedType, nil, "")
	})
	sc.When(`^"([^"]*)" submits the procedure with every mandatory requirement and:$`, func(user string, table *godog.Table) error {
		return w.submit(user, submittedType, table, "")
	})
	sc.When(`^"([^"]*)" submits the procedure with every mandatory requirement but the "([^"]*)"$`, func(user, left string) error {
		return w.submit(user, submittedType, nil, left)
	})
	sc.When(`^"([^"]*)" submits a procedure of a type that is not published$`, func(user string) error {
		return w.submit(user, unpublishedType, nil, "")
	})

	sc.Then(`^the procedure is started with an identifier$`, w.thenStarted)
	sc.Then(`^its applicant is "([^"]*)"$`, w.thenApplicant)
	sc.Then(`^it waits on the "([^"]*)" step, in an office$`, w.thenWaitsInOffice)
	sc.Then(`^the procedure records "([^"]*)" as "([^"]*)" on the "([^"]*)" step$`, w.thenRecords)
	sc.Then(`^the procedure records no "([^"]*)"$`, w.thenRecordsNo)
	sc.Then(`^the "([^"]*)" of the procedure is uploaded by "([^"]*)" on the "([^"]*)" step$`, w.thenUploaded)
	sc.Then(`^the procedure has no "([^"]*)"$`, w.thenHasNo)
	sc.Then(`^the submission is rejected because the "([^"]*)" (is missing|is not valid|is not asked when starting)$`, w.thenRejectedFor)
	sc.Then(`^the submission is rejected because (its type is not accepted|it is too large|it is empty)$`, w.thenFileRejected)
	sc.Then(`^no procedure is started$`, w.thenNothingStarted)
	sc.Then(`^the procedure type is not found$`, func() error {
		if !errors.Is(w.err, output.ErrVersionNotFound) {
			return fmt.Errorf("the submission failed with %w, want no published version", w.err)
		}
		return nil
	})
}

func (w *submissionWorld) setUp() error {
	*w = submissionWorld{known: map[string]bool{}}
	dir, err := os.MkdirTemp("", "sigta-submissions-")
	if err != nil {
		return err
	}
	w.dir = dir
	files, err := disk.NewFiles(dir)
	if err != nil {
		return err
	}
	areas := authmem.NewAreaRepository()
	if err := authapp.EnsureSecretariat(context.Background(), areas); err != nil {
		return err
	}
	administrators := authmem.NewAdministratorRepository(scenarioAdministrator)
	stored := authmem.NewMembershipRepository(
		authdomain.Membership{ID: "M-proposer", UserID: defaultAuthors.ProposedBy, Area: authinput.SecretariatArea, Permissions: []string{authinput.EditProcedureTypes}},
		authdomain.Membership{ID: "M-approver", UserID: defaultAuthors.ApprovedBy, Area: authinput.SecretariatArea, Permissions: []string{authinput.PublishProcedureTypes}},
	)
	w.areas = authapp.NewAreaService(areas, administrators, stored, time.Now)
	w.people = authmem.NewUserRepository()
	w.memberships = authapp.NewMembershipService(areas, administrators, stored, w.people, w.newID, time.Now)
	w.types = application.NewProcedureTypeService(memory.NewProcedureTypeVersionRepository(), w.areas, w.memberships, time.Now)
	w.procedures = memory.NewProcedureRepository()
	w.documents = memory.NewDocumentRepository(w.procedures)
	procedures := application.NewProcedureService(w.types, w.procedures, w.documents, w.memberships, w.newID, time.Now)
	w.service = application.NewSubmissionService(procedures, files, w.newID, time.Now)
	return nil
}

func (w *submissionWorld) newID() string {
	w.nextID++
	return "S-" + strconv.Itoa(w.nextID)
}

func (w *submissionWorld) givenFields(table *godog.Table) error {
	w.def.Name = "Trámite para iniciar"
	w.def.Description = "Un trámite para las pruebas de inicio"
	for _, row := range table.Rows[1:] {
		field := definition.Field{ID: row.Cells[0].Value, Name: row.Cells[0].Value, Type: definition.FieldType(row.Cells[1].Value)}
		for _, option := range list(row.Cells[2].Value) {
			field.Options = append(field.Options, definition.Option{ID: option, Name: strings.ToUpper(option[:1]) + option[1:]})
		}
		w.def.Fields = append(w.def.Fields, field)
	}
	return nil
}

func (w *submissionWorld) givenDocuments(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		maxSize, err := strconv.ParseInt(row.Cells[2].Value, 10, 64)
		if err != nil {
			return err
		}
		w.def.Documents = append(w.def.Documents, definition.Document{
			ID: row.Cells[0].Value, Name: row.Cells[0].Value, AcceptedTypes: list(row.Cells[1].Value), MaxSizeBytes: maxSize,
		})
	}
	return nil
}

// givenSteps reads "the applicant" or an area and a quoted permission as who
// does each step, which enters data or uploads documents as what it asks for
// is a field or a document. It chains the steps in order and publishes.
func (w *submissionWorld) givenSteps(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		step := definition.Step{ID: row.Cells[0].Value, Name: row.Cells[0].Value}
		if doneBy := row.Cells[1].Value; doneBy == "the applicant" {
			step.Role = definition.RoleApplicant
		} else {
			area, permission, found := strings.Cut(doneBy, ", ")
			if !found {
				return fmt.Errorf("unknown author %q", doneBy)
			}
			permission = strings.Trim(permission, `"`)
			if err := ensure(w.areas, w.known, idOf(area), area, permission); err != nil {
				return err
			}
			step.Role, step.Area, step.Permissions = definition.RoleStaff, idOf(area), []string{permission}
		}
		mandatory, optional := list(row.Cells[2].Value), list(row.Cells[3].Value)
		if w.isDocument(slices.Concat(mandatory, optional)[0]) {
			step.Action, step.Documents, step.OptionalDocuments = definition.ActionUploadDocument, mandatory, optional
		} else {
			step.Action, step.Fields, step.OptionalFields = definition.ActionEnterData, mandatory, optional
		}
		if n := len(w.def.Steps); n > 0 {
			w.def.Steps[n-1].Next = step.ID
		}
		w.def.Steps = append(w.def.Steps, step)
	}
	_, err := w.types.Publish(context.Background(), input.Publication{
		ProcedureType: submittedType,
		Definition:    w.def,
		Authorship:    version.Authorship{ProposedBy: "proponente", ApprovedBy: "aprobador"},
	})
	return err
}

func (w *submissionWorld) givenStaff(user, area, permission string) error {
	if err := ensure(w.areas, w.known, idOf(area), area, permission); err != nil {
		return err
	}
	w.people.Add(signedIn(user))
	_, err := w.memberships.Grant(context.Background(), scenarioAdministrator, authinput.MembershipGrant{
		UserID: user, Area: idOf(area), Permissions: []string{permission},
	})
	return err
}

func (w *submissionWorld) isDocument(id string) bool {
	return slices.ContainsFunc(w.def.Documents, func(d definition.Document) bool { return d.ID == id })
}

func (w *submissionWorld) field(id string) (definition.Field, bool) {
	i := slices.IndexFunc(w.def.Fields, func(f definition.Field) bool { return f.ID == id })
	if i < 0 {
		return definition.Field{}, false
	}
	return w.def.Fields[i], true
}

// set reads a value as a file for a document, as a list split at commas for
// a list of subjects and as text otherwise. A value in quotes keeps its
// spaces.
func (w *submissionWorld) set(requirement, value string) error {
	if unquoted, err := strconv.Unquote(value); err == nil {
		value = unquoted
	}
	if w.isDocument(requirement) {
		content, err := fileContent(value)
		if err != nil {
			return err
		}
		w.files[requirement] = content
		return nil
	}
	if f, ok := w.field(requirement); ok && f.Type == definition.FieldSubjectList {
		w.values[requirement] = submission.Value{Items: strings.Split(value, ",")}
		return nil
	}
	w.values[requirement] = submission.Value{Text: value}
	return nil
}

func (w *submissionWorld) submit(user string, procedureType int64, extra *godog.Table, left string) error {
	w.values, w.files = map[string]submission.Value{}, map[string][]byte{}
	mandatory := [][2]string{
		{"reason", "Cursé la materia en otra facultad"},
		{"credits", "8"},
		{"since", "2026-03-01"},
		{"modality", "full"},
		{"subjects", "Análisis Matemático I"},
		{"transcript", "a PDF of 200 bytes"},
	}
	for _, m := range mandatory {
		if m[0] == left {
			continue
		}
		if err := w.set(m[0], m[1]); err != nil {
			return err
		}
	}
	if extra != nil {
		for _, row := range extra.Rows[1:] {
			if err := w.set(row.Cells[0].Value, row.Cells[1].Value); err != nil {
				return err
			}
		}
	}
	w.started, w.err = w.service.Submit(context.Background(), user, input.Submission{
		ProcedureType: procedureType,
		Values:        w.values,
		Files:         w.submittedFiles(),
	})
	return nil
}

// submittedFiles yields the files in the order the definition declares their
// documents, so that each run sends them alike.
func (w *submissionWorld) submittedFiles() iter.Seq2[input.SubmittedFile, error] {
	return func(yield func(input.SubmittedFile, error) bool) {
		for _, d := range w.def.Documents {
			content, ok := w.files[d.ID]
			if !ok {
				continue
			}
			file := input.SubmittedFile{Requirement: d.ID, FileName: d.ID + ".pdf", Content: bytes.NewReader(content)}
			if !yield(file, nil) {
				return
			}
		}
	}
}

func (w *submissionWorld) thenStarted() error {
	if w.err != nil {
		return fmt.Errorf("the submission failed: %w", w.err)
	}
	if w.started.Procedure.ID == "" {
		return errors.New("the procedure has no identifier")
	}
	_, err := w.procedures.Get(context.Background(), w.started.Procedure.ID)
	return err
}

func (w *submissionWorld) thenApplicant(user string) error {
	if w.started.Procedure.Applicant != user {
		return fmt.Errorf("the applicant is %q", w.started.Procedure.Applicant)
	}
	return nil
}

func (w *submissionWorld) thenWaitsInOffice(step string) error {
	p, err := w.procedures.Get(context.Background(), w.started.Procedure.ID)
	if err != nil {
		return err
	}
	state, err := domain.StateOf(p, w.def)
	if err != nil {
		return err
	}
	if p.OpenStep != step || state != domain.StateInOffice {
		return fmt.Errorf("the procedure waits on %q, %s", p.OpenStep, state)
	}
	if w.started.State != state {
		return fmt.Errorf("the start answers %s, the procedure is %s", w.started.State, state)
	}
	return nil
}

func (w *submissionWorld) value(field string) (domain.FieldValue, bool, error) {
	values, err := w.procedures.Values(context.Background(), w.started.Procedure.ID)
	if err != nil {
		return domain.FieldValue{}, false, err
	}
	i := slices.IndexFunc(values, func(v domain.FieldValue) bool { return v.Field == field })
	if i < 0 {
		return domain.FieldValue{}, false, nil
	}
	return values[i], true, nil
}

func (w *submissionWorld) thenRecords(field, want, step string) error {
	if w.err != nil {
		return w.err
	}
	v, ok, err := w.value(field)
	if err != nil {
		return err
	}
	got := v.Text
	if v.Items != nil {
		got = strings.Join(v.Items, ", ")
	}
	if !ok || got != want || v.Step != step || v.EnteredBy != w.started.Procedure.Applicant {
		return fmt.Errorf("the %s is %+v", field, v)
	}
	return nil
}

func (w *submissionWorld) thenRecordsNo(field string) error {
	v, ok, err := w.value(field)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("the %s is %+v", field, v)
	}
	return nil
}

func (w *submissionWorld) document(requirement string) (domain.Document, bool, error) {
	documents, err := w.documents.List(context.Background(), w.started.Procedure.ID)
	if err != nil {
		return domain.Document{}, false, err
	}
	i := slices.IndexFunc(documents, func(d domain.Document) bool { return d.Requirement == requirement })
	if i < 0 {
		return domain.Document{}, false, nil
	}
	return documents[i], true, nil
}

func (w *submissionWorld) thenUploaded(requirement, user, step string) error {
	if w.err != nil {
		return w.err
	}
	d, ok, err := w.document(requirement)
	if err != nil {
		return err
	}
	if !ok || d.UploadedBy != user || d.Step != step {
		return fmt.Errorf("the %s is %+v", requirement, d)
	}
	return nil
}

func (w *submissionWorld) thenHasNo(requirement string) error {
	d, ok, err := w.document(requirement)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("the %s is %+v", requirement, d)
	}
	return nil
}

func (w *submissionWorld) problems() ([]submission.Problem, error) {
	var rejected *submission.Error
	if !errors.As(w.err, &rejected) {
		return nil, fmt.Errorf("the submission was not rejected for its content: %w", w.err)
	}
	return rejected.Problems, nil
}

var reasonCodes = map[string]submission.ProblemCode{
	"is missing":                 submission.CodeMissing,
	"is not valid":               submission.CodeInvalid,
	"is not asked when starting": submission.CodeNotAsked,
	"its type is not accepted":   submission.CodeTypeNotAccepted,
	"it is too large":            submission.CodeTooLarge,
	"it is empty":                submission.CodeEmpty,
}

func (w *submissionWorld) thenRejectedFor(requirement, reason string) error {
	problems, err := w.problems()
	if err != nil {
		return err
	}
	code := reasonCodes[reason]
	if !slices.ContainsFunc(problems, func(p submission.Problem) bool { return p.Code == code && p.ID == requirement }) {
		return fmt.Errorf("the problems are %v, want %s %s", problems, requirement, code)
	}
	return nil
}

func (w *submissionWorld) thenFileRejected(reason string) error {
	return w.thenRejectedFor("transcript", reason)
}

func (w *submissionWorld) thenNothingStarted() error {
	ctx := context.Background()
	all, err := w.procedures.List(ctx)
	if err != nil {
		return err
	}
	digests, err := w.documents.Digests(ctx)
	if err != nil {
		return err
	}
	if len(all) > 0 || len(digests) > 0 {
		return fmt.Errorf("there are %d procedures and %d documents", len(all), len(digests))
	}
	return nil
}
