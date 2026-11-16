package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/input"
)

type fakeSteps struct {
	input.Procedures
	event    engine.Event
	err      error
	user     string
	received input.StepConfirmation
}

func (f *fakeSteps) Confirm(_ context.Context, userID, _ string, c input.StepConfirmation) (engine.Event, error) {
	f.user, f.received = userID, c
	return f.event, f.err
}

func confirmStep(t *testing.T, steps *fakeSteps, cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, rest.NewProcedureHandler(fakeAuth{}, fakePeople{}, fakeProcedures{}, steps))
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/procedures/P-1/steps/check/confirmation", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Cookie", "sigta_session="+cookie)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

func TestConfirmStepAnswersTheRecordedEvent(t *testing.T) {
	at := time.Date(2026, 11, 16, 10, 30, 0, 0, time.UTC)
	steps := &fakeSteps{event: engine.Event{
		ProcedureID: "P-1",
		Step:        "check",
		Action:      definition.ActionDecide,
		By:          "staff-1",
		At:          at,
		Observation: "Documentación completa",
		Outcome:     "accept",
		Next:        "evaluate",
		Referral:    &engine.ReferralRecord{From: "affairs", To: "commission", ExternalFileNumber: "EX-1"},
	}}

	recorder := confirmStep(t, steps, session, `{"observation":"Documentación completa","outcome":"accept","external_file_number":"EX-1"}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	want := input.StepConfirmation{Step: "check", Observation: "Documentación completa", Outcome: "accept", ExternalFileNumber: "EX-1"}
	if steps.user != "ana" || steps.received != want {
		t.Errorf("Confirm got user %q and %+v, want ana and %+v", steps.user, steps.received, want)
	}
	var got api.ProcedureEvent
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Step != "check" || got.By.Name != "Sofía Ejemplo" || !got.At.Equal(at) || got.Next == nil || *got.Next != "evaluate" || got.Outcome == nil || *got.Outcome != "accept" {
		t.Errorf("event = %+v", got)
	}
	if got.Referral == nil || got.Referral.To != "commission" || got.Referral.ExternalFileNumber == nil || *got.Referral.ExternalFileNumber != "EX-1" {
		t.Errorf("referral = %+v, want to commission with EX-1", got.Referral)
	}
}

func TestConfirmStepThatClosesOmitsNext(t *testing.T) {
	steps := &fakeSteps{event: engine.Event{Step: "approve", Action: definition.ActionConfirmExternalFact, By: "staff-1", Observation: "ok", Declared: true}}

	recorder := confirmStep(t, steps, session, `{"observation":"ok"}`)

	var got map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	for _, absent := range []string{"next", "outcome", "referral"} {
		if _, ok := got[absent]; ok {
			t.Errorf("%s = %v, want it absent", absent, got[absent])
		}
	}
	if got["declared"] != true {
		t.Errorf("declared = %v, want true", got["declared"])
	}
}

func TestConfirmStepErrors(t *testing.T) {
	tests := []struct {
		name   string
		cookie string
		body   string
		err    error
		status int
	}{
		{"without a session", "other", `{"observation":"ok"}`, nil, http.StatusUnauthorized},
		{"without a body", session, ``, nil, http.StatusBadRequest},
		{"a procedure the person cannot see", session, `{"observation":"ok"}`, domain.ErrProcedureNotFound, http.StatusNotFound},
		{"without the permissions of the step", session, `{"observation":"ok"}`, domain.ErrAccessDenied, http.StatusForbidden},
		{"a step that is no longer open", session, `{"observation":"ok"}`, engine.ErrStepNotOpen, http.StatusConflict},
		{"an empty observation", session, `{"observation":" "}`, engine.ErrNoObservation, http.StatusBadRequest},
		{"an outcome that is not of the step", session, `{"observation":"ok","outcome":"x"}`, engine.ErrInvalidOutcome, http.StatusBadRequest},
		{"a file number without referral", session, `{"observation":"ok","external_file_number":"EX-1"}`, engine.ErrNotAReferral, http.StatusBadRequest},
		{"a step of the applicant", session, `{"observation":"ok"}`, engine.ErrApplicantStep, http.StatusBadRequest},
		{"an unexpected failure", session, `{"observation":"ok"}`, errors.New("store down"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := &fakeSteps{err: tt.err}
			if tt.err != nil {
				steps.err = fmt.Errorf("procedure P-1: %w", tt.err)
			}
			recorder := confirmStep(t, steps, tt.cookie, tt.body)
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d: %s", recorder.Code, tt.status, recorder.Body)
			}
			var body api.Error
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.Message == "" {
				t.Errorf("body = %s, want an Error with a message", recorder.Body)
			}
		})
	}
}
