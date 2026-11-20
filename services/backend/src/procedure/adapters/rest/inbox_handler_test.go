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
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/input"
)

type fakeInbox struct {
	list  inbox.List
	err   error
	user  string
	query inbox.Query
	asked bool
}

func (f *fakeInbox) List(_ context.Context, userID string, q inbox.Query) (inbox.List, error) {
	f.user, f.query, f.asked = userID, q, true
	return f.list, f.err
}

// getProcedures asks with the session of fakeAuth, or with none if cookie is
// empty.
func listInbox(t *testing.T, box *fakeInbox, cookie, query string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, nil, rest.NewInboxHandler(fakeAuth{}, fakePeople{}, box), nil)
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures"+query, http.NoBody)
	if cookie != "" {
		r.Header.Set("Cookie", "sigta_session="+cookie)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func TestListProceduresAnswersTheList(t *testing.T) {
	started := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	box := &fakeInbox{list: inbox.List{
		Items: []inbox.Item{
			{Summary: inbox.Summary{
				ID: "p-1", ProcedureType: 7, ProcedureTypeName: "Trámite", ApplicantID: "ana",
				StepName: "Dictamen", Area: &inbox.Option[string]{ID: "committee", Name: "Comisión"},
				State: domain.StateInOffice, Actionable: true, StartedAt: started, LastMovedAt: started,
			}, AgeDays: 38},
			{Summary: inbox.Summary{
				ID: "p-2", ProcedureType: 7, ProcedureTypeName: "Trámite", ApplicantID: "unknown",
				StepName: "Solicitud", State: domain.StateWithApplicant, StartedAt: started, LastMovedAt: started,
			}},
		},
		ProcedureTypes: []inbox.Option[int64]{{ID: 7, Name: "Trámite"}},
		Areas:          []inbox.Option[string]{{ID: "committee", Name: "Comisión"}},
	}}

	recorder := listInbox(t, box, session, "?view=all&procedure_type=7&state=in_office&area=committee&sort=last_moved")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body)
	}
	want := inbox.Query{View: inbox.ViewAll, ProcedureType: 7, State: domain.StateInOffice, Area: "committee", Sort: inbox.SortLastMoved}
	if box.user != "ana" || box.query != want {
		t.Errorf("asked %q for %+v, want ana for %+v", box.user, box.query, want)
	}
	var got api.ProcedureList
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	item := got.Items[0]
	if item.Id != "p-1" || item.State != api.InOffice || !item.Actionable || item.AgeDays != 38 ||
		item.Applicant != (api.Person{Id: "ana", Name: "Ana Prueba"}) || item.Area == nil || *item.Area != (api.NamedOption{Id: "committee", Name: "Comisión"}) {
		t.Errorf("item = %+v", item)
	}
	// A step of the applicant has no area, and an unknown person no name.
	if other := got.Items[1]; other.Area != nil || other.Applicant != (api.Person{Id: "unknown"}) {
		t.Errorf("item of the applicant = %+v", other)
	}
	if got.ProcedureTypes[0] != (api.NamedOption{Id: "7", Name: "Trámite"}) {
		t.Errorf("type options = %v", got.ProcedureTypes)
	}
}

func TestListProceduresDefaultsToMineByAge(t *testing.T) {
	box := &fakeInbox{}

	recorder := listInbox(t, box, session, "")

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
		name   string
		cookie string
		query  string
		want   int
	}{
		{"without a session", "", "", http.StatusUnauthorized},
		{"with a session that is not valid", "expired", "", http.StatusUnauthorized},
		{"an unknown view", session, "?view=everything", http.StatusBadRequest},
		{"an unknown state", session, "?state=approved", http.StatusBadRequest},
		{"an unknown sort", session, "?sort=name", http.StatusBadRequest},
		{"a procedure type that is not a number", session, "?procedure_type=abc", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := &fakeInbox{}
			if recorder := listInbox(t, box, tt.cookie, tt.query); recorder.Code != tt.want {
				t.Errorf("status = %d, want %d", recorder.Code, tt.want)
			}
			if box.asked {
				t.Error("the inbox was queried")
			}
		})
	}
}

func TestListProceduresRefusesWhoIsNotStaff(t *testing.T) {
	recorder := listInbox(t, &fakeInbox{err: input.ErrNotStaff}, session, "")

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", recorder.Code)
	}
}

func TestListProceduresFailsWhenTheInboxFails(t *testing.T) {
	recorder := listInbox(t, &fakeInbox{err: errors.New("store down")}, session, "")

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", recorder.Code)
	}
}
