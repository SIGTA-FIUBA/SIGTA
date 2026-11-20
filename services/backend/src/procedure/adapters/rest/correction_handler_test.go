package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/input"
)

type fakeCorrections struct {
	input.Procedures
	event    domain.Event
	err      error
	user     string
	received input.CorrectionRequest
}

func (f *fakeCorrections) RequestCorrection(_ context.Context, userID, _ string, r input.CorrectionRequest) (domain.Event, error) {
	f.user, f.received = userID, r
	return f.event, f.err
}

func requestCorrection(t *testing.T, steps *fakeCorrections, cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, rest.NewProcedureHandler(fakeAuth{}, fakePeople{}, fakeProcedures{}, nil, steps), nil, nil)
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/procedures/P-1/steps/check/correction-requests", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Cookie", "sigta_session="+cookie)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func TestRequestCorrectionAnswersTheRecordedEvent(t *testing.T) {
	at := time.Date(2026, 11, 17, 9, 0, 0, 0, time.UTC)
	steps := &fakeCorrections{event: domain.Event{
		ProcedureID:       "P-1",
		Step:              "check",
		By:                "staff-1",
		At:                at,
		Observation:       "faltan los códigos",
		CorrectionRequest: &domain.CorrectionRecord{Fields: []string{"subjects"}, Documents: []string{"transcript"}},
	}}

	recorder := requestCorrection(t, steps, session, `{"reason":"faltan los códigos","fields":["subjects"],"documents":["transcript"]}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	want := input.CorrectionRequest{Step: "check", Reason: "faltan los códigos", Fields: []string{"subjects"}, Documents: []string{"transcript"}}
	if steps.user != "ana" || !reflect.DeepEqual(steps.received, want) {
		t.Errorf("RequestCorrection got user %q and %+v, want ana and %+v", steps.user, steps.received, want)
	}
	var got api.ProcedureEvent
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Step != "check" || got.By.Name != "Sofía Ejemplo" || !got.At.Equal(at) || got.Observation != "faltan los códigos" || got.Next != nil {
		t.Errorf("event = %+v", got)
	}
	if r := got.CorrectionRequest; r == nil || !reflect.DeepEqual(r.Fields, []string{"subjects"}) || !reflect.DeepEqual(r.Documents, []string{"transcript"}) {
		t.Errorf("correction request = %+v, want subjects and transcript", got.CorrectionRequest)
	}
}

func TestRequestCorrectionErrors(t *testing.T) {
	body := `{"reason":"faltan los códigos","fields":["subjects"]}`
	tests := []struct {
		name   string
		cookie string
		body   string
		err    error
		status int
	}{
		{"without a session", "other", body, nil, http.StatusUnauthorized},
		{"without a body", session, ``, nil, http.StatusBadRequest},
		{"a procedure the person cannot see", session, body, domain.ErrProcedureNotFound, http.StatusNotFound},
		{"without the permissions of the step", session, body, domain.ErrAccessDenied, http.StatusForbidden},
		{"a step that is no longer open", session, body, engine.ErrStepNotOpen, http.StatusConflict},
		{"an empty reason", session, `{"reason":" ","fields":["subjects"]}`, engine.ErrNoReason, http.StatusBadRequest},
		{"nothing asked for", session, `{"reason":"faltan"}`, engine.ErrNothingRequested, http.StatusBadRequest},
		{"a requirement the type does not have", session, body, engine.ErrNotARequirement, http.StatusBadRequest},
		{"a requirement twice", session, body, engine.ErrRepeatedRequirement, http.StatusBadRequest},
		{"a step of the applicant", session, body, engine.ErrCorrectionOnApplicant, http.StatusBadRequest},
		{"an unexpected failure", session, body, errors.New("store down"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := &fakeCorrections{}
			if tt.err != nil {
				steps.err = fmt.Errorf("procedure P-1: %w", tt.err)
			}
			recorder := requestCorrection(t, steps, tt.cookie, tt.body)
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d: %s", recorder.Code, tt.status, recorder.Body)
			}
			var got api.Error
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil || got.Message == "" {
				t.Errorf("body = %s, want an Error with a message", recorder.Body)
			}
		})
	}
}
