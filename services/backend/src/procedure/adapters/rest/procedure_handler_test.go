package rest_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/ports/input"
)

// fakeProcedures opens detail, or answers err.
type fakeProcedures struct {
	detail input.ProcedureDetail
	err    error
}

func (f fakeProcedures) Detail(context.Context, string, string) (input.ProcedureDetail, error) {
	return f.detail, f.err
}

// fakePeople knows the names of ana and of the staff member staff-1.
type fakePeople struct{ authinput.People }

func (fakePeople) Names(_ context.Context, ids []string) (map[string]string, error) {
	known := map[string]string{"ana": "Ana Prueba", "staff-1": "Sofía Ejemplo"}
	names := map[string]string{}
	for _, id := range ids {
		if name, ok := known[id]; ok {
			names[id] = name
		}
	}
	return names, nil
}

func openedDetail(canAct bool) input.ProcedureDetail {
	return input.ProcedureDetail{
		Procedure: domain.Procedure{ID: "P-1", ProcedureType: 7, Version: 2, Applicant: "ana", OpenStep: "documents"},
		Name:      "Equivalencias",
		State:     domain.StateWithApplicant,
		Steps: []detail.Step{
			{
				Step:      definition.Step{ID: "documents", Name: "Adjuntar documentos", Action: definition.ActionUploadDocument, Role: definition.RoleApplicant, Next: "review"},
				Fields:    []definition.FieldRequirement{{Field: definition.Field{ID: "reason", Name: "Motivo", Type: definition.FieldLongText}, Required: false}},
				Documents: []definition.DocumentRequirement{{Document: definition.Document{ID: "transcript", Name: "Historia académica"}, Required: true}},
				Files: []domain.Document{
					{ID: "D-1", Requirement: "transcript", Step: "documents", FileName: "historia.pdf", MediaType: "application/pdf", Size: 3, UploadedBy: "staff-1"},
				},
			},
			{
				Step: definition.Step{ID: "review", Name: "Revisar", Action: definition.ActionDecide, Role: definition.RoleStaff, Area: "office", Outcomes: []definition.Outcome{
					{ID: "accept", Name: "Aceptar", Next: "documents"},
				}},
				Area: &detail.Area{ID: "office", Name: "Dirección de Alumnos"},
			},
		},
		CanAct: canAct,
	}
}

func getProcedure(t *testing.T, procedures fakeProcedures, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, rest.NewProcedureHandler(fakeAuth{}, fakePeople{}, procedures, nil))
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures/P-1", nil)
	r.Header.Set("Cookie", "sigta_session="+cookie)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func procedureDetail(t *testing.T, recorder *httptest.ResponseRecorder) api.ProcedureDetail {
	t.Helper()
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	var got api.ProcedureDetail
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestGetProcedureShowsItsSteps(t *testing.T) {
	got := procedureDetail(t, getProcedure(t, fakeProcedures{detail: openedDetail(true)}, session))

	if got.Id != "P-1" || got.ProcedureType != 7 || got.Version != 2 || got.Name != "Equivalencias" || !got.CanAct {
		t.Errorf("detail = %+v", got)
	}
	if got.State != api.WithApplicant || got.OpenStep == nil || *got.OpenStep != "documents" {
		t.Errorf("state %s, open step %v; want with_applicant on documents", got.State, got.OpenStep)
	}
	if got.Applicant != (api.Person{Id: "ana", Name: "Ana Prueba"}) {
		t.Errorf("applicant = %+v", got.Applicant)
	}
	if len(got.Steps) != 2 {
		t.Fatalf("got %d steps, want 2", len(got.Steps))
	}
	first, second := got.Steps[0], got.Steps[1]
	if first.Id != "documents" || first.Action != api.UploadDocument || first.Role != api.Applicant || first.Area != nil || first.Next == nil || *first.Next != "review" {
		t.Errorf("first step = %+v", first)
	}
	if len(first.Fields) != 1 || !reflect.DeepEqual(first.Fields[0], api.FieldRequirement{Id: "reason", Name: "Motivo", Required: false, Type: api.LongText, Options: []api.NamedOption{}}) {
		t.Errorf("fields of the first step = %+v", first.Fields)
	}
	if len(first.Documents) != 1 || first.Documents[0].Id != "transcript" || !first.Documents[0].Required {
		t.Errorf("documents of the first step = %+v", first.Documents)
	}
	if len(first.Files) != 1 || first.Files[0].Id != "D-1" || first.Files[0].Document != "transcript" || first.Files[0].FileName != "historia.pdf" || !first.Files[0].Current {
		t.Errorf("files of the first step = %+v", first.Files)
	}
	if first.Files[0].UploadedBy != (api.Person{Id: "staff-1", Name: "Sofía Ejemplo"}) {
		t.Errorf("uploaded by = %+v", first.Files[0].UploadedBy)
	}
	if second.Role != api.Staff || second.Area == nil || *second.Area != (api.NamedOption{Id: "office", Name: "Dirección de Alumnos"}) || second.Next != nil {
		t.Errorf("second step = %+v", second)
	}
	if len(second.Outcomes) != 1 || second.Outcomes[0] != (api.StepOutcome{Id: "accept", Name: "Aceptar", Next: "documents"}) {
		t.Errorf("outcomes of the second step = %+v", second.Outcomes)
	}
}

func TestGetProcedureOfAnApplicantWhoIsNotRegistered(t *testing.T) {
	detail := openedDetail(false)
	detail.Procedure.Applicant = "gone"
	got := procedureDetail(t, getProcedure(t, fakeProcedures{detail: detail}, session))

	if got.Applicant != (api.Person{Id: "gone"}) {
		t.Errorf("applicant = %+v, want the id with no name", got.Applicant)
	}
}

func TestGetProcedureThatIsClosed(t *testing.T) {
	detail := openedDetail(false)
	detail.Procedure.OpenStep = ""
	detail.State = domain.StateClosed
	recorder := getProcedure(t, fakeProcedures{detail: detail}, session)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["open_step"]; ok {
		t.Errorf("a closed procedure has open_step %s", raw["open_step"])
	}
	if got := procedureDetail(t, recorder); got.State != api.Closed || got.CanAct {
		t.Errorf("state %s and can act %t, want closed and false", got.State, got.CanAct)
	}
}

func TestGetProcedureAnswersEmptyListsAsArrays(t *testing.T) {
	recorder := getProcedure(t, fakeProcedures{detail: openedDetail(false)}, session)

	var got struct {
		CanAct bool `json:"can_act"`
		Steps  []struct {
			Fields    json.RawMessage `json:"fields"`
			Documents json.RawMessage `json:"documents"`
			Outcomes  json.RawMessage `json:"outcomes"`
			Files     json.RawMessage `json:"files"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.CanAct {
		t.Error("someone with no actions can act on the open step")
	}
	if string(got.Steps[1].Fields) != "[]" || string(got.Steps[0].Outcomes) != "[]" {
		t.Errorf("fields of a step that asks none = %s, outcomes of a step that is not decide = %s, want []", got.Steps[1].Fields, got.Steps[0].Outcomes)
	}
	if string(got.Steps[1].Documents) != "[]" {
		t.Errorf("documents of a step that asks none = %s, want []", got.Steps[1].Documents)
	}
	if string(got.Steps[1].Files) != "[]" {
		t.Errorf("files of a step that asks no document = %s, want []", got.Steps[1].Files)
	}
}

func TestGetProcedureAnswersNotFound(t *testing.T) {
	recorder := getProcedure(t, fakeProcedures{err: domain.ErrProcedureNotFound}, session)

	if recorder.Code != http.StatusNotFound || message(t, recorder) != "No existe el trámite." {
		t.Errorf("status %d: %s", recorder.Code, recorder.Body)
	}
}

func TestGetProcedureAsksToSignIn(t *testing.T) {
	recorder := getProcedure(t, fakeProcedures{detail: openedDetail(true)}, "other")

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status %d: %s", recorder.Code, recorder.Body)
	}
}
