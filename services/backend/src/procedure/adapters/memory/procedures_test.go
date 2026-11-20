package memory

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
)

func onStep(step string) domain.Procedure {
	return domain.Procedure{ID: "P-1", ProcedureType: 1, Version: 1, Applicant: "applicant@fi.uba.ar", OpenStep: step}
}

func confirmed(step, next string) (domain.Procedure, domain.Event) {
	return onStep(next), domain.Event{ProcedureID: "P-1", Step: step, By: "staff@fi.uba.ar", Observation: "done", Next: next}
}

func TestAdvanceStoresTheProcedureAndAppendsTheEvent(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}

	for _, step := range [][2]string{{"file", "approve"}, {"approve", ""}} {
		p, e := confirmed(step[0], step[1])
		if err := s.Advance(ctx, p, e); err != nil {
			t.Fatalf("Advance(%s) = %v", step[0], err)
		}
	}

	got, err := s.Get(ctx, "P-1")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Closed() {
		t.Errorf("open step = %q, want the procedure closed", got.OpenStep)
	}
	events, err := s.Events(ctx, "P-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Step != "file" || events[1].Step != "approve" {
		t.Errorf("events = %+v, want file and then approve", events)
	}
}

func TestAddRecordsTheEventsOfTheStartBeforeLaterOnes(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	start := []domain.Event{
		{ProcedureID: "P-1", Step: "request", By: "applicant@fi.uba.ar", Next: "attach"},
		{ProcedureID: "P-1", Step: "attach", By: "applicant@fi.uba.ar", Next: "file"},
	}
	if err := s.Add(ctx, onStep("file"), start); err != nil {
		t.Fatal(err)
	}
	p, e := confirmed("file", "approve")
	if err := s.Advance(ctx, p, e); err != nil {
		t.Fatal(err)
	}

	events, err := s.Events(ctx, "P-1")
	if err != nil {
		t.Fatal(err)
	}
	var steps []string
	for _, e := range events {
		steps = append(steps, e.Step)
	}
	if !slices.Equal(steps, []string{"request", "attach", "file"}) {
		t.Errorf("steps = %v, want request, attach and file", steps)
	}
}

func TestAdvanceFromAStepThatIsNotOpenChangesNothing(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}

	p, e := confirmed("approve", "")
	if err := s.Advance(ctx, p, e); !errors.Is(err, engine.ErrStepNotOpen) {
		t.Fatalf("Advance = %v, want ErrStepNotOpen", err)
	}

	got, err := s.Get(ctx, "P-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.OpenStep != "file" {
		t.Errorf("open step = %q, want it to stay on file", got.OpenStep)
	}
	if events, _ := s.Events(ctx, "P-1"); len(events) != 0 {
		t.Errorf("events = %+v, want none", events)
	}
}

func TestAdvanceAnUnknownProcedure(t *testing.T) {
	p, e := confirmed("file", "approve")
	if err := NewProcedureRepository().Advance(context.Background(), p, e); !errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("Advance = %v, want ErrProcedureNotFound", err)
	}
	if _, err := NewProcedureRepository().Events(context.Background(), "P-1"); !errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("Events = %v, want ErrProcedureNotFound", err)
	}
}

func TestRecordAppendsTheEventWithoutMovingTheProcedure(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}
	attached := domain.Event{ProcedureID: "P-1", Step: "file", By: "staff@fi.uba.ar", Document: &domain.DocumentRecord{ID: "D-1", Requirement: "ruling", FileName: "acta.pdf"}}
	if err := s.Record(ctx, attached); err != nil {
		t.Fatal(err)
	}
	p, e := confirmed("file", "approve")
	if err := s.Advance(ctx, p, e); err != nil {
		t.Fatalf("Advance after Record = %v", err)
	}

	events, err := s.Events(ctx, "P-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Document == nil || *events[0].Document != *attached.Document || events[1].Step != "file" {
		t.Errorf("events = %+v, want the attachment and then the confirmation of file", events)
	}

	attached.Document.FileName = "changed"
	events[0].Document.ID = "changed"
	stored, _ := s.Events(ctx, "P-1")
	if *stored[0].Document != (domain.DocumentRecord{ID: "D-1", Requirement: "ruling", FileName: "acta.pdf"}) {
		t.Errorf("stored document = %+v, want it unchanged", *stored[0].Document)
	}
}

func TestRecordOnAnUnknownProcedure(t *testing.T) {
	e := domain.Event{ProcedureID: "P-1", Step: "file"}
	if err := NewProcedureRepository().Record(context.Background(), e); !errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("Record = %v, want ErrProcedureNotFound", err)
	}
}

func TestConcurrentConfirmationsOfTheSameStepAdvanceOnce(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}

	const attempts = 20
	var wg sync.WaitGroup
	errs := make(chan error, attempts)
	for range attempts {
		wg.Go(func() {
			p, e := confirmed("file", "approve")
			errs <- s.Advance(ctx, p, e)
		})
	}
	wg.Wait()
	close(errs)

	advanced := 0
	for err := range errs {
		switch {
		case err == nil:
			advanced++
		case !errors.Is(err, engine.ErrStepNotOpen):
			t.Errorf("Advance = %v, want nil or ErrStepNotOpen", err)
		}
	}
	if advanced != 1 {
		t.Errorf("%d confirmations advanced the procedure, want 1", advanced)
	}
	if events, _ := s.Events(ctx, "P-1"); len(events) != 1 {
		t.Errorf("%d events recorded, want 1", len(events))
	}
}

func TestStoredReferralsCannotBeChangedFromOutside(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("check"), nil); err != nil {
		t.Fatal(err)
	}
	since := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)
	p := onStep("evaluate")
	p.Referral = &domain.Referral{From: "affairs", To: "commission", Since: since}
	e := domain.Event{ProcedureID: "P-1", Step: "check", Next: "evaluate", Referral: &domain.ReferralRecord{From: "affairs", To: "commission"}}
	if err := s.Advance(ctx, p, e); err != nil {
		t.Fatal(err)
	}

	p.Referral.To = "changed"
	e.Referral.To = "changed"
	got, _ := s.Get(ctx, "P-1")
	got.Referral.From = "changed"
	events, _ := s.Events(ctx, "P-1")
	events[0].Referral.From = "changed"

	stored, _ := s.Get(ctx, "P-1")
	if *stored.Referral != (domain.Referral{From: "affairs", To: "commission", Since: since}) {
		t.Errorf("stored referral = %+v, want it unchanged", *stored.Referral)
	}
	storedEvents, _ := s.Events(ctx, "P-1")
	if *storedEvents[0].Referral != (domain.ReferralRecord{From: "affairs", To: "commission"}) {
		t.Errorf("stored event referral = %+v, want it unchanged", *storedEvents[0].Referral)
	}
}

func requested(step string) (domain.CorrectionRequest, domain.Event) {
	r := domain.CorrectionRequest{Step: step, By: "staff@fi.uba.ar", Reason: "missing codes", Fields: []string{"subjects"}}
	return r, domain.Event{ProcedureID: "P-1", Step: step, By: "staff@fi.uba.ar", Observation: "missing codes", CorrectionRequest: &domain.CorrectionRecord{Fields: []string{"subjects"}}}
}

func TestRequestCorrectionKeepsEachRequestPendingAndAppendsItsEvent(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		r, e := requested("file")
		if err := s.RequestCorrection(ctx, r, e); err != nil {
			t.Fatal(err)
		}
	}

	got, _ := s.Get(ctx, "P-1")
	if got.OpenStep != "file" || len(got.CorrectionRequests) != 2 {
		t.Errorf("procedure = %+v, want it on file with two pending requests", got)
	}
	if events, _ := s.Events(ctx, "P-1"); len(events) != 2 || events[1].CorrectionRequest == nil {
		t.Errorf("events = %+v, want both requests", events)
	}
}

func TestRequestCorrectionOnAStepThatIsNotOpenChangesNothing(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}
	r, e := requested("approve")
	if err := s.RequestCorrection(ctx, r, e); !errors.Is(err, engine.ErrStepNotOpen) {
		t.Fatalf("RequestCorrection = %v, want ErrStepNotOpen", err)
	}
	got, _ := s.Get(ctx, "P-1")
	events, _ := s.Events(ctx, "P-1")
	if len(got.CorrectionRequests) != 0 || len(events) != 0 {
		t.Errorf("procedure = %+v with events %+v, want no request", got, events)
	}
	if err := NewProcedureRepository().RequestCorrection(ctx, r, e); !errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("RequestCorrection on an unknown procedure = %v, want ErrProcedureNotFound", err)
	}
}

// A confirmation read before a correction request must not move the procedure
// once the request is stored.
func TestAdvanceWhileTheApplicantHasACorrectionToMakeChangesNothing(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}
	r, e := requested("file")
	if err := s.RequestCorrection(ctx, r, e); err != nil {
		t.Fatal(err)
	}

	p, confirmation := confirmed("file", "approve")
	if err := s.Advance(ctx, p, confirmation); !errors.Is(err, engine.ErrAwaitingApplicant) {
		t.Fatalf("Advance = %v, want ErrAwaitingApplicant", err)
	}
	if got, _ := s.Get(ctx, "P-1"); got.OpenStep != "file" || len(got.CorrectionRequests) != 1 {
		t.Errorf("procedure = %+v, want it on file with its request", got)
	}
}

func TestStoredCorrectionRequestsCannotBeChangedFromOutside(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}
	r, e := requested("file")
	if err := s.RequestCorrection(ctx, r, e); err != nil {
		t.Fatal(err)
	}

	r.Fields[0] = "changed"
	e.CorrectionRequest.Fields[0] = "changed"
	got, _ := s.Get(ctx, "P-1")
	got.CorrectionRequests[0].Fields[0] = "changed"
	events, _ := s.Events(ctx, "P-1")
	events[0].CorrectionRequest.Fields[0] = "changed"

	stored, _ := s.Get(ctx, "P-1")
	storedEvents, _ := s.Events(ctx, "P-1")
	if stored.CorrectionRequests[0].Fields[0] != "subjects" || storedEvents[0].CorrectionRequest.Fields[0] != "subjects" {
		t.Errorf("stored request = %+v and event = %+v, want them unchanged", stored.CorrectionRequests[0], *storedEvents[0].CorrectionRequest)
	}
}

func TestEnterValueOnlyWhileACorrectionIsPending(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}
	v := domain.FieldValue{Procedure: "P-1", Field: "subjects", Step: "request", Items: []string{"61.03"}}
	e := domain.Event{ProcedureID: "P-1", Step: "request", Value: &domain.ValueRecord{Field: "subjects"}}
	if err := s.EnterValue(ctx, v, e); !errors.Is(err, engine.ErrNoCorrectionPending) {
		t.Errorf("nothing pending: err = %v, want ErrNoCorrectionPending", err)
	}

	r, requestEvent := requested("file")
	if err := s.RequestCorrection(ctx, r, requestEvent); err != nil {
		t.Fatal(err)
	}
	if err := s.EnterValue(ctx, v, e); err != nil {
		t.Fatal(err)
	}
	values, _ := s.Values(ctx, "P-1")
	events, _ := s.Events(ctx, "P-1")
	if len(values) != 1 || values[0].Field != "subjects" || len(events) != 2 || events[1].Value == nil || events[1].Value.Field != "subjects" {
		t.Errorf("values = %+v, events = %+v; want the value and its event", values, events)
	}
}

func TestAnswerCorrectionRefusesWhenAnotherRequestCameInBetween(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureRepository()
	if err := s.Add(ctx, onStep("file"), nil); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		r, e := requested("file")
		if err := s.RequestCorrection(ctx, r, e); err != nil {
			t.Fatal(err)
		}
	}
	answered := onStep("file")
	answer := domain.Event{ProcedureID: "P-1", Step: "file", CorrectionAnswer: &domain.CorrectionRecord{}}

	if err := s.AnswerCorrection(ctx, answered, 1, answer); !errors.Is(err, engine.ErrCorrectionsChanged) {
		t.Errorf("answering one of two: err = %v, want ErrCorrectionsChanged", err)
	}
	if err := s.AnswerCorrection(ctx, answered, 2, answer); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get(ctx, "P-1")
	events, _ := s.Events(ctx, "P-1")
	if got.AwaitingApplicant() || len(events) != 3 || events[2].CorrectionAnswer == nil {
		t.Errorf("procedure = %+v, events = %+v; want no pending corrections and the answer last", got, events)
	}
}
