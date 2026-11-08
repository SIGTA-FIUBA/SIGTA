package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/input"
)

type fakeInbox struct {
	list   inbox.List
	err    error
	person inbox.Person
	query  inbox.Query
	asked  bool
}

func (f *fakeInbox) List(_ context.Context, person inbox.Person, q inbox.Query) (inbox.List, error) {
	f.person, f.query, f.asked = person, q, true
	return f.list, f.err
}

func getProcedures(t *testing.T, box *fakeInbox, role string, authErr error, query string) *httptest.ResponseRecorder {
	t.Helper()
	authenticate := func(context.Context, string) (inbox.Person, error) {
		return inbox.Person{ID: "u1", Role: role}, authErr
	}
	server := api.NewServer(nil, nil, rest.NewInboxHandler(box, authenticate))
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures"+query, http.NoBody))
	return recorder
}

func TestListProceduresAnswersTheList(t *testing.T) {
	started := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	days := 4
	box := &fakeInbox{list: inbox.List{
		Items: []inbox.Item{{
			Summary: inbox.Summary{
				ID: "p-1", ProcedureType: 7, ProcedureTypeName: "Trámite", ApplicantName: "Titular de prueba",
				StepName: "Dictamen", Area: definition.Area{ID: "committee", Name: "Comisión", Kind: definition.AreaExternal},
				StartedAt: started, LastMovedAt: started,
			},
			State: inbox.StateReferred, AgeDays: 38, DaysReferred: &days,
		}},
		ProcedureTypes: []inbox.Option[int64]{{ID: 7, Name: "Trámite"}},
		Areas:          []inbox.Option[string]{{ID: "committee", Name: "Comisión"}},
	}}

	recorder := getProcedures(t, box, "staff_approver", nil, "?view=all&procedure_type=7&state=referred&area=committee&sort=last_moved")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body)
	}
	want := inbox.Query{View: inbox.ViewAll, ProcedureType: 7, State: inbox.StateReferred, Area: "committee", Sort: inbox.SortLastMoved}
	if box.person != (inbox.Person{ID: "u1", Role: "staff_approver"}) || box.query != want {
		t.Errorf("asked %+v for %+v, want %+v", box.person, box.query, want)
	}
	var got api.ProcedureList
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	item := got.Items[0]
	if item.Id != "p-1" || item.State != api.Referred || item.DaysReferred == nil || *item.DaysReferred != 4 || item.AreaName != "Comisión" {
		t.Errorf("item = %+v", item)
	}
	if got.ProcedureTypes[0] != (api.NamedOption{Id: "7", Name: "Trámite"}) {
		t.Errorf("type options = %v", got.ProcedureTypes)
	}
}

func TestListProceduresDefaultsToMineByAge(t *testing.T) {
	box := &fakeInbox{}

	recorder := getProcedures(t, box, "staff_operative", nil, "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if box.query != (inbox.Query{View: inbox.ViewMine, Sort: inbox.SortOldest}) {
		t.Errorf("query = %+v", box.query)
	}
	// Lists are [] in the JSON, never null.
	if body := recorder.Body.String(); body != `{"areas":[],"items":[],"procedure_types":[]}`+"\n" {
		t.Errorf("body = %s", body)
	}
}

func TestListProceduresRefuses(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		authErr error
		query   string
		want    int
	}{
		{"without a session", "", rest.ErrNoSession, "", http.StatusUnauthorized},
		{"an unknown view", "staff_operative", nil, "?view=everything", http.StatusBadRequest},
		{"an unknown state", "staff_operative", nil, "?state=approved", http.StatusBadRequest},
		{"an unknown sort", "staff_operative", nil, "?sort=name", http.StatusBadRequest},
		{"a procedure type that is not a number", "staff_operative", nil, "?procedure_type=abc", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := &fakeInbox{}
			if recorder := getProcedures(t, box, tt.role, tt.authErr, tt.query); recorder.Code != tt.want {
				t.Errorf("status = %d, want %d", recorder.Code, tt.want)
			}
			if box.asked {
				t.Error("the inbox was queried")
			}
		})
	}
}

func TestListProceduresRefusesWhoIsNotStaff(t *testing.T) {
	recorder := getProcedures(t, &fakeInbox{err: input.ErrNotStaff}, "applicant", nil, "")

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", recorder.Code)
	}
}

func TestListProceduresFailsWhenTheInboxFails(t *testing.T) {
	recorder := getProcedures(t, &fakeInbox{err: errors.New("store down")}, "staff_operative", nil, "")

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", recorder.Code)
	}
}
