package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

const (
	applicant = "titular@fi.uba.ar"
	reviewer  = "revisora@fi.uba.ar"
	maxSize   = 1024
	// documentID is the identifier of the first document attached.
	documentID = "D-1"
)

var attachedAt = time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)

func documentsDefinition() definition.ProcedureTypeDefinition {
	return definition.ProcedureTypeDefinition{
		Name:        "Trámite con documentos",
		Description: "Un trámite para las pruebas",
		Documents: []definition.Document{
			{ID: "statement", Name: "Constancia", AcceptedTypes: []string{"application/pdf", "image/png"}, MaxSizeBytes: maxSize},
			{ID: "annex", Name: "Anexo", AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: maxSize},
			{ID: "resolution", Name: "Resolución", AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: maxSize},
		},
		Steps: []definition.Step{
			{ID: "documents", Name: "Documentación", Role: definition.RoleApplicant, Action: definition.ActionUploadDocument, Documents: []string{"statement"}, OptionalDocuments: []string{"annex"}, Next: "review"},
			{ID: "review", Name: "Revisión", Role: definition.RoleStaff, Area: office.ID, Permissions: []string{"review"}, Action: definition.ActionUploadDocument, Documents: []string{"resolution"}},
		},
	}
}

// memoryFiles keeps the files in memory, under the digest the disk adapter
// would give them.
type memoryFiles struct{ files map[string][]byte }

func (m *memoryFiles) Put(_ context.Context, content io.Reader) (string, int64, error) {
	data, err := io.ReadAll(content)
	if err != nil {
		return "", 0, err
	}
	sum := sha256.Sum256(data)
	digest := "sha256:" + hex.EncodeToString(sum[:])
	m.files[digest] = data
	return digest, int64(len(data)), nil
}

func (m *memoryFiles) Open(_ context.Context, digest string) (io.ReadCloser, error) {
	data, ok := m.files[digest]
	if !ok {
		return nil, output.ErrNoFile
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *memoryFiles) Digests(context.Context) ([]string, error) {
	return slices.Collect(maps.Keys(m.files)), nil
}

func (m *memoryFiles) Remove(_ context.Context, digest string) error {
	delete(m.files, digest)
	return nil
}

type documentsFixture struct {
	service    *DocumentService
	procedures *memory.ProcedureRepository
	documents  *memory.DocumentRepository
	files      *memoryFiles
	// digest is the one of the version that the procedures follow.
	digest string
}

func newDocumentsFixture(t *testing.T) documentsFixture {
	t.Helper()
	ctx := t.Context()
	areas := authmem.NewAreaRepository(office)
	administrators := authmem.NewAdministrators("admin")
	stored := authmem.NewMembershipRepository()
	memberships := authapp.NewMembershipService(areas, administrators, stored, func() string { return "M-1" }, time.Now)
	if _, err := memberships.Grant(ctx, "admin", authinput.MembershipGrant{UserID: reviewer, Area: office.ID, Permissions: []string{"review"}}); err != nil {
		t.Fatal(err)
	}
	types := newService(memory.NewProcedureVersions(), office)
	published, err := types.Publish(ctx, publication(1, 0, documentsDefinition()))
	if err != nil {
		t.Fatal(err)
	}
	f := documentsFixture{
		digest:     published.Digest,
		procedures: memory.NewProcedureRepository(),
		documents:  memory.NewDocumentRepository(),
		files:      &memoryFiles{files: map[string][]byte{}},
	}
	procedures := NewProcedureService(types, f.procedures, memberships, func() string { return "P-new" })
	attached := 0
	newID := func() string {
		attached++
		return fmt.Sprintf("D-%d", attached)
	}
	f.service = NewDocumentService(procedures, f.documents, f.files, newID, func() time.Time { return attachedAt })
	f.on(t, "P-1", "documents")
	return f
}

func (f documentsFixture) on(t *testing.T, procedureID, step string) {
	t.Helper()
	p := domain.Procedure{ID: procedureID, ProcedureType: 1, Version: 1, VersionDigest: f.digest, Applicant: applicant, OpenStep: step}
	if err := f.procedures.Add(t.Context(), p); err != nil {
		t.Fatal(err)
	}
}

func pdf(size int) []byte {
	return append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), size-len("%PDF-1.7\n"))...)
}

func upload(requirement, name string, content []byte) input.Upload {
	return input.Upload{ProcedureID: "P-1", Requirement: requirement, FileName: name, Content: bytes.NewReader(content)}
}

func TestAttachStoresTheFileAndRecordsIt(t *testing.T) {
	f := newDocumentsFixture(t)
	content := pdf(100)

	d, err := f.service.Attach(t.Context(), applicant, upload("statement", "constancia.pdf", content))
	if err != nil {
		t.Fatal(err)
	}

	want := domain.Document{
		ID: documentID, Procedure: "P-1", Requirement: "statement", Step: "documents",
		FileName: "constancia.pdf", MediaType: "application/pdf", Size: 100, Digest: d.Digest,
		UploadedBy: applicant, UploadedAt: attachedAt,
	}
	if d != want {
		t.Errorf("document = %+v, want %+v", d, want)
	}
	if !bytes.Equal(f.files.files[d.Digest], content) {
		t.Error("the store does not hold the content under the recorded digest")
	}
	recorded, err := f.documents.List(t.Context(), "P-1")
	if err != nil || len(recorded) != 1 || recorded[0] != want {
		t.Errorf("recorded = %+v, %v; want the document", recorded, err)
	}
}

func TestAttachAcceptsAnOptionalDocument(t *testing.T) {
	f := newDocumentsFixture(t)
	if _, err := f.service.Attach(t.Context(), applicant, upload("annex", "anexo.pdf", pdf(100))); err != nil {
		t.Fatal(err)
	}
}

func TestAttachChecksTheSizeAtTheLimit(t *testing.T) {
	f := newDocumentsFixture(t)
	if _, err := f.service.Attach(t.Context(), applicant, upload("statement", "a.pdf", pdf(maxSize))); err != nil {
		t.Fatalf("at the maximum size: %v", err)
	}
	_, err := f.service.Attach(t.Context(), applicant, upload("statement", "b.pdf", pdf(maxSize+1)))
	if !errors.Is(err, domain.ErrDocumentTooLarge) {
		t.Fatalf("one byte over: err = %v, want ErrDocumentTooLarge", err)
	}
	if recorded, _ := f.documents.List(t.Context(), "P-1"); len(recorded) != 1 {
		t.Errorf("recorded %d documents, want only the first", len(recorded))
	}
}

func TestAttachRejects(t *testing.T) {
	tests := []struct {
		name        string
		user        string
		requirement string
		fileName    string
		content     []byte
		want        error
	}{
		{"text named as a PDF", applicant, "statement", "constancia.pdf", []byte("no soy un PDF"), domain.ErrDocumentType},
		{"accepted elsewhere but not here", applicant, "annex", "anexo.png", []byte("\x89PNG\r\n\x1a\n0000"), domain.ErrDocumentType},
		{"empty file", applicant, "statement", "vacio.pdf", nil, domain.ErrEmptyDocument},
		{"a document of another step", applicant, "resolution", "resolucion.pdf", pdf(100), domain.ErrDocumentNotRequested},
		{"a document the type does not have", applicant, "passport", "pasaporte.pdf", pdf(100), domain.ErrDocumentNotRequested},
		{"staff on the applicant's step", reviewer, "statement", "constancia.pdf", pdf(100), domain.ErrAccessDenied},
		{"another applicant", "otra@fi.uba.ar", "statement", "constancia.pdf", pdf(100), domain.ErrProcedureNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newDocumentsFixture(t)
			_, err := f.service.Attach(t.Context(), tt.user, upload(tt.requirement, tt.fileName, tt.content))
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if len(f.files.files) != 0 {
				t.Error("the store kept a rejected file")
			}
		})
	}
}

func TestStaffAttachOnTheirStep(t *testing.T) {
	f := newDocumentsFixture(t)
	f.on(t, "P-2", "review")
	u := upload("resolution", "resolucion.pdf", pdf(100))
	u.ProcedureID = "P-2"
	if _, err := f.service.Attach(t.Context(), reviewer, u); err != nil {
		t.Fatal(err)
	}
	u.Content = bytes.NewReader(pdf(100))
	if _, err := f.service.Attach(t.Context(), applicant, u); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("applicant on the staff step: err = %v, want ErrAccessDenied", err)
	}
}

func TestOpenReturnsTheStoredContent(t *testing.T) {
	f := newDocumentsFixture(t)
	content := pdf(100)
	if _, err := f.service.Attach(t.Context(), applicant, upload("statement", "constancia.pdf", content)); err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{applicant, reviewer} {
		_, file, err := f.service.Open(t.Context(), user, "P-1", documentID)
		if err != nil {
			t.Fatalf("%s: %v", user, err)
		}
		got, err := io.ReadAll(file)
		if err != nil || !bytes.Equal(got, content) {
			t.Errorf("%s reads %d bytes, %v; want the attached content", user, len(got), err)
		}
	}
}

func TestOpenFindsTheDocumentOnlyInItsProcedure(t *testing.T) {
	f := newDocumentsFixture(t)
	if _, err := f.service.Attach(t.Context(), applicant, upload("statement", "constancia.pdf", pdf(100))); err != nil {
		t.Fatal(err)
	}
	f.on(t, "P-2", "documents")
	if _, _, err := f.service.Open(t.Context(), applicant, "P-2", documentID); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Errorf("err = %v, want ErrDocumentNotFound", err)
	}
	if _, _, err := f.service.Open(t.Context(), "otra@fi.uba.ar", "P-1", documentID); !errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("another applicant: err = %v, want ErrProcedureNotFound", err)
	}
}

func TestListMarksTheCurrentDocuments(t *testing.T) {
	f := newDocumentsFixture(t)
	for _, name := range []string{"primera.pdf", "segunda.pdf"} {
		if _, err := f.service.Attach(t.Context(), applicant, upload("statement", name, pdf(100))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.service.Attach(t.Context(), applicant, upload("annex", "anexo.pdf", pdf(100))); err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{applicant, reviewer} {
		attached, err := f.service.List(t.Context(), user, "P-1")
		if err != nil {
			t.Fatalf("%s: %v", user, err)
		}
		var got []string
		for _, d := range attached {
			got = append(got, fmt.Sprintf("%s %t", d.FileName, d.Current))
		}
		want := []string{"primera.pdf false", "segunda.pdf true", "anexo.pdf true"}
		if !slices.Equal(got, want) {
			t.Errorf("%s: List = %v, want %v", user, got, want)
		}
	}
}

func TestListAnswersNotFoundToWhoCannotSeeTheProcedure(t *testing.T) {
	f := newDocumentsFixture(t)
	for _, procedureID := range []string{"P-1", "P-missing"} {
		if _, err := f.service.List(t.Context(), "otra@fi.uba.ar", procedureID); !errors.Is(err, domain.ErrProcedureNotFound) {
			t.Errorf("%s: err = %v, want ErrProcedureNotFound", procedureID, err)
		}
	}
}

func TestWithExtension(t *testing.T) {
	tests := []struct{ name, mediaType, want string }{
		{"constancia.pdf", "application/pdf", "constancia.pdf"},
		{"CONSTANCIA.PDF", "application/pdf", "CONSTANCIA.PDF"},
		{"constancia.html", "application/pdf", "constancia.html.pdf"},
		{"constancia", "image/png", "constancia.png"},
		{"datos.bin", "application/x-unknown-to-mime", "datos.bin"},
	}
	for _, tt := range tests {
		if got := withExtension(tt.name, tt.mediaType); got != tt.want {
			t.Errorf("withExtension(%q, %q) = %q, want %q", tt.name, tt.mediaType, got, tt.want)
		}
	}
}

func TestCleanFileName(t *testing.T) {
	long := strings.Repeat("a", 300) + ".pdf"
	tests := []struct{ name, want string }{
		{"constancia.pdf", "constancia.pdf"},
		{"../../etc/passwd", "passwd"},
		{`C:\Users\ana\historia académica.pdf`, "historia académica.pdf"},
		{"con\r\nsalto.pdf", "consalto.pdf"},
		{"inválido\xff.pdf", "inválido.pdf"},
		{"constancia\u202efdp.exe", "constanciafdp.exe"},
		{"", defaultFileName},
		{"..", defaultFileName},
		{"  ", defaultFileName},
		{long, long[len(long)-maxFileName:]},
	}
	for _, tt := range tests {
		if got := cleanFileName(tt.name); got != tt.want {
			t.Errorf("cleanFileName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestRemoveUnreferencedFiles keeps the file of an attached document and
// removes one that no document names, as an upload whose record failed
// leaves.
func TestRemoveUnreferencedFiles(t *testing.T) {
	f := newDocumentsFixture(t)
	d, err := f.service.Attach(t.Context(), applicant, upload("statement", "constancia.pdf", pdf(100)))
	if err != nil {
		t.Fatal(err)
	}
	orphan, _, err := f.files.Put(t.Context(), bytes.NewReader(pdf(200)))
	if err != nil {
		t.Fatal(err)
	}
	removed, err := f.service.RemoveUnreferencedFiles(t.Context())
	if err != nil || removed != 1 {
		t.Fatalf("RemoveUnreferencedFiles() = %d, %v; want 1, nil", removed, err)
	}
	if _, ok := f.files.files[orphan]; ok {
		t.Error("the file that no document names is still stored")
	}
	if _, ok := f.files.files[d.Digest]; !ok {
		t.Error("the file of the attached document was removed")
	}
}

func TestAttachStopsAtTheMaximumOfFilesPerDocument(t *testing.T) {
	f := newDocumentsFixture(t)
	for i := range maxFilesPerDocument {
		if _, err := f.service.Attach(t.Context(), applicant, upload("statement", "constancia.pdf", pdf(100+i))); err != nil {
			t.Fatalf("file %d: %v", i+1, err)
		}
	}
	stored := len(f.files.files)
	_, err := f.service.Attach(t.Context(), applicant, upload("statement", "constancia.pdf", pdf(50)))
	if !errors.Is(err, domain.ErrTooManyDocuments) {
		t.Fatalf("one more file: err = %v, want ErrTooManyDocuments", err)
	}
	if len(f.files.files) != stored {
		t.Error("the store kept the file past the maximum")
	}
	if _, err := f.service.Attach(t.Context(), applicant, upload("annex", "anexo.pdf", pdf(100))); err != nil {
		t.Errorf("another document of the same procedure: %v", err)
	}
}
