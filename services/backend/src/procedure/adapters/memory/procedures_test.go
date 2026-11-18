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
