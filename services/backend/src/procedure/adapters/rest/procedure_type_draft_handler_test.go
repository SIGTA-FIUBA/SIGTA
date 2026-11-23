package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/draft"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

// fakeDrafts answers err, and otherwise view or published, and records what
// it is given.
type fakeDrafts struct {
	view      input.DraftView
	published version.ProcedureTypeVersion
	editable  input.EditableProcedureTypes
	err       error

	user       string
	definition definition.ProcedureTypeDefinition
	revision   int
	reason     string
}

func (f *fakeDrafts) Create(_ context.Context, user string, d definition.ProcedureTypeDefinition) (input.DraftView, error) {
	f.user, f.definition = user, d
	return f.view, f.err
}

func (f *fakeDrafts) Open(_ context.Context, user string, _ int64) (input.DraftView, error) {
	f.user = user
	return f.view, f.err
}

func (f *fakeDrafts) Get(_ context.Context, user string, _ int64) (input.DraftView, error) {
	f.user = user
	return f.view, f.err
}

func (f *fakeDrafts) List(_ context.Context, user string) ([]input.DraftView, error) {
	f.user = user
	return []input.DraftView{f.view}, f.err
}

func (f *fakeDrafts) Editable(_ context.Context, user string) (input.EditableProcedureTypes, error) {
	f.user = user
	return f.editable, f.err
}

func (f *fakeDrafts) Edit(_ context.Context, user string, _ int64, revision int, d definition.ProcedureTypeDefinition) (input.DraftView, error) {
	f.user, f.revision, f.definition = user, revision, d
	return f.view, f.err
}

func (f *fakeDrafts) Propose(_ context.Context, user string, _ int64, revision int) (input.DraftView, error) {
	f.user, f.revision = user, revision
	return f.view, f.err
}

func (f *fakeDrafts) Reject(_ context.Context, user string, _ int64, revision int, reason string) (input.DraftView, error) {
	f.user, f.revision, f.reason = user, revision, reason
	return f.view, f.err
}

func (f *fakeDrafts) Approve(_ context.Context, user string, _ int64, revision int) (version.ProcedureTypeVersion, error) {
	f.user, f.revision = user, revision
	return f.published, f.err
}

func (f *fakeDrafts) Discard(_ context.Context, user string, _ int64, revision int) error {
	f.user, f.revision = user, revision
	return f.err
}

var editedAt = time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)

func proposedView() input.DraftView {
	return input.DraftView{
		Draft: draft.Draft{
			ProcedureTypeID: 7,
			Base:            2,
			Definition: definition.ProcedureTypeDefinition{
				Name:      "Equivalencias",
				Documents: []definition.Document{{ID: "transcript", Name: "Historia académica", AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: 1024, Resolution: true}},
				Steps:     []definition.Step{{ID: "documents", Name: "Adjuntar", Role: definition.RoleApplicant, Action: definition.ActionUploadDocument, Documents: []string{"transcript"}}},
			},
			Status:     draft.StatusProposed,
			Revision:   4,
			EditedBy:   "5RnF8i6-BEyF",
			EditedAt:   editedAt,
			Editors:    []string{"3Las2Qqyn", "5RnF8i6-BEyF"},
			ProposedBy: "5RnF8i6-BEyF",
			ProposedAt: editedAt,
		},
	}
}

func callDrafts(t *testing.T, drafts *fakeDrafts, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, nil, nil, rest.NewProcedureTypeDraftHandler(fakeAuth{}, fakePeople{}, drafts))
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Cookie", "sigta_session="+session)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func TestListEditableProcedureTypesSaysWhatThePersonCanDo(t *testing.T) {
	proposed := proposedView()
	proposed.CanReview = true
	drafts := &fakeDrafts{editable: input.EditableProcedureTypes{
		CanCreate: true,
		ProcedureTypes: []input.EditableProcedureType{
			{ProcedureTypeID: 1, Name: "Equivalencia", Latest: 3, PublishedAt: editedAt, CanOpenDraft: true},
			{ProcedureTypeID: 7, Name: "Equivalencias", Draft: &proposed},
		},
		Areas: []definition.Area{{ID: "commission", Name: "Comisión Curricular", Permissions: []definition.Permission{{ID: "evaluate", Name: "Dictaminar"}}}},
	}}

	recorder := callDrafts(t, drafts, http.MethodGet, "/editable-procedure-types", "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	var got api.EditableProcedureTypes
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if drafts.user != "ana" || !got.CanCreate || len(got.ProcedureTypes) != 2 || len(got.Areas) != 1 {
		t.Fatalf("answer = %+v for %q", got, drafts.user)
	}
	published, fresh := got.ProcedureTypes[0], got.ProcedureTypes[1]
	if published.Version != 3 || published.PublishedAt == nil || !published.PublishedAt.Equal(editedAt) || !published.CanOpenDraft || published.Draft != nil {
		t.Errorf("published type = %+v", published)
	}
	if fresh.Version != 0 || fresh.PublishedAt != nil || fresh.CanOpenDraft || fresh.Draft == nil || !fresh.Draft.CanReview || fresh.Draft.CanEdit {
		t.Errorf("type with only a draft = %+v", fresh)
	}
	if area := got.Areas[0]; area.Id != "commission" || len(area.Permissions) != 1 || area.Permissions[0].Name != "Dictaminar" {
		t.Errorf("area = %+v", area)
	}
}

func TestListEditableProcedureTypesWithoutPermissionsIsEmpty(t *testing.T) {
	recorder := callDrafts(t, &fakeDrafts{}, http.MethodGet, "/editable-procedure-types", "")

	if recorder.Code != http.StatusOK || recorder.Body.String() != `{"areas":[],"can_create":false,"procedure_types":[]}`+"\n" {
		t.Errorf("status %d: %s, want empty lists", recorder.Code, recorder.Body)
	}
}

func TestGetProcedureTypeDraftShowsTheDraft(t *testing.T) {
	recorder := callDrafts(t, &fakeDrafts{view: proposedView()}, http.MethodGet, "/procedure-types/7/draft", "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	var got api.ProcedureTypeDraft
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ProcedureTypeId != 7 || got.Base != 2 || got.Revision != 4 || got.Status != "proposed" || got.EditedBy != (api.Person{Id: "5RnF8i6-BEyF", Name: "Ana Prueba"}) || !got.EditedAt.Equal(editedAt) ||
		!slices.Equal(got.Editors, []api.Person{{Id: "3Las2Qqyn", Name: "Beto Ejemplo"}, {Id: "5RnF8i6-BEyF", Name: "Ana Prueba"}}) {
		t.Errorf("draft = %+v", got)
	}
	// The identifiers are opaque, as those of Google accounts: the screen shows the names.
	if got.ProposedBy == nil || got.ProposedBy.Name != "Ana Prueba" || got.Rejection != nil {
		t.Errorf("proposed by %v, rejection %v", got.ProposedBy, got.Rejection)
	}
	documents := *got.Definition.Documents
	if got.Definition.Name != "Equivalencias" || len(documents) != 1 || *documents[0].MaxSize != 1024 || documents[0].Resolution == nil || !*documents[0].Resolution || got.Definition.Steps[0].Documents == nil {
		t.Errorf("definition = %+v", got.Definition)
	}
	if !strings.Contains(recorder.Body.String(), `"problems":[]`) {
		t.Errorf("a draft without problems answers %s, want an empty list", recorder.Body)
	}
}

func TestDraftProblemsSayWhereTheyAre(t *testing.T) {
	view := proposedView()
	view.Problems = []definition.ValidationProblem{{
		Code:  definition.CodeUnknownArea,
		Path:  []definition.Element{{Kind: definition.KindStep, ID: "review"}},
		Value: "office",
	}}
	recorder := callDrafts(t, &fakeDrafts{view: view}, http.MethodGet, "/procedure-types/7/draft", "")

	var got api.ProcedureTypeDraft
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Problems) != 1 {
		t.Fatalf("problems = %+v", got.Problems)
	}
	p := got.Problems[0]
	if p.Code != "unknown_area" || *p.Value != "office" || p.Other != nil || len(p.Path) != 1 || p.Path[0].Kind != "step" || *p.Path[0].Id != "review" || p.Path[0].Position != nil {
		t.Errorf("problem = %+v", p)
	}
	if p.Text != `step "review": unknown area "office"` {
		t.Errorf("text = %q", p.Text)
	}
}

func TestEditProcedureTypeDraftPassesTheDefinition(t *testing.T) {
	drafts := &fakeDrafts{view: proposedView()}
	body := `{"revision": 3, "definition": {"name": "Equivalencias", "documents": [{"id": "transcript", "name": "Historia", "max_size": 2048, "resolution": true}], "steps": [{"id": "documents", "name": "Adjuntar", "role": "applicant", "action": "upload_document", "documents": ["transcript"]}]}}`
	recorder := callDrafts(t, drafts, http.MethodPut, "/procedure-types/7/draft", body)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	d := drafts.definition
	if drafts.user != "ana" || drafts.revision != 3 || d.Name != "Equivalencias" || len(d.Documents) != 1 || d.Documents[0].MaxSizeBytes != 2048 || !d.Documents[0].Resolution || d.Steps[0].Role != definition.RoleApplicant {
		t.Errorf("edited by %s at revision %d with %+v", drafts.user, drafts.revision, d)
	}
}

func TestApproveProcedureTypeDraftAnswersTheVersion(t *testing.T) {
	drafts := &fakeDrafts{published: version.ProcedureTypeVersion{
		ProcedureTypeID: 7,
		Number:          3,
		Digest:          "sha256:abc",
		Authorship:      version.Authorship{ProposedBy: "ana", ApprovedBy: "beto"},
		PublishedAt:     editedAt,
	}}
	recorder := callDrafts(t, drafts, http.MethodPost, "/procedure-types/7/draft/approval", `{"revision": 4}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	var got api.PublishedVersion
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := api.PublishedVersion{ProcedureTypeId: 7, Number: 3, Digest: "sha256:abc", ProposedBy: "ana", ApprovedBy: "beto", PublishedAt: editedAt}
	if got != want || drafts.revision != 4 {
		t.Errorf("approved at revision %d: %+v, want %+v", drafts.revision, got, want)
	}
}

func TestRejectProcedureTypeDraftPassesTheReason(t *testing.T) {
	drafts := &fakeDrafts{view: proposedView()}
	recorder := callDrafts(t, drafts, http.MethodPost, "/procedure-types/7/draft/rejection", `{"revision": 4, "reason": "Falta el plazo"}`)

	if recorder.Code != http.StatusOK || drafts.reason != "Falta el plazo" || drafts.revision != 4 {
		t.Errorf("status %d, reason %q, revision %d", recorder.Code, drafts.reason, drafts.revision)
	}
}

func TestDraftErrorsAnswerTheirStatus(t *testing.T) {
	invalid := &definition.ValidationError{Cause: definition.ErrInvalidDefinition, Problems: []definition.ValidationProblem{{Code: definition.CodeNoSteps, Path: []definition.Element{{Kind: definition.KindDefinition}}}}}
	calls := map[string]struct{ method, path, body string }{
		"create":  {http.MethodPost, "/procedure-types", `{"definition": {"name": "A", "steps": []}}`},
		"get":     {http.MethodGet, "/procedure-types/7/draft", ""},
		"open":    {http.MethodPost, "/procedure-types/7/draft", ""},
		"edit":    {http.MethodPut, "/procedure-types/7/draft", `{"revision": 1, "definition": {"name": "A", "steps": []}}`},
		"discard": {http.MethodDelete, "/procedure-types/7/draft?revision=1", ""},
		"propose": {http.MethodPost, "/procedure-types/7/draft/proposal", `{"revision": 1}`},
		"reject":  {http.MethodPost, "/procedure-types/7/draft/rejection", `{"revision": 1, "reason": ""}`},
		"approve": {http.MethodPost, "/procedure-types/7/draft/approval", `{"revision": 1}`},
	}
	cases := []struct {
		call   string
		err    error
		status int
	}{
		{"create", domain.ErrAccessDenied, http.StatusForbidden},
		{"get", domain.ErrAccessDenied, http.StatusForbidden},
		{"get", output.ErrDraftNotFound, http.StatusNotFound},
		{"open", output.ErrVersionNotFound, http.StatusNotFound},
		{"open", output.ErrDraftExists, http.StatusConflict},
		{"edit", output.ErrDraftChanged, http.StatusConflict},
		{"edit", draft.ErrDraftProposed, http.StatusConflict},
		{"discard", draft.ErrDraftProposed, http.StatusConflict},
		{"discard", output.ErrDraftNotFound, http.StatusNotFound},
		{"propose", invalid, http.StatusUnprocessableEntity},
		{"reject", draft.ErrRejectionReasonMissing, http.StatusBadRequest},
		{"reject", draft.ErrSelfApproval, http.StatusForbidden},
		{"approve", draft.ErrDraftNotProposed, http.StatusConflict},
		{"approve", &input.OutdatedBaseError{Base: 1, Latest: 2}, http.StatusConflict},
		{"approve", invalid, http.StatusUnprocessableEntity},
		{"approve", errors.New("disk full"), http.StatusInternalServerError},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s %v", c.call, c.err), func(t *testing.T) {
			call := calls[c.call]
			recorder := callDrafts(t, &fakeDrafts{err: fmt.Errorf("wrapped: %w", c.err)}, call.method, call.path, call.body)
			if recorder.Code != c.status {
				t.Errorf("status %d, want %d: %s", recorder.Code, c.status, recorder.Body)
			}
		})
	}
}

func TestInvalidDefinitionListsItsProblems(t *testing.T) {
	invalid := &definition.ValidationError{Cause: definition.ErrInvalidDefinition, Problems: []definition.ValidationProblem{{Code: definition.CodeNoSteps, Path: []definition.Element{{Kind: definition.KindDefinition}}}}}
	recorder := callDrafts(t, &fakeDrafts{err: invalid}, http.MethodPost, "/procedure-types/7/draft/proposal", `{"revision": 1}`)

	var got api.InvalidDefinition
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Message == "" || len(got.Problems) != 1 || got.Problems[0].Code != "no_steps" {
		t.Errorf("answer = %+v", got)
	}
}

func TestDraftsAskToSignIn(t *testing.T) {
	server := api.NewServer(nil, nil, nil, nil, nil, nil, rest.NewProcedureTypeDraftHandler(fakeAuth{}, fakePeople{}, &fakeDrafts{}))
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedure-type-drafts", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", recorder.Code)
	}
}
