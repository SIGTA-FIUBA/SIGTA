package rest_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

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
	mine   []input.MyProcedure
	err    error
}

func (f fakeProcedures) Detail(context.Context, string, string) (input.ProcedureDetail, error) {
	return f.detail, f.err
}

func (f fakeProcedures) Mine(context.Context, string) ([]input.MyProcedure, error) {
	return f.mine, f.err
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
					{ID: "send_back", Name: "Devolver", Next: "documents", SendBack: true},
				}},
				Area: &detail.Area{ID: "office", Name: "Dirección de Alumnos"},
			},
		},
		CanAct: canAct,
	}
}

func getProcedure(t *testing.T, procedures fakeProcedures, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, rest.NewProcedureHandler(fakeAuth{}, fakePeople{}, procedures, nil, nil), nil, nil)
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
	if first.Files[0].UploadedBy == nil || *first.Files[0].UploadedBy != (api.Person{Id: "staff-1", Name: "Sofía Ejemplo"}) {
		t.Errorf("uploaded by = %+v", first.Files[0].UploadedBy)
	}
	if second.Role != api.Staff || second.Area == nil || *second.Area != (api.NamedOption{Id: "office", Name: "Dirección de Alumnos"}) || second.Next != nil {
		t.Errorf("second step = %+v", second)
	}
	wantOutcomes := []api.StepOutcome{{Id: "accept", Name: "Aceptar", Next: "documents"}, {Id: "send_back", Name: "Devolver", Next: "documents", SendBack: true}}
	if !reflect.DeepEqual(second.Outcomes, wantOutcomes) {
		t.Errorf("outcomes of the second step = %+v", second.Outcomes)
	}
}

func TestGetProcedureShowsTheValuesEntered(t *testing.T) {
	at := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	d := openedDetail(false)
	d.Steps[0].Fields = append(d.Steps[0].Fields, definition.FieldRequirement{Field: definition.Field{ID: "subjects", Name: "Asignaturas", Type: definition.FieldSubjectList}})
	d.Steps[0].Values = []domain.FieldValue{
		{Procedure: "P-1", Field: "reason", Step: "documents", Text: "Pase de carrera", EnteredBy: "ana", EnteredAt: at},
		{Procedure: "P-1", Field: "subjects", Step: "documents", Items: []string{"Análisis II", "Física I"}, EnteredBy: "ana", EnteredAt: at},
	}

	got := procedureDetail(t, getProcedure(t, fakeProcedures{detail: d}, session))

	text, items := "Pase de carrera", []string{"Análisis II", "Física I"}
	ana := api.Person{Id: "ana", Name: "Ana Prueba"}
	want := []api.ProcedureValue{
		{Field: "reason", Step: "documents", Text: &text, EnteredBy: ana, EnteredAt: at},
		{Field: "subjects", Step: "documents", Items: &items, EnteredBy: ana, EnteredAt: at},
	}
	if !reflect.DeepEqual(got.Steps[0].Values, want) {
		t.Errorf("values = %+v, want %+v", got.Steps[0].Values, want)
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
		CanAct             bool            `json:"can_act"`
		CorrectionRequests json.RawMessage `json:"correction_requests"`
		History            json.RawMessage `json:"history"`
		Steps              []struct {
			Fields    json.RawMessage `json:"fields"`
			Documents json.RawMessage `json:"documents"`
			Outcomes  json.RawMessage `json:"outcomes"`
			Values    json.RawMessage `json:"values"`
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
	if string(got.Steps[1].Values) != "[]" {
		t.Errorf("values of a step that asks no field = %s, want []", got.Steps[1].Values)
	}
	if string(got.Steps[1].Files) != "[]" {
		t.Errorf("files of a step that asks no document = %s, want []", got.Steps[1].Files)
	}
	if string(got.History) != "[]" {
		t.Errorf("history of a procedure with no steps done = %s, want []", got.History)
	}
	if string(got.CorrectionRequests) != "[]" {
		t.Errorf("correction requests of a procedure with none pending = %s, want []", got.CorrectionRequests)
	}
}

func TestGetProcedureShowsThePendingCorrectionRequests(t *testing.T) {
	at := time.Date(2026, 11, 17, 9, 0, 0, 0, time.UTC)
	d := openedDetail(true)
	d.Procedure.OpenStep = "review"
	d.CorrectionRequests = []domain.CorrectionRequest{
		{Step: "review", By: "staff-1", At: at, Reason: "faltan los códigos", Fields: []string{"subjects"}},
		{Step: "review", By: "staff-1", At: at.Add(time.Hour), Reason: "falta la firma", Documents: []string{"statement"}},
	}

	got := procedureDetail(t, getProcedure(t, fakeProcedures{detail: d}, session))

	staff := api.Person{Id: "staff-1", Name: "Sofía Ejemplo"}
	want := []api.CorrectionRequest{
		{Step: "review", By: &staff, At: at, Reason: "faltan los códigos", Fields: []string{"subjects"}, Documents: []string{}},
		{Step: "review", By: &staff, At: at.Add(time.Hour), Reason: "falta la firma", Fields: []string{}, Documents: []string{"statement"}},
	}
	if !reflect.DeepEqual(got.CorrectionRequests, want) {
		t.Errorf("correction requests = %+v, want %+v", got.CorrectionRequests, want)
	}
}

func TestGetProcedureShowsItsHistoryInOrder(t *testing.T) {
	at := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	d := openedDetail(false)
	d.History = []domain.Event{
		{ProcedureID: "P-1", Step: "documents", Action: definition.ActionUploadDocument, By: "ana", At: at, Next: "review"},
		{ProcedureID: "P-1", Step: "review", Action: definition.ActionDecide, By: "staff-1", At: at.Add(time.Hour), Observation: "completo", Outcome: "accept", SendBack: true, Next: "documents",
			Referral: &domain.ReferralRecord{From: "office", To: "commission", ExternalFileNumber: "EX-1"}},
		{ProcedureID: "P-1", Step: "documents", Action: definition.ActionUploadDocument, By: "ana", At: at.Add(2 * time.Hour),
			Document: &domain.DocumentRecord{ID: "D-1", Requirement: "statement", FileName: "constancia.pdf"}},
		{ProcedureID: "P-1", Step: "review", By: "staff-1", At: at.Add(3 * time.Hour), Observation: "faltan los códigos",
			CorrectionRequest: &domain.CorrectionRecord{Fields: []string{"subjects"}}},
	}

	got := procedureDetail(t, getProcedure(t, fakeProcedures{detail: d}, session))

	next, outcome, fileNumber, back := "review", "accept", "EX-1", "documents"
	ana, sofia := api.Person{Id: "ana", Name: "Ana Prueba"}, api.Person{Id: "staff-1", Name: "Sofía Ejemplo"}
	want := []api.ProcedureEvent{
		{Step: "documents", By: &ana, At: at, Next: &next},
		{Step: "review", By: &sofia, At: at.Add(time.Hour), Observation: "completo", Outcome: &outcome, SendBack: true, Next: &back,
			Referral: &api.Referral{From: "office", To: "commission", ExternalFileNumber: &fileNumber}},
		{Step: "documents", By: &ana, At: at.Add(2 * time.Hour),
			Document: &api.HistoryDocument{Id: "D-1", Document: "statement", FileName: "constancia.pdf"}},
		{Step: "review", By: &sofia, At: at.Add(3 * time.Hour), Observation: "faltan los códigos",
			CorrectionRequest: &api.HistoryCorrectionRequest{Fields: []string{"subjects"}, Documents: []string{}}},
	}
	if !reflect.DeepEqual(got.History, want) {
		t.Errorf("history = %+v, want %+v", got.History, want)
	}
}

// The applicant does not see who did a step of staff: the event has no one.
func TestGetProcedureLeavesOutWhoDidAStepWhenTheEventHasNoOne(t *testing.T) {
	d := openedDetail(false)
	d.History = []domain.Event{{ProcedureID: "P-1", Step: "review", Action: definition.ActionDecide}}

	got := procedureDetail(t, getProcedure(t, fakeProcedures{detail: d}, session))

	if len(got.History) != 1 || got.History[0].By != nil {
		t.Errorf("history = %+v, want one event without by", got.History)
	}
}

// The applicant does not see who asked for a correction: the request has no one.
func TestGetProcedureLeavesOutWhoAskedForACorrectionWhenTheRequestHasNoOne(t *testing.T) {
	d := openedDetail(false)
	d.CorrectionRequests = []domain.CorrectionRequest{{Step: "review", Reason: "faltan los códigos", Fields: []string{"subjects"}}}

	got := procedureDetail(t, getProcedure(t, fakeProcedures{detail: d}, session))

	if len(got.CorrectionRequests) != 1 || got.CorrectionRequests[0].By != nil {
		t.Errorf("correction requests = %+v, want one without by", got.CorrectionRequests)
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

func TestGetProcedureShowsWhereAReferredProcedureIs(t *testing.T) {
	referred := openedDetail(false)
	referred.State = domain.StateReferred
	referred.Referral = &detail.Referral{
		Area:              detail.Area{ID: "commission", Name: "Comisión Curricular"},
		From:              detail.Area{ID: "office", Name: "Dirección de Alumnos"},
		DaysWithoutReturn: 4,
	}

	got := procedureDetail(t, getProcedure(t, fakeProcedures{detail: referred}, session))

	if got.State != api.Referred || got.Referral == nil {
		t.Fatalf("state %s, referral %+v; want referred with its referral", got.State, got.Referral)
	}
	want := api.ProcedureReferral{
		Area:              api.NamedOption{Id: "commission", Name: "Comisión Curricular"},
		From:              api.NamedOption{Id: "office", Name: "Dirección de Alumnos"},
		DaysWithoutReturn: 4,
	}
	if *got.Referral != want {
		t.Errorf("referral = %+v, want %+v", *got.Referral, want)
	}
}

func TestGetProcedureOmitsTheReferralWhenNotReferred(t *testing.T) {
	recorder := getProcedure(t, fakeProcedures{detail: openedDetail(true)}, session)
	var got map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["referral"]; ok {
		t.Errorf("referral = %v, want it absent", got["referral"])
	}
}

func listMine(t *testing.T, procedures fakeProcedures, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, rest.NewProcedureHandler(fakeAuth{}, fakePeople{}, procedures, nil, nil), nil, nil)
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/me/procedures", nil)
	r.Header.Set("Cookie", "sigta_session="+cookie)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func TestListMyProceduresShowsWhereEachOneIs(t *testing.T) {
	started := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	moved := time.Date(2026, 11, 12, 15, 30, 0, 0, time.UTC)
	procedures := fakeProcedures{mine: []input.MyProcedure{{
		Procedure: domain.Procedure{ID: "P-1", ProcedureType: 7, Applicant: "ana", OpenStep: "check", StartedAt: started, MovedAt: moved},
		Name:      "Equivalencias",
		StepName:  "Revisión",
		State:     domain.StateInOffice,
	}, {
		Procedure: domain.Procedure{ID: "P-2", ProcedureType: 7, Applicant: "ana", StartedAt: started, MovedAt: moved},
		Name:      "Equivalencias",
		State:     domain.StateClosed,
		Resolutions: []detail.Resolution{{Name: "Resolución de Decano", File: domain.Document{
			ID: "D-1", Requirement: "dean_resolution", FileName: "resolucion.pdf", MediaType: "application/pdf", Size: 2048,
		}}},
	}, {
		Procedure: domain.Procedure{ID: "P-3", ProcedureType: 7, Applicant: "ana", OpenStep: "check", StartedAt: started, MovedAt: moved,
			CorrectionRequests: []domain.CorrectionRequest{{Step: "check", Reason: "falta la firma", Fields: []string{"subjects"}}}},
		Name:     "Equivalencias",
		StepName: "Revisión",
		State:    domain.StateWithApplicant,
	}}}

	recorder := listMine(t, procedures, session)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	var got []api.MyProcedure
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %+v, want three procedures", got)
	}
	if got[0].CorrectionPending || !got[2].CorrectionPending {
		t.Errorf("correction pending = %t, %t; want only the third one", got[0].CorrectionPending, got[2].CorrectionPending)
	}
	step := "Revisión"
	open := api.MyProcedure{Id: "P-1", ProcedureType: 7, ProcedureTypeName: "Equivalencias", StepName: &step, State: api.InOffice, StartedAt: started, LastMovedAt: moved, Resolutions: []api.ResolutionFile{}}
	if !reflect.DeepEqual(got[0], open) {
		t.Errorf("open procedure = %+v, want %+v", got[0], open)
	}
	if got[1].State != api.Closed || got[1].StepName != nil {
		t.Errorf("closed procedure = %+v, want it closed without an open step", got[1])
	}
	resolution := api.ResolutionFile{Id: "D-1", Document: "dean_resolution", DocumentName: "Resolución de Decano", FileName: "resolucion.pdf", MediaType: "application/pdf", Size: 2048}
	if !reflect.DeepEqual(got[1].Resolutions, []api.ResolutionFile{resolution}) {
		t.Errorf("resolutions = %+v, want %+v", got[1].Resolutions, resolution)
	}
}

func TestListMyProceduresAnswersAnEmptyList(t *testing.T) {
	recorder := listMine(t, fakeProcedures{mine: []input.MyProcedure{}}, session)
	if recorder.Code != http.StatusOK || recorder.Body.String() != "[]\n" {
		t.Errorf("got %d %q, want 200 and an empty list", recorder.Code, recorder.Body)
	}
}

func TestListMyProceduresNeedsASession(t *testing.T) {
	if recorder := listMine(t, fakeProcedures{}, "other"); recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", recorder.Code)
	}
}
