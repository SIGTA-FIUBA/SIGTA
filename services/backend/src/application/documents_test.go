package application_test

import (
	"backend/src/adapters/memory"
	"backend/src/application"
	"backend/src/domain"
	"backend/src/ports/input"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

var (
	otherApplicant = domain.User{ID: "other", Role: domain.RoleApplicant}
	approver       = domain.User{ID: "approver", Role: domain.RoleStaffApprover}

	uploadedAt = time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	pdfUpTo1KB = domain.DocumentRule{AcceptedTypes: []string{"application/pdf"}, MaxSize: 1024}
)

// Synthetic files: only their first bytes matter for type detection.
func pdf(size int) []byte {
	return append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), size-9)...)
}

var png = []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("x", 100))

type documentsFixture struct {
	service *application.DocumentService
	store   *memory.DocumentStore
	repo    *memory.DocumentRepository
}

func newDocuments(t *testing.T) documentsFixture {
	t.Helper()
	procedures := memory.NewProcedures(
		memory.Procedure{ID: "p1", ApplicantID: applicant.ID, Requirements: map[string]domain.DocumentRule{"transcript": pdfUpTo1KB}},
		memory.Procedure{ID: "p2", ApplicantID: otherApplicant.ID, Requirements: map[string]domain.DocumentRule{"transcript": pdfUpTo1KB}},
	)
	f := documentsFixture{store: memory.NewDocumentStore(), repo: memory.NewDocumentRepository()}
	f.service = application.NewDocumentService(
		application.DocumentConfig{UnlinkedRule: pdfUpTo1KB},
		procedures, procedures, f.repo, f.store,
		func() time.Time { return uploadedAt },
	)
	return f
}

func transcript(content []byte) input.Upload {
	return input.Upload{ProcedureID: "p1", RequirementID: "transcript", Filename: "analitico.pdf", Content: bytes.NewReader(content)}
}

func TestUploadStoresADocumentThatMeetsTheRequirement(t *testing.T) {
	f := newDocuments(t)

	document, err := f.service.Upload(t.Context(), applicant, transcript(pdf(1024)))
	if err != nil {
		t.Fatal(err)
	}

	want := domain.Document{
		ID:            document.ID,
		ProcedureID:   "p1",
		RequirementID: "transcript",
		Filename:      "analitico.pdf",
		ContentType:   "application/pdf",
		Size:          1024,
		UploadedBy:    applicant.ID,
		UploadedAt:    uploadedAt,
	}
	if document != want {
		t.Errorf("document = %+v, want %+v", document, want)
	}
	if saved, err := f.repo.FindByID(t.Context(), document.ID); err != nil || saved != want {
		t.Errorf("saved = %+v, %v; want %+v", saved, err, want)
	}
	content, err := f.store.Open(t.Context(), document.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := io.ReadAll(content); !bytes.Equal(got, pdf(1024)) {
		t.Error("stored content differs from the upload")
	}
}

func TestUploadRejectsFilesTheRequirementDoesNotAccept(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    error
	}{
		{"wrong type", png, domain.ErrFileTypeNotAccepted},
		{"one byte too large", pdf(1025), domain.ErrFileTooLarge},
		{"empty", nil, domain.ErrEmptyFile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newDocuments(t)

			_, err := f.service.Upload(t.Context(), applicant, transcript(tt.content))

			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if f.store.Len() != 0 {
				t.Error("a rejected file was left in the store")
			}
		})
	}
}

func TestRejectedFileCarriesTheRuleItBroke(t *testing.T) {
	f := newDocuments(t)

	_, err := f.service.Upload(t.Context(), applicant, transcript(png))

	var rejected *domain.FileRejectedError
	if !errors.As(err, &rejected) || rejected.Rule.MaxSize != pdfUpTo1KB.MaxSize {
		t.Fatalf("err = %v, want the rule of the requirement", err)
	}
}

func TestUploadDetectsTheTypeFromTheContentNotTheName(t *testing.T) {
	f := newDocuments(t)
	upload := transcript(png)
	upload.Filename = "analitico.pdf"

	if _, err := f.service.Upload(t.Context(), applicant, upload); !errors.Is(err, domain.ErrFileTypeNotAccepted) {
		t.Fatalf("err = %v, want %v", err, domain.ErrFileTypeNotAccepted)
	}
}

func TestApplicantMustCoverAnExistingRequirement(t *testing.T) {
	tests := []struct {
		name        string
		requirement string
		want        error
	}{
		{"unknown requirement", "passport", domain.ErrUnknownRequirement},
		{"no requirement", "", domain.ErrRequirementRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newDocuments(t)
			upload := transcript(pdf(100))
			upload.RequirementID = tt.requirement

			if _, err := f.service.Upload(t.Context(), applicant, upload); !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestStaffAttachDocumentsThatCoverNoRequirement(t *testing.T) {
	for _, staff := range []domain.User{operative, approver} {
		t.Run(string(staff.Role), func(t *testing.T) {
			f := newDocuments(t)
			upload := transcript(pdf(100))
			upload.RequirementID = ""

			document, err := f.service.Upload(t.Context(), staff, upload)
			if err != nil {
				t.Fatal(err)
			}
			if document.ProcedureID != "p1" || document.RequirementID != "" || document.UploadedBy != staff.ID {
				t.Errorf("document = %+v", document)
			}
		})
	}
}

func TestUploadNeedsAccessToTheProcedure(t *testing.T) {
	for _, procedureID := range []string{"p2", "missing"} {
		t.Run(procedureID, func(t *testing.T) {
			f := newDocuments(t)
			upload := transcript(pdf(100))
			upload.ProcedureID = procedureID

			if _, err := f.service.Upload(t.Context(), applicant, upload); !errors.Is(err, domain.ErrNoProcedureAccess) {
				t.Fatalf("err = %v, want %v", err, domain.ErrNoProcedureAccess)
			}
		})
	}
}

func TestOpenFollowsThePermissionOnTheProcedure(t *testing.T) {
	f := newDocuments(t)
	document, err := f.service.Upload(t.Context(), applicant, transcript(pdf(100)))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		user domain.User
		want error
	}{
		{"its applicant", applicant, nil},
		{"staff", operative, nil},
		{"another applicant", otherApplicant, domain.ErrNoProcedureAccess},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, content, err := f.service.Open(t.Context(), tt.user, "p1", document.ID)
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if content != nil {
				_ = content.Close()
			}
		})
	}
}

func TestDocumentIsOnlyReachableThroughItsProcedure(t *testing.T) {
	f := newDocuments(t)
	document, err := f.service.Upload(t.Context(), applicant, transcript(pdf(100)))
	if err != nil {
		t.Fatal(err)
	}

	// The other applicant can access p2, but the document belongs to p1.
	if _, _, err := f.service.Open(t.Context(), otherApplicant, "p2", document.ID); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("err = %v, want %v", err, domain.ErrDocumentNotFound)
	}
}

func TestUploadKeepsOnlyTheBaseNameOfTheFile(t *testing.T) {
	tests := map[string]string{
		"../../etc/passwd":            "passwd",
		`C:\Users\titular\nota.pdf`:   "nota.pdf",
		"\tnota\x00\n.pdf":            "nota.pdf",
		"":                            "documento",
		strings.Repeat("á", 200) + "": strings.Repeat("á", 127),
	}
	for name, want := range tests {
		f := newDocuments(t)
		upload := transcript(pdf(100))
		upload.Filename = name

		document, err := f.service.Upload(t.Context(), applicant, upload)
		if err != nil {
			t.Fatal(err)
		}
		if document.Filename != want {
			t.Errorf("Filename(%q) = %q, want %q", name, document.Filename, want)
		}
	}
}
