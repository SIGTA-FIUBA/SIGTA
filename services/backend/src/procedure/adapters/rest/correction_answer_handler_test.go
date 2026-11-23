package rest_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/submission"
)

func TestEnterProcedureValueAnswersTheValueOnTheStepOfTheApplicant(t *testing.T) {
	at := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	steps := &fakeSteps{value: domain.FieldValue{Procedure: "P-1", Field: "subjects", Step: "request", Items: []string{"61.03"}, EnteredBy: "ana", EnteredAt: at}}

	recorder := postToSteps(t, steps, "/values", session, `{"field":"subjects","items":["61.03"]}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	if steps.user != "ana" || steps.field != "subjects" || !reflect.DeepEqual(steps.entered, submission.Value{Items: []string{"61.03"}}) {
		t.Errorf("EnterValue got %q, %q and %+v", steps.user, steps.field, steps.entered)
	}
	var got api.ProcedureValue
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Field != "subjects" || got.Step != "request" || got.Items == nil || (*got.Items)[0] != "61.03" || got.EnteredBy.Name != "Ana Prueba" || !got.EnteredAt.Equal(at) {
		t.Errorf("value = %+v", got)
	}
}

func TestEnterProcedureValueErrors(t *testing.T) {
	body := `{"field":"subjects","text":"61.03"}`
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
		{"someone who is not the applicant", session, body, domain.ErrAccessDenied, http.StatusForbidden},
		{"nothing pending", session, body, engine.ErrNoCorrectionPending, http.StatusConflict},
		{"a field no pending correction asks for", session, body, engine.ErrNotRequested, http.StatusBadRequest},
		{"an invalid value", session, body, submission.ErrInvalidValue, http.StatusBadRequest},
		{"an unexpected failure", session, body, errors.New("store down"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := &fakeSteps{}
			if tt.err != nil {
				steps.err = fmt.Errorf("procedure P-1: %w", tt.err)
			}
			recorder := postToSteps(t, steps, "/values", tt.cookie, tt.body)
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d: %s", recorder.Code, tt.status, recorder.Body)
			}
		})
	}
}

func TestAnswerCorrectionAnswersTheRecordedEvent(t *testing.T) {
	at := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	steps := &fakeSteps{event: domain.Event{ProcedureID: "P-1", Step: "check", By: "ana", At: at,
		CorrectionAnswer: &domain.CorrectionRecord{Fields: []string{"subjects"}, Documents: []string{"syllabus"}}}}

	recorder := postToSteps(t, steps, "/correction-answer", session, ``)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	var got api.ProcedureEvent
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if steps.user != "ana" || got.Step != "check" || got.By == nil || got.By.Name != "Ana Prueba" || !got.At.Equal(at) {
		t.Errorf("event = %+v", got)
	}
	if a := got.CorrectionAnswer; a == nil || !reflect.DeepEqual(a.Fields, []string{"subjects"}) || !reflect.DeepEqual(a.Documents, []string{"syllabus"}) {
		t.Errorf("correction answer = %+v, want subjects and syllabus", got.CorrectionAnswer)
	}
}

func TestAnswerCorrectionNamesWhatIsMissing(t *testing.T) {
	steps := &fakeSteps{err: &engine.MissingAnswerError{
		Fields:    []definition.Field{{ID: "subjects", Name: "Asignaturas"}},
		Documents: []definition.Document{{ID: "syllabus", Name: "Programa"}},
	}}

	recorder := postToSteps(t, steps, "/correction-answer", session, ``)

	var got api.Error
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusBadRequest || got.Message != `Falta cargar de nuevo "Asignaturas" y "Programa".` {
		t.Errorf("got %d %q, want 400 naming both", recorder.Code, got.Message)
	}
}

func TestAnswerCorrectionErrors(t *testing.T) {
	tests := []struct {
		name   string
		cookie string
		err    error
		status int
	}{
		{"without a session", "other", nil, http.StatusUnauthorized},
		{"a procedure the person cannot see", session, domain.ErrProcedureNotFound, http.StatusNotFound},
		{"someone who is not the applicant", session, domain.ErrAccessDenied, http.StatusForbidden},
		{"nothing pending", session, engine.ErrNoCorrectionPending, http.StatusConflict},
		{"a request made in between", session, engine.ErrCorrectionsChanged, http.StatusConflict},
		{"an unexpected failure", session, errors.New("store down"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := &fakeSteps{}
			if tt.err != nil {
				steps.err = fmt.Errorf("procedure P-1: %w", tt.err)
			}
			recorder := postToSteps(t, steps, "/correction-answer", tt.cookie, ``)
			if recorder.Code != tt.status {
				t.Errorf("status = %d, want %d: %s", recorder.Code, tt.status, recorder.Body)
			}
			if tt.status != http.StatusInternalServerError && !strings.Contains(recorder.Body.String(), `"message"`) {
				t.Errorf("body = %s, want a message", recorder.Body)
			}
		})
	}
}
