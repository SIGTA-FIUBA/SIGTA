package rest_test

import (
	"backend/src/adapters/dev"
	"backend/src/adapters/filesystem"
	"backend/src/adapters/memory"
	"backend/src/adapters/rest"
	"backend/src/application"
	"backend/src/domain"
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var pdfUpTo1KB = domain.DocumentRule{AcceptedTypes: []string{"application/pdf"}, MaxSize: 1024}

type documentsServer struct {
	url string
}

// newDocumentsServer runs the real handlers over the file system store, with
// procedure p1 started by titular@ and p2 by otro@.
func newDocumentsServer(t *testing.T) documentsServer {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	auth := application.NewAuthService(
		application.AuthConfig{IdleTimeout: time.Hour, LoginLinkTTL: time.Hour, AllowedDomain: "fi.uba.ar", LoginLinkURL: "#token="},
		dev.NewIdentityProvider(srv.URL+"/auth/callback"),
		memory.NewUserRepository(
			domain.User{ID: "u1", Email: "titular@fi.uba.ar", Role: domain.RoleApplicant},
			domain.User{ID: "u2", Email: "operativo@fi.uba.ar", Role: domain.RoleStaffOperative},
			domain.User{ID: "u3", Email: "otro@fi.uba.ar", Role: domain.RoleApplicant},
		),
		memory.NewSessionStore(),
		memory.NewLoginLinkStore(),
		&outbox{},
		time.Now,
	)
	store, err := filesystem.NewDocumentStore(filepath.Join(t.TempDir(), "documents"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	procedures := memory.NewProcedures(
		memory.Procedure{ID: "p1", ApplicantID: "u1", Requirements: map[string]domain.DocumentRule{"transcript": pdfUpTo1KB}},
		memory.Procedure{ID: "p2", ApplicantID: "u3", Requirements: map[string]domain.DocumentRule{"transcript": pdfUpTo1KB}},
	)
	documents := application.NewDocumentService(
		application.DocumentConfig{UnlinkedRule: pdfUpTo1KB},
		procedures, procedures, memory.NewDocumentRepository(), store, time.Now,
	)

	api := rest.NewServer(nil, rest.NewAuthHandler(auth, srv.URL+"/me", false), rest.NewDocumentHandler(auth, documents))
	rest.HandlerFromMux(rest.NewStrictHandler(api, nil), mux)
	return documentsServer{url: srv.URL}
}

// login returns a client with an open session for email, through the dev
// identity provider.
func (s documentsServer) login(t *testing.T, email string) *http.Client {
	t.Helper()
	client := newClient(t)
	res, err := get(t, client, s.url+"/auth/login?login_hint="+url.QueryEscape(email))
	expectStatus(t, res, err, http.StatusOK)
	return client
}

func upload(t *testing.T, client *http.Client, target, filename string, content []byte) (int, io.Reader) {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, target, &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", form.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, bytes.NewReader(resBody)
}

func pdf(size int) []byte {
	return append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), size-9)...)
}

func TestUploadAndDownloadADocument(t *testing.T) {
	s := newDocumentsServer(t)
	client := s.login(t, "titular@fi.uba.ar")

	status, body := upload(t, client, s.url+"/procedures/p1/documents?requirementId=transcript", "analítico.pdf", pdf(500))
	if status != http.StatusCreated {
		t.Fatalf("upload: status %d", status)
	}
	var document rest.Document
	if err := json.NewDecoder(body).Decode(&document); err != nil {
		t.Fatal(err)
	}
	if document.ProcedureId != "p1" || document.RequirementId == nil || *document.RequirementId != "transcript" ||
		document.ContentType != "application/pdf" || document.Size != 500 || document.UploadedBy != "u1" {
		t.Errorf("document = %+v", document)
	}

	res, err := get(t, client, s.url+"/procedures/p1/documents/"+document.Id+"/content")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("download: status %d", res.StatusCode)
	}
	if got, _ := io.ReadAll(res.Body); !bytes.Equal(got, pdf(500)) {
		t.Error("downloaded content differs from the upload")
	}
	if got := res.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	disposition, params, err := mime.ParseMediaType(res.Header.Get("Content-Disposition"))
	if err != nil || disposition != "attachment" || params["filename"] != "analítico.pdf" {
		t.Errorf("Content-Disposition = %q", res.Header.Get("Content-Disposition"))
	}
}

func TestUploadErrors(t *testing.T) {
	s := newDocumentsServer(t)
	titular := s.login(t, "titular@fi.uba.ar")
	png := []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("x", 100))

	tests := []struct {
		name    string
		client  *http.Client
		url     string
		content []byte
		want    int
	}{
		{"wrong type", titular, "/procedures/p1/documents?requirementId=transcript", png, http.StatusUnsupportedMediaType},
		{"too large", titular, "/procedures/p1/documents?requirementId=transcript", pdf(1025), http.StatusRequestEntityTooLarge},
		{"empty", titular, "/procedures/p1/documents?requirementId=transcript", nil, http.StatusBadRequest},
		{"unknown requirement", titular, "/procedures/p1/documents?requirementId=passport", pdf(100), http.StatusBadRequest},
		{"applicant without requirement", titular, "/procedures/p1/documents", pdf(100), http.StatusBadRequest},
		{"another applicant's procedure", titular, "/procedures/p2/documents?requirementId=transcript", pdf(100), http.StatusForbidden},
		{"missing procedure", titular, "/procedures/nope/documents?requirementId=transcript", pdf(100), http.StatusForbidden},
		{"no session", newClient(t), "/procedures/p1/documents?requirementId=transcript", pdf(100), http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := upload(t, tt.client, s.url+tt.url, "archivo.pdf", tt.content)
			if status != tt.want {
				t.Fatalf("status = %d, want %d", status, tt.want)
			}
			var e rest.Error
			if err := json.NewDecoder(body).Decode(&e); err != nil || e.Message == "" {
				t.Errorf("error body = %+v, %v", e, err)
			}
		})
	}
}

func TestStaffAttachADocumentThatCoversNoRequirement(t *testing.T) {
	s := newDocumentsServer(t)

	status, body := upload(t, s.login(t, "operativo@fi.uba.ar"), s.url+"/procedures/p1/documents", "resolucion.pdf", pdf(100))
	if status != http.StatusCreated {
		t.Fatalf("status = %d, want %d", status, http.StatusCreated)
	}
	var document rest.Document
	if err := json.NewDecoder(body).Decode(&document); err != nil {
		t.Fatal(err)
	}
	if document.RequirementId != nil || document.UploadedBy != "u2" {
		t.Errorf("document = %+v", document)
	}
}

func TestUploadWithoutAFilePart(t *testing.T) {
	s := newDocumentsServer(t)
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	_ = form.WriteField("other", "x")
	_ = form.Close()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, s.url+"/procedures/p1/documents?requirementId=transcript", &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", form.FormDataContentType())

	res, err := s.login(t, "titular@fi.uba.ar").Do(req)
	expectStatus(t, res, err, http.StatusBadRequest)
}

func TestDownloadFollowsThePermissionOnTheProcedure(t *testing.T) {
	s := newDocumentsServer(t)
	status, body := upload(t, s.login(t, "titular@fi.uba.ar"), s.url+"/procedures/p1/documents?requirementId=transcript", "a.pdf", pdf(100))
	if status != http.StatusCreated {
		t.Fatalf("upload: status %d", status)
	}
	var document rest.Document
	if err := json.NewDecoder(body).Decode(&document); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		client *http.Client
		path   string
		want   int
	}{
		{"staff", s.login(t, "operativo@fi.uba.ar"), "/procedures/p1/documents/" + document.Id + "/content", http.StatusOK},
		{"another applicant", s.login(t, "otro@fi.uba.ar"), "/procedures/p1/documents/" + document.Id + "/content", http.StatusForbidden},
		{"through another procedure", s.login(t, "otro@fi.uba.ar"), "/procedures/p2/documents/" + document.Id + "/content", http.StatusNotFound},
		{"no session", newClient(t), "/procedures/p1/documents/" + document.Id + "/content", http.StatusUnauthorized},
		{"missing document", s.login(t, "titular@fi.uba.ar"), "/procedures/p1/documents/nope/content", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := get(t, tt.client, s.url+tt.path)
			expectStatus(t, res, err, tt.want)
		})
	}
}
