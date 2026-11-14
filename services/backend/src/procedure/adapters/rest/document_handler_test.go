package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/src/api"
	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
)

const session = "token"

type fakeAuth struct{ authinput.Authenticator }

func (fakeAuth) Authenticate(_ context.Context, token string) (authdomain.User, error) {
	if token != session {
		return authdomain.User{}, authdomain.ErrSessionNotFound
	}
	return authdomain.User{ID: "ana"}, nil
}

// fakeDocuments answers err, and otherwise what it is given, and records the
// upload it receives.
type fakeDocuments struct {
	attached []domain.Document
	listed   []input.AttachedDocument
	err      error

	upload  input.Upload
	content []byte
}

func (f *fakeDocuments) Attach(_ context.Context, _ string, u input.Upload) (domain.Document, error) {
	f.upload = u
	content, err := io.ReadAll(u.Content)
	if err != nil {
		return domain.Document{}, err
	}
	f.content = content
	if f.err != nil {
		return domain.Document{}, f.err
	}
	return domain.Document{ID: "D-1", Requirement: u.Requirement, FileName: u.FileName, MediaType: "application/pdf", Size: int64(len(content))}, nil
}

func (f *fakeDocuments) List(context.Context, string, string) ([]input.AttachedDocument, error) {
	return f.listed, f.err
}

func (f *fakeDocuments) Open(context.Context, string, string, string) (domain.Document, io.ReadCloser, error) {
	if f.err != nil {
		return domain.Document{}, nil, f.err
	}
	return f.attached[0], io.NopCloser(strings.NewReader("%PDF-1.7")), nil
}

func serve(t *testing.T, documents *fakeDocuments, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, rest.NewDocumentHandler(fakeAuth{}, documents))
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r.Header.Set("Cookie", "sigta_session="+session)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

type part struct{ name, fileName, content string }

func attachRequest(t *testing.T, parts ...part) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, p := range parts {
		var w io.Writer
		var err error
		if p.fileName != "" {
			w, err = writer.CreateFormFile(p.name, p.fileName)
		} else {
			w, err = writer.CreateFormField(p.name)
		}
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, p.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/procedures/P-1/documents", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r
}

func message(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var e api.Error
	if err := json.Unmarshal(recorder.Body.Bytes(), &e); err != nil {
		t.Fatalf("body %s: %v", recorder.Body, err)
	}
	return e.Message
}

func TestAttachStreamsTheFileToTheUseCase(t *testing.T) {
	documents := &fakeDocuments{}
	recorder := serve(t, documents, attachRequest(t, part{"document", "", "transcript"}, part{"file", "historia.pdf", "%PDF-1.7 contenido"}))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s; want 201", recorder.Code, recorder.Body)
	}
	if documents.upload.ProcedureID != "P-1" || documents.upload.Requirement != "transcript" || documents.upload.FileName != "historia.pdf" {
		t.Errorf("upload = %+v", documents.upload)
	}
	if string(documents.content) != "%PDF-1.7 contenido" {
		t.Errorf("content = %q", documents.content)
	}
	var got api.ProcedureDocument
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Id != "D-1" || got.Document != "transcript" || !got.Current {
		t.Errorf("body = %s", recorder.Body)
	}
}

func TestAttachNeedsTheDocumentBeforeTheFile(t *testing.T) {
	tests := map[string][]part{
		"file first":  {{"file", "historia.pdf", "%PDF"}, {"document", "", "transcript"}},
		"no file":     {{"document", "", "transcript"}},
		"other field": {{"document", "", "transcript"}, {"attachment", "historia.pdf", "%PDF"}},
		"long id":     {{"document", "", strings.Repeat("a", 300)}, {"file", "historia.pdf", "%PDF"}},
	}
	for name, parts := range tests {
		t.Run(name, func(t *testing.T) {
			documents := &fakeDocuments{}
			recorder := serve(t, documents, attachRequest(t, parts...))
			if recorder.Code != http.StatusBadRequest || message(t, recorder) != api.InvalidRequestMessage {
				t.Errorf("status = %d, body %s; want 400 with the invalid request message", recorder.Code, recorder.Body)
			}
			if documents.upload.Content != nil {
				t.Error("the use case was called")
			}
		})
	}
}

func TestAttachAnswersABrokenUploadAsABadRequest(t *testing.T) {
	r := attachRequest(t, part{"document", "", "transcript"}, part{"file", "historia.pdf", "%PDF-1.7 contenido"})
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	r.Body = io.NopCloser(bytes.NewReader(body[:len(body)-20]))

	recorder := serve(t, &fakeDocuments{}, r)
	if recorder.Code != http.StatusBadRequest || message(t, recorder) != api.InvalidRequestMessage {
		t.Errorf("status = %d, body %s; want 400 with the invalid request message", recorder.Code, recorder.Body)
	}
}

func TestAttachStopsAtTheRequestLimit(t *testing.T) {
	file := strings.Repeat("a", api.MaxRequestBytes)
	recorder := serve(t, &fakeDocuments{}, attachRequest(t, part{"document", "", "transcript"}, part{"file", "historia.pdf", file}))
	if recorder.Code != http.StatusBadRequest || !strings.Contains(message(t, recorder), "tamaño máximo") {
		t.Errorf("status = %d, body %s; want 400 saying it is too large", recorder.Code, recorder.Body)
	}
}

func TestAttachAnswersEachRejection(t *testing.T) {
	tests := []struct {
		err    error
		status int
		says   string
	}{
		{domain.ErrDocumentType, http.StatusBadRequest, "tipo"},
		{domain.ErrDocumentTooLarge, http.StatusBadRequest, "tamaño máximo"},
		{domain.ErrEmptyDocument, http.StatusBadRequest, "vacío"},
		{domain.ErrTooManyDocuments, http.StatusBadRequest, "cantidad máxima"},
		{domain.ErrDocumentNotRequested, http.StatusBadRequest, "no pide"},
		{domain.ErrAccessDenied, http.StatusForbidden, "No podés adjuntar"},
		{domain.ErrProcedureNotFound, http.StatusNotFound, "No existe el trámite"},
		{errors.New("disk full"), http.StatusInternalServerError, "inesperado"},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			recorder := serve(t, &fakeDocuments{err: tt.err}, attachRequest(t, part{"document", "", "transcript"}, part{"file", "historia.pdf", "%PDF"}))
			if recorder.Code != tt.status || !strings.Contains(message(t, recorder), tt.says) {
				t.Errorf("status = %d, body %s; want %d saying %q", recorder.Code, recorder.Body, tt.status, tt.says)
			}
		})
	}
}

func TestDocumentsNeedASession(t *testing.T) {
	requests := map[string]*http.Request{
		"list":     httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures/P-1/documents", http.NoBody),
		"attach":   attachRequest(t, part{"document", "", "transcript"}, part{"file", "historia.pdf", "%PDF"}),
		"download": httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures/P-1/documents/D-1/content", http.NoBody),
	}
	for name, r := range requests {
		t.Run(name, func(t *testing.T) {
			server := api.NewServer(nil, nil, nil, rest.NewDocumentHandler(fakeAuth{}, &fakeDocuments{}))
			recorder := httptest.NewRecorder()
			api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler)).ServeHTTP(recorder, r)
			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", recorder.Code)
			}
		})
	}
}

func TestListSendsWhichDocumentsAreCurrent(t *testing.T) {
	uploaded := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	documents := &fakeDocuments{listed: []input.AttachedDocument{
		{Document: domain.Document{ID: "D-1", Requirement: "transcript", UploadedAt: uploaded}},
		{Document: domain.Document{ID: "D-2", Requirement: "transcript", UploadedAt: uploaded}, Current: true},
	}}
	recorder := serve(t, documents, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures/P-1/documents", http.NoBody))

	var got []api.ProcedureDocument
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Current || !got[1].Current {
		t.Errorf("body = %s; want D-2 as the current one", recorder.Body)
	}
}

func TestListIsEmptyAndNotNull(t *testing.T) {
	recorder := serve(t, &fakeDocuments{}, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures/P-1/documents", http.NoBody))
	if strings.TrimSpace(recorder.Body.String()) != "[]" {
		t.Errorf("body = %s, want []", recorder.Body)
	}
}

func TestDownloadIsAnAttachmentWithTheDetectedType(t *testing.T) {
	documents := &fakeDocuments{attached: []domain.Document{{ID: "D-1", FileName: "historia académica.pdf", MediaType: "application/pdf", Size: 8}}}
	recorder := serve(t, documents, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures/P-1/documents/D-1/content", http.NoBody))

	if recorder.Code != http.StatusOK || recorder.Body.String() != "%PDF-1.7" {
		t.Fatalf("status = %d, body %q", recorder.Code, recorder.Body)
	}
	headers := recorder.Header()
	want := map[string]string{
		"Content-Type":            "application/pdf",
		"Content-Length":          "8",
		"X-Content-Type-Options":  "nosniff",
		"Content-Security-Policy": "sandbox",
		"Cache-Control":           "no-store",
	}
	for name, value := range want {
		if got := headers.Get(name); got != value {
			t.Errorf("%s = %q, want %q", name, got, value)
		}
	}
	if got := headers.Get("Content-Disposition"); got != "attachment; filename*=utf-8''historia%20acad%C3%A9mica.pdf" {
		t.Errorf("Content-Disposition = %s", got)
	}
}

func TestDownloadDoesNotRevealOtherProcedures(t *testing.T) {
	for _, err := range []error{domain.ErrProcedureNotFound, domain.ErrDocumentNotFound} {
		recorder := serve(t, &fakeDocuments{err: err}, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures/P-1/documents/D-1/content", http.NoBody))
		if recorder.Code != http.StatusNotFound {
			t.Errorf("%v: status = %d, want 404", err, recorder.Code)
		}
	}
}
