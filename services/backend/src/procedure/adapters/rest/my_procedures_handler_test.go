package rest_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
)

func listMine(t *testing.T, procedures fakeProcedures, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, rest.NewProcedureHandler(fakeAuth{}, fakePeople{}, procedures, nil, nil), nil)
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
	}}}

	recorder := listMine(t, procedures, session)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	var got []api.MyProcedure
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := api.MyProcedure{Id: "P-1", ProcedureType: 7, ProcedureTypeName: "Equivalencias", StepName: "Revisión", State: api.InOffice, StartedAt: started, LastMovedAt: moved}
	if len(got) != 1 || got[0] != want {
		t.Errorf("got %+v, want [%+v]", got, want)
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
