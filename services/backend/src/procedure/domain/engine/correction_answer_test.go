package engine_test

import (
	"errors"
	"slices"
	"testing"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
)

const applicant = "applicant@fi.uba.ar"

// awaiting is on file with the corrections asked, each one with its event in
// the history.
func awaiting(requests ...domain.CorrectionRequest) (domain.Procedure, []domain.Event) {
	p := on("file")
	p.CorrectionRequests = requests
	var history []domain.Event
	for _, r := range requests {
		history = append(history, domain.Event{Step: r.Step, By: r.By, CorrectionRequest: &domain.CorrectionRecord{Fields: r.Fields, Documents: r.Documents}})
	}
	return p, history
}

func entered(by, field string) domain.Event {
	return domain.Event{Step: "request", By: by, Value: &domain.ValueRecord{Field: field}}
}

func attached(by, document string) domain.Event {
	return domain.Event{Step: "request", By: by, Document: &domain.DocumentRecord{Requirement: document}}
}

func TestTheAnswerLandsOnTheStepOfTheApplicantThatAsksForIt(t *testing.T) {
	p, _ := awaiting(correction("file", []string{"plan"}, []string{"transcript"}))

	if s, err := engine.CorrectionField(p, withRequirements, "plan"); err != nil || s.ID != "request" {
		t.Errorf("CorrectionField(plan) = %q, %v; want request, the step that asks for it as optional", s.ID, err)
	}
	if s, err := engine.CorrectionDocument(p, withRequirements, "transcript"); err != nil || s.ID != "request" {
		t.Errorf("CorrectionDocument(transcript) = %q, %v; want request", s.ID, err)
	}
}

func TestOnlyWhatAPendingCorrectionAsksForIsLoaded(t *testing.T) {
	p, _ := awaiting(correction("file", []string{"plan"}, nil))

	if _, err := engine.CorrectionField(p, withRequirements, "subjects"); !errors.Is(err, engine.ErrNotRequested) {
		t.Errorf("a field it does not ask for: err = %v, want ErrNotRequested", err)
	}
	if _, err := engine.CorrectionDocument(p, withRequirements, "plan"); !errors.Is(err, engine.ErrNotRequested) {
		t.Errorf("a field loaded as a document: err = %v, want ErrNotRequested", err)
	}
	if _, err := engine.CorrectionField(on("file"), withRequirements, "plan"); !errors.Is(err, engine.ErrNoCorrectionPending) {
		t.Errorf("nothing pending: err = %v, want ErrNoCorrectionPending", err)
	}
}

func TestAnswerCorrectionTakesTheProcedureBackToStaffOnTheSameStep(t *testing.T) {
	p, history := awaiting(
		correction("file", []string{"subjects"}, nil),
		correction("file", []string{"subjects"}, []string{"transcript"}),
	)
	history = append(history, entered(applicant, "subjects"), attached(applicant, "transcript"))

	answered, event, err := engine.AnswerCorrection(p, withRequirements, history, applicant, monday)
	if err != nil {
		t.Fatal(err)
	}
	if answered.AwaitingApplicant() || answered.OpenStep != "file" {
		t.Errorf("procedure = %+v, want it on file without pending corrections", answered)
	}
	record := event.CorrectionAnswer
	if event.Step != "file" || event.By != applicant || !event.At.Equal(monday) || record == nil ||
		!slices.Equal(record.Fields, []string{"subjects"}) || !slices.Equal(record.Documents, []string{"transcript"}) {
		t.Errorf("event = %+v, want the answer on file with everything asked, once", event)
	}
	if !p.AwaitingApplicant() {
		t.Error("AnswerCorrection changed the procedure it was given")
	}
}

func TestAnswerCorrectionNamesWhatWasNotLoadedAgain(t *testing.T) {
	first := correction("file", []string{"subjects"}, []string{"transcript"})
	p, history := awaiting(first)
	// Loaded before the second request, which asks for subjects again, and
	// by staff: neither counts.
	history = append(history, entered(applicant, "subjects"), attached("staff@fi.uba.ar", "transcript"))
	second := correction("file", []string{"subjects"}, nil)
	p.CorrectionRequests = append(p.CorrectionRequests, second)
	history = append(history, domain.Event{Step: "file", CorrectionRequest: &domain.CorrectionRecord{Fields: second.Fields}})

	_, _, err := engine.AnswerCorrection(p, withRequirements, history, applicant, monday)

	var missing *engine.MissingAnswerError
	if !errors.As(err, &missing) || len(missing.Fields) != 1 || missing.Fields[0].ID != "subjects" || len(missing.Documents) != 1 || missing.Documents[0].ID != "transcript" {
		t.Errorf("err = %v, want subjects and transcript missing", err)
	}
}

func TestAnswerCorrectionNeedsAPendingCorrection(t *testing.T) {
	if _, _, err := engine.AnswerCorrection(on("file"), withRequirements, nil, applicant, monday); !errors.Is(err, engine.ErrNoCorrectionPending) {
		t.Errorf("err = %v, want ErrNoCorrectionPending", err)
	}
}
