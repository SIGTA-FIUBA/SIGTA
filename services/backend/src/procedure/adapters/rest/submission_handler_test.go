package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"backend/src/api"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/submission"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

// fakeSubmissions reads every file it is given, as the use case does, and
// answers err or a procedure waiting on openStep.
type fakeSubmissions struct {
	openStep string
	err      error

	user     string
	received input.Submission
	files    map[string]string
}

func (f *fakeSubmissions) Submit(_ context.Context, userID string, s input.Submission) (domain.Procedure, error) {
	f.user, f.received, f.files = userID, s, map[string]string{}
	for file, err := range s.Files {
		if err != nil {
			return domain.Procedure{}, err
		}
		content, err := io.ReadAll(file.Content)
		if err != nil {
			return domain.Procedure{}, err
		}
		f.files[file.Requirement] = file.FileName + ":" + string(content)
	}
	if f.err != nil {
		return domain.Procedure{}, f.err
	}
	return domain.Procedure{ID: "P-9", Applicant: userID, OpenStep: f.openStep}, nil
}

func submitRequest(t *testing.T, parts ...part) *http.Request {
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
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/procedures", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("Cookie", "sigta_session="+session)
	return r
}

func serveSubmission(t *testing.T, submissions *fakeSubmissions, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, rest.NewProcedureHandler(fakeAuth{}, fakePeople{}, nil, submissions))
	recorder := httptest.NewRecorder()
	api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler)).ServeHTTP(recorder, r)
	return recorder
}

func TestSubmitPassesTheValuesAndStreamsTheFiles(t *testing.T) {
	submissions := &fakeSubmissions{openStep: "review"}
	recorder := serveSubmission(t, submissions, submitRequest(t,
		part{"procedure_type", "", "7"},
		part{"values", "", `{"reason":"Cursé afuera","subjects":["Física I","Química"]}`},
		part{"document", "", "transcript"},
		part{"file", "historia.pdf", "%PDF-1"},
		part{"document", "", "program"},
		part{"file", "programa.pdf", "%PDF-2"},
	))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body %s", recorder.Code, recorder.Body)
	}
	var got api.StartedProcedure
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Id != "P-9" || got.State != api.ProcedureState(domain.StateInOffice) || got.OpenStep == nil || *got.OpenStep != "review" {
		t.Errorf("body = %s", recorder.Body)
	}
	if submissions.user != "ana" || submissions.received.ProcedureType != 7 {
		t.Errorf("Submit() got user %q and type %d", submissions.user, submissions.received.ProcedureType)
	}
	values := submissions.received.Values
	if values["reason"].Text != "Cursé afuera" || !slices.Equal(values["subjects"].Items, []string{"Física I", "Química"}) {
		t.Errorf("values = %+v", values)
	}
	if submissions.files["transcript"] != "historia.pdf:%PDF-1" || submissions.files["program"] != "programa.pdf:%PDF-2" {
		t.Errorf("files = %v", submissions.files)
	}
}

func TestSubmitWithoutFiles(t *testing.T) {
	submissions := &fakeSubmissions{}
	recorder := serveSubmission(t, submissions, submitRequest(t, part{"procedure_type", "", "7"}, part{"values", "", `{}`}))
	var got api.StartedProcedure
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusCreated || got.State != api.ProcedureState(domain.StateClosed) || got.OpenStep != nil {
		t.Errorf("status = %d, body %s; want a closed procedure", recorder.Code, recorder.Body)
	}
}

func TestSubmitAnswersTheProblems(t *testing.T) {
	rejected := &submission.Error{Problems: []submission.Problem{
		submission.FieldProblem(submission.CodeMissing, "reason"),
		submission.DocumentProblem(submission.CodeTooLarge, "transcript"),
	}}
	recorder := serveSubmission(t, &fakeSubmissions{err: fmt.Errorf("wrapped: %w", rejected)}, submitRequest(t, part{"procedure_type", "", "7"}, part{"values", "", `{}`}))

	var got api.SubmissionRejected
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := []api.SubmissionProblem{
		{Code: api.Missing, Kind: api.Field, Id: "reason"},
		{Code: api.TooLarge, Kind: api.Document, Id: "transcript"},
	}
	if recorder.Code != http.StatusBadRequest || !slices.Equal(got.Problems, want) {
		t.Errorf("status = %d, body %s", recorder.Code, recorder.Body)
	}
}

func TestSubmitOfATypeThatIsNotPublished(t *testing.T) {
	recorder := serveSubmission(t, &fakeSubmissions{err: output.ErrNoVersion}, submitRequest(t, part{"procedure_type", "", "99"}, part{"values", "", `{}`}))
	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", recorder.Code)
	}
}

func TestSubmitRejectsAMalformedRequest(t *testing.T) {
	tests := map[string][]part{
		"no parts":                   nil,
		"values first":               {{"values", "", `{}`}, {"procedure_type", "", "7"}},
		"a type that is not numeric": {{"procedure_type", "", "siete"}, {"values", "", `{}`}},
		"values that are not JSON":   {{"procedure_type", "", "7"}, {"values", "", `reason=x`}},
		"a number as a value":        {{"procedure_type", "", "7"}, {"values", "", `{"credits":12}`}},
		"a list of numbers":          {{"procedure_type", "", "7"}, {"values", "", `{"subjects":[1]}`}},
		"a file without document":    {{"procedure_type", "", "7"}, {"values", "", `{}`}, {"file", "a.pdf", "%PDF"}},
		"a document without file":    {{"procedure_type", "", "7"}, {"values", "", `{}`}, {"document", "", "transcript"}},
	}
	for name, parts := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := serveSubmission(t, &fakeSubmissions{}, submitRequest(t, parts...))
			var got api.SubmissionRejected
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != http.StatusBadRequest || got.Message != api.InvalidRequestMessage || got.Problems == nil {
				t.Errorf("status = %d, body %s; want 400 with no problems", recorder.Code, recorder.Body)
			}
		})
	}
}

func TestSubmitNeedsASession(t *testing.T) {
	r := submitRequest(t, part{"procedure_type", "", "7"}, part{"values", "", `{}`})
	r.Header.Del("Cookie")
	if recorder := serveSubmission(t, &fakeSubmissions{}, r); recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", recorder.Code)
	}
}

func TestSubmitPastTheMostARequestCarries(t *testing.T) {
	tooLarge := &fakeSubmissions{err: fmt.Errorf("store the file: %w", &http.MaxBytesError{Limit: api.MaxRequestBytes})}
	recorder := serveSubmission(t, tooLarge, submitRequest(t, part{"procedure_type", "", "7"}, part{"values", "", `{}`}))

	var got api.SubmissionRejected
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusBadRequest || !strings.Contains(got.Message, "25 MB") {
		t.Errorf("status = %d, body %s; want 400 naming the limit", recorder.Code, recorder.Body)
	}
}
