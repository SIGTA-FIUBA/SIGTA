package rest_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/src/api"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/catalog"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/ports/input"
)

// fakeProcedures opens view, or answers err.
type fakeProcedures struct {
	input.Procedures
	view input.ProcedureView
	err  error
}

func (f fakeProcedures) Open(context.Context, string, string) (input.ProcedureView, error) {
	return f.view, f.err
}

func openedView(actions ...string) input.ProcedureView {
	return input.ProcedureView{
		Procedure: domain.Procedure{ID: "P-1", ProcedureType: 7, Version: 2, Applicant: "ana", OpenStep: "documents"},
		Name:      "Equivalencias",
		Steps: []detail.Step{
			{
				Step:      definition.Step{ID: "documents", Name: "Adjuntar documentos", Action: definition.ActionUploadDocument, Role: definition.RoleApplicant, Next: "review"},
				Documents: []catalog.DocumentRequirement{{Document: definition.Document{ID: "transcript", Name: "Historia académica"}, Required: true}},
			},
			{
				Step:     definition.Step{ID: "review", Name: "Revisar", Action: definition.ActionConfirmExternalFact, Role: definition.RoleStaff, Area: "office"},
				AreaName: "Dirección de Alumnos",
			},
		},
		Actions: actions,
	}
}

func getProcedure(t *testing.T, procedures fakeProcedures, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, rest.NewProcedureHandler(fakeAuth{}, procedures))
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures/P-1", nil)
	r.Header.Set("Cookie", "sigta_session="+cookie)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func TestGetProcedureShowsItsSteps(t *testing.T) {
	recorder := getProcedure(t, fakeProcedures{view: openedView("documents")}, session)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	var got api.ProcedureDetail
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Id != "P-1" || got.ProcedureType != 7 || got.Version != 2 || got.Name != "Equivalencias" || got.OpenStep != "documents" || !got.CanAct {
		t.Errorf("detail = %+v", got)
	}
	if len(got.Steps) != 2 {
		t.Fatalf("got %d steps, want 2", len(got.Steps))
	}
	first, second := got.Steps[0], got.Steps[1]
	if first.Id != "documents" || first.Action != api.UploadDocument || first.Role != api.Applicant || first.AreaName != "" {
		t.Errorf("first step = %+v", first)
	}
	if len(first.Documents) != 1 || first.Documents[0].Id != "transcript" || !first.Documents[0].Required {
		t.Errorf("documents of the first step = %+v", first.Documents)
	}
	if second.Role != api.Staff || second.AreaName != "Dirección de Alumnos" {
		t.Errorf("second step = %+v", second)
	}
}

func TestGetProcedureAnswersEmptyListsAsArrays(t *testing.T) {
	recorder := getProcedure(t, fakeProcedures{view: openedView()}, session)

	var got struct {
		CanAct bool `json:"can_act"`
		Steps  []struct {
			Documents json.RawMessage `json:"documents"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.CanAct {
		t.Error("someone with no actions can act on the open step")
	}
	if string(got.Steps[1].Documents) != "[]" {
		t.Errorf("documents of a step that asks none = %s, want []", got.Steps[1].Documents)
	}
}

func TestGetProcedureAnswersNotFound(t *testing.T) {
	recorder := getProcedure(t, fakeProcedures{err: domain.ErrProcedureNotFound}, session)

	if recorder.Code != http.StatusNotFound || message(t, recorder) != "No existe el trámite." {
		t.Errorf("status %d: %s", recorder.Code, recorder.Body)
	}
}

func TestGetProcedureAsksToSignIn(t *testing.T) {
	recorder := getProcedure(t, fakeProcedures{view: openedView("documents")}, "other")

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status %d: %s", recorder.Code, recorder.Body)
	}
}
