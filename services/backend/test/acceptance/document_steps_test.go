package acceptance

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var pdfOfSize = regexp.MustCompile(`^a PDF of (\d+) bytes$`)

// documentWorld wires the services as the server does, with the files on a
// disk adapter in a temporary directory.
type documentWorld struct {
	areas       *authapp.AreaService
	memberships *authapp.MembershipService
	people      *authmem.UserRepository
	types       *application.ProcedureTypeService
	procedures  movableProcedures
	documents   *memory.DocumentRepository
	service     *application.DocumentService
	dir         string

	known     map[string]bool
	procedure domain.Procedure
	nextID    int

	attached domain.Document
	content  []byte
	// contents keeps what was attached under each file name.
	contents   map[string][]byte
	downloaded []byte
	file       domain.Document
	err        error
}

// registerDocumentSteps returns its world so that the detail of a procedure
// can share the procedure type, the staff and the procedure it sets up.
func registerDocumentSteps(sc *godog.ScenarioContext) *documentWorld {
	w := &documentWorld{known: map[string]bool{}, contents: map[string][]byte{}}
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return ctx, w.setUp()
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		if w.dir == "" {
			return ctx, err
		}
		return ctx, errors.Join(err, os.RemoveAll(w.dir))
	})

	sc.Given(`^a procedure type whose steps ask for these documents:$`, w.givenProcedureType)
	sc.Given(`^these staff members:$`, w.givenStaff)
	sc.Given(`^"([^"]*)" has a procedure waiting on the "([^"]*)" step$`, w.givenProcedure)
	sc.Given(`^the procedure moved on to the "([^"]*)" step$`, w.givenMovedOn)
	sc.Given(`^"([^"]*)" attached (.+) named "([^"]*)" as the "([^"]*)"$`, w.givenAttached)
	sc.Given(`^"([^"]*)" is in the "([^"]*)" with "([^"]*)" only for another procedure type$`, w.givenScopedStaff)

	sc.When(`^"([^"]*)" attaches (.+) named "([^"]*)" as the "([^"]*)"$`, w.attach)
	sc.When(`^"([^"]*)" downloads the "([^"]*)"$`, w.download)
	sc.When(`^"([^"]*)" downloads "([^"]*)"$`, w.downloadFile)

	sc.Then(`^the "([^"]*)" of the procedure is "([^"]*)", uploaded by "([^"]*)" on the "([^"]*)" step$`, w.thenCurrent)
	sc.Then(`^the "([^"]*)" of the procedure has the type "([^"]*)"$`, w.thenType)
	sc.Then(`^the file is rejected because (.+)$`, w.thenRejected)
	sc.Then(`^the procedure has no attached documents$`, func() error { return w.thenAttachedCount(0) })
	sc.Then(`^the procedure has (\d+) attached documents$`, w.thenAttachedCount)
	sc.Then(`^they get the same content that was attached$`, w.thenSameContent)
	sc.Then(`^they get "([^"]*)" with the content that was attached$`, w.thenFileContent)
	sc.Then(`^the file comes as "([^"]*)" with the type "([^"]*)"$`, w.thenFileType)
	sc.Then(`^the procedure is not found$`, func() error { return w.thenFails(domain.ErrProcedureNotFound) })
	sc.Then(`^the document store keeps the file under the SHA-256 digest of its content$`, w.thenStoredByDigest)
	sc.Then(`^the record of the document names that digest and its size$`, w.thenRecordNamesDigest)
	return w
}

func (w *documentWorld) setUp() error {
	dir, err := os.MkdirTemp("", "sigta-documents-")
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
	// The authors of the procedure type in the Secretariat, so that it can be
	// published.
	stored := authmem.NewMembershipRepository(
		authdomain.Membership{ID: "M-proposer", UserID: defaultAuthors.ProposedBy, Area: authinput.SecretariatArea, Permissions: []string{authinput.EditProcedureTypes}},
		authdomain.Membership{ID: "M-approver", UserID: defaultAuthors.ApprovedBy, Area: authinput.SecretariatArea, Permissions: []string{authinput.PublishProcedureTypes}},
	)
	// The Secretariat comes with SIGTA: it is not created again.
	w.known[authinput.SecretariatArea] = true
	w.areas = authapp.NewAreaService(areas, administrators, stored, time.Now)
	w.people = authmem.NewUserRepository()
	w.memberships = authapp.NewMembershipService(areas, administrators, stored, w.people, w.newID, time.Now)
	w.types = application.NewProcedureTypeService(memory.NewProcedureTypeVersionRepository(), w.areas, w.memberships, time.Now)
	recorded := memory.NewProcedureRepository()
	w.procedures = movableProcedures{ProcedureRepository: recorded, openStep: map[string]string{}, entered: map[string][]domain.FieldValue{}}
	w.documents = memory.NewDocumentRepository(recorded)
	procedures := application.NewProcedureService(w.types, w.procedures, w.documents, w.memberships, w.newID, time.Now)
	w.service = application.NewDocumentService(procedures, w.documents, files, w.newID, time.Now)
	return nil
}

func (w *documentWorld) newID() string {
	w.nextID++
	return "D-" + strconv.Itoa(w.nextID)
}

// ensure creates the area, with the given ID, and the permission the first
// time they are named.
func ensure(areas *authapp.AreaService, known map[string]bool, id, area, permission string) error {
	ctx := context.Background()
	if !known[area] {
		if _, err := areas.Create(ctx, scenarioAdministrator, authdomain.Area{ID: id, Name: area}); err != nil {
			return fmt.Errorf("area %s: %w", area, err)
		}
		known[area] = true
	}
	if key := area + "/" + permission; !known[key] {
		if _, err := areas.CreatePermission(ctx, scenarioAdministrator, id, authdomain.Permission{ID: permission, Name: permission}); err != nil {
			return fmt.Errorf("permission %s of %s: %w", permission, area, err)
		}
		known[key] = true
	}
	return nil
}

// givenProcedureType reads "the applicant" or an area and a quoted permission
// as who does each step, and a row with no document as a step that asks for
// the text fields of its "fields" column, all optional.
func (w *documentWorld) givenProcedureType(table *godog.Table) error {
	var def definition.ProcedureTypeDefinition
	def.Name = "Trámite con documentos"
	def.Description = "Un trámite para las pruebas de documentos"
	for _, row := range tableRows(table) {
		stepID, doneBy, documentID := row["step"], row["done by"], row["document"]
		step := definition.Step{ID: stepID, Name: stepID}
		if documentID == "" {
			step.Action, step.OptionalFields = definition.ActionEnterData, list(row["fields"])
			for _, id := range step.OptionalFields {
				def.Fields = append(def.Fields, definition.Field{ID: id, Name: id, Type: definition.FieldText})
			}
		} else {
			maxSize, err := strconv.ParseInt(row["max size"], 10, 64)
			if err != nil {
				return err
			}
			def.Documents = append(def.Documents, definition.Document{
				ID: documentID, Name: documentID, AcceptedTypes: list(row["types"]), MaxSizeBytes: maxSize,
			})
			step.Action, step.Documents = definition.ActionUploadDocument, []string{documentID}
		}
		if doneBy == "the applicant" {
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
		if n := len(def.Steps); n > 0 {
			def.Steps[n-1].Next = step.ID
		}
		def.Steps = append(def.Steps, step)
	}
	_, err := w.types.Publish(context.Background(), input.Publication{
		ProcedureType: 1,
		Definition:    def,
		Authorship:    version.Authorship{ProposedBy: "proponente", ApprovedBy: "aprobador"},
	})
	return err
}

func (w *documentWorld) givenStaff(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		user, area, permissions := row.Cells[0].Value, row.Cells[1].Value, list(row.Cells[2].Value)
		for _, permission := range permissions {
			if err := ensure(w.areas, w.known, idOf(area), area, permission); err != nil {
				return err
			}
		}
		w.people.Add(signedIn(user))
		_, err := w.memberships.Grant(context.Background(), scenarioAdministrator, authinput.MembershipGrant{
			UserID: user, Area: idOf(area), Permissions: permissions,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *documentWorld) givenScopedStaff(user, area, permission string) error {
	if err := ensure(w.areas, w.known, idOf(area), area, permission); err != nil {
		return err
	}
	w.people.Add(signedIn(user))
	_, err := w.memberships.Grant(context.Background(), scenarioAdministrator, authinput.MembershipGrant{
		UserID: user, Area: idOf(area), Permissions: []string{permission}, Scope: []int64{2},
	})
	return err
}

func (w *documentWorld) givenProcedure(applicant, step string) error {
	published, err := w.types.Version(context.Background(), 1, 1)
	if err != nil {
		return err
	}
	w.procedure = domain.Procedure{ID: "P-1", ProcedureType: 1, Version: 1, VersionDigest: published.Digest, Applicant: applicant, OpenStep: step}
	return w.procedures.Add(context.Background(), output.Start{Procedure: w.procedure})
}

func (w *documentWorld) givenMovedOn(step string) error {
	return moveTo(w.procedures.ProcedureRepository, w.procedure.ID, step)
}

// moveTo records that the open step of the procedure was confirmed and the
// procedure moved to step, without the checks of a confirmation, which are
// another story.
func moveTo(procedures *memory.ProcedureRepository, procedureID, step string) error {
	p, err := procedures.Get(context.Background(), procedureID)
	if err != nil {
		return err
	}
	moved := p
	moved.OpenStep = step
	return procedures.Advance(context.Background(), moved, domain.Event{
		ProcedureID: p.ID,
		Step:        p.OpenStep,
		Action:      definition.ActionConfirmExternalFact,
		By:          "staff@fi.uba.ar",
		At:          time.Now(),
		Next:        step,
	})
}

// movableProcedures shows a procedure on another step without its advance,
// and adds values to it after its start, which are other stories.
type movableProcedures struct {
	*memory.ProcedureRepository
	openStep map[string]string
	entered  map[string][]domain.FieldValue
}

func (m movableProcedures) Values(ctx context.Context, procedureID string) ([]domain.FieldValue, error) {
	values, err := m.ProcedureRepository.Values(ctx, procedureID)
	return append(values, m.entered[procedureID]...), err
}

func (m movableProcedures) Get(ctx context.Context, id string) (domain.Procedure, error) {
	p, err := m.ProcedureRepository.Get(ctx, id)
	if step, ok := m.openStep[id]; ok && err == nil {
		p.OpenStep = step
	}
	return p, err
}

func (m movableProcedures) List(ctx context.Context) ([]domain.Procedure, error) {
	all, err := m.ProcedureRepository.List(ctx)
	for i, p := range all {
		if step, ok := m.openStep[p.ID]; ok {
			all[i].OpenStep = step
		}
	}
	return all, err
}

func (w *documentWorld) givenAttached(user, file, name, requirement string) error {
	w.attach(user, file, name, requirement)
	return w.err
}

func fileContent(file string) ([]byte, error) {
	if m := pdfOfSize.FindStringSubmatch(file); m != nil {
		size, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, err
		}
		header := "%PDF-1.7\n"
		return append([]byte(header), bytes.Repeat([]byte("x"), size-len(header))...), nil
	}
	switch file {
	case "a text file":
		return []byte("una historia académica en texto plano\n"), nil
	case "a ZIP file":
		return append([]byte("PK\x03\x04"), bytes.Repeat([]byte{0}, 60)...), nil
	case "a PNG image":
		return append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 60)...), nil
	case "an empty file":
		return []byte{}, nil
	}
	return nil, fmt.Errorf("unknown file %q", file)
}

func (w *documentWorld) attach(user, file, name, requirement string) {
	content, err := fileContent(file)
	if err != nil {
		w.err = err
		return
	}
	w.content = content
	w.attached, w.err = w.service.Attach(context.Background(), user, input.Upload{
		ProcedureID: w.procedure.ID,
		Requirement: requirement,
		FileName:    name,
		Content:     bytes.NewReader(content),
	})
	if w.err == nil {
		w.contents[w.attached.FileName] = content
	}
}

func (w *documentWorld) download(user, requirement string) error {
	ctx := context.Background()
	attached, err := w.documents.List(ctx, w.procedure.ID)
	if err != nil {
		return err
	}
	var id string
	for _, d := range domain.Current(attached) {
		if d.Requirement == requirement {
			id = d.ID
		}
	}
	_, file, err := w.service.Open(ctx, user, w.procedure.ID, id)
	w.err = err
	if err != nil {
		return nil
	}
	defer func() { _ = file.Close() }()
	w.downloaded, err = io.ReadAll(file)
	return err
}

// downloadFile downloads the file attached under that name, current or not.
func (w *documentWorld) downloadFile(user, name string) error {
	ctx := context.Background()
	attached, err := w.documents.List(ctx, w.procedure.ID)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(attached, func(d domain.Document) bool { return d.FileName == name })
	if i < 0 {
		return fmt.Errorf("no file is named %s", name)
	}
	d, file, err := w.service.Open(ctx, user, w.procedure.ID, attached[i].ID)
	w.file, w.err = d, err
	if err != nil {
		return nil
	}
	defer func() { _ = file.Close() }()
	w.downloaded, err = io.ReadAll(file)
	return err
}

func (w *documentWorld) current(requirement string) (domain.Document, error) {
	attached, err := w.documents.List(context.Background(), w.procedure.ID)
	if err != nil {
		return domain.Document{}, err
	}
	for _, d := range domain.Current(attached) {
		if d.Requirement == requirement {
			return d, nil
		}
	}
	return domain.Document{}, fmt.Errorf("the procedure has no %s", requirement)
}

func (w *documentWorld) thenCurrent(requirement, name, user, step string) error {
	if w.err != nil {
		return w.err
	}
	d, err := w.current(requirement)
	if err != nil {
		return err
	}
	if d.FileName != name || d.UploadedBy != user || d.Step != step || d.Procedure != w.procedure.ID {
		return fmt.Errorf("the %s is %+v", requirement, d)
	}
	return nil
}

func (w *documentWorld) thenType(requirement, mediaType string) error {
	if w.err != nil {
		return w.err
	}
	d, err := w.current(requirement)
	if err != nil {
		return err
	}
	if d.MediaType != mediaType {
		return fmt.Errorf("the type is %s, want %s", d.MediaType, mediaType)
	}
	return nil
}

var rejections = map[string]error{
	"its type is not accepted":          domain.ErrDocumentTypeNotAccepted,
	"it is too large":                   domain.ErrDocumentTooLarge,
	"it is empty":                       domain.ErrDocumentEmpty,
	"the open step does not ask for it": domain.ErrDocumentNotRequested,
	"they cannot act on that step":      domain.ErrAccessDenied,
	"the procedure is not found":        domain.ErrProcedureNotFound,
}

func (w *documentWorld) thenRejected(reason string) error {
	want, ok := rejections[reason]
	if !ok {
		return fmt.Errorf("unknown reason %q", reason)
	}
	return w.thenFails(want)
}

func (w *documentWorld) thenFails(want error) error {
	if !errors.Is(w.err, want) {
		return fmt.Errorf("want %w, got %s", want, fmt.Sprint(w.err))
	}
	return nil
}

func (w *documentWorld) thenAttachedCount(n int) error {
	attached, err := w.documents.List(context.Background(), w.procedure.ID)
	if err != nil {
		return err
	}
	if len(attached) != n {
		return fmt.Errorf("%d attached documents, want %d", len(attached), n)
	}
	return nil
}

func (w *documentWorld) thenSameContent() error {
	if w.err != nil {
		return w.err
	}
	if !bytes.Equal(w.downloaded, w.content) {
		return fmt.Errorf("got %d bytes, want the %d that were attached", len(w.downloaded), len(w.content))
	}
	return nil
}

func (w *documentWorld) thenFileContent(name string) error {
	if w.err != nil {
		return w.err
	}
	if w.file.FileName != name {
		return fmt.Errorf("got %s, want %s", w.file.FileName, name)
	}
	if !bytes.Equal(w.downloaded, w.contents[name]) {
		return fmt.Errorf("got %d bytes, want the %d attached as %s", len(w.downloaded), len(w.contents[name]), name)
	}
	return nil
}

func (w *documentWorld) thenFileType(name, mediaType string) error {
	if w.err != nil {
		return w.err
	}
	if w.file.FileName != name || w.file.MediaType != mediaType {
		return fmt.Errorf("got %s with the type %s, want %s with %s", w.file.FileName, w.file.MediaType, name, mediaType)
	}
	return nil
}

func (w *documentWorld) thenStoredByDigest() error {
	if w.err != nil {
		return w.err
	}
	sum := sha256.Sum256(w.content)
	stored, err := os.ReadFile(filepath.Join(w.dir, hex.EncodeToString(sum[:])))
	if err != nil {
		return err
	}
	if !bytes.Equal(stored, w.content) {
		return errors.New("the stored file is not the attached content")
	}
	return nil
}

func (w *documentWorld) thenRecordNamesDigest() error {
	sum := sha256.Sum256(w.content)
	if want := "sha256:" + hex.EncodeToString(sum[:]); w.attached.Digest != want || w.attached.Size != int64(len(w.content)) {
		return fmt.Errorf("the record names %s and %d bytes, want %s and %d", w.attached.Digest, w.attached.Size, want, len(w.content))
	}
	return nil
}
