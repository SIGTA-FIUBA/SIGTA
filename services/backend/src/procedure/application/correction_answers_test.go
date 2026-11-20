package application

import (
	"context"
	"errors"
	"sync"
	"testing"

	proceduredomain "backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/submission"
)

const titular = "applicant@fi.uba.ar"

func TestEnterValueIsOnlyForTheApplicantAndOnlyValid(t *testing.T) {
	ctx := context.Background()
	service, procedures := confirmationFixture(t)
	if _, err := service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", fileCorrection()); err != nil {
		t.Fatal(err)
	}

	if _, err := service.EnterValue(ctx, "operative@fi.uba.ar", "P-1", "subjects", submission.Value{Text: "61.03"}); !errors.Is(err, proceduredomain.ErrAccessDenied) {
		t.Errorf("staff entering it: err = %v, want ErrAccessDenied", err)
	}
	if _, err := service.EnterValue(ctx, titular, "P-1", "subjects", submission.Value{Text: "  "}); !errors.Is(err, submission.ErrInvalidValue) {
		t.Errorf("a blank value: err = %v, want ErrInvalidValue", err)
	}
	v, err := service.EnterValue(ctx, titular, "P-1", "subjects", submission.Value{Text: " 61.03 "})
	if err != nil {
		t.Fatal(err)
	}
	if v.Step != "request" || v.Text != "61.03" || v.EnteredBy != titular || !v.EnteredAt.Equal(confirmedAt) {
		t.Errorf("value = %+v, want 61.03 on request, by the applicant", v)
	}
	values, _ := procedures.Values(ctx, "P-1")
	if len(values) != 1 {
		t.Errorf("values = %+v, want only the valid one", values)
	}
}

// The first answer goes through; staff then asks for something else; a second
// answer read before that request, with as many pending as then, must not
// answer it unseen.
func TestAnAnswerDoesNotAnswerARequestMadeAfterItWasRead(t *testing.T) {
	ctx := context.Background()
	service, procedures := confirmationFixture(t)
	if _, err := service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", fileCorrection()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.EnterValue(ctx, titular, "P-1", "subjects", submission.Value{Text: "61.03"}); err != nil {
		t.Fatal(err)
	}
	read, _ := procedures.Get(ctx, "P-1")
	history, _ := procedures.Events(ctx, "P-1")
	seen := engine.CorrectionRequestsIn(history)
	answered, event, err := engine.AnswerCorrection(read, mustDefinition(t, service, read), history, titular, confirmedAt)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.AnswerCorrection(ctx, titular, "P-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", fileCorrection()); err != nil {
		t.Fatal(err)
	}

	if err := procedures.AnswerCorrection(ctx, answered, seen, event); !errors.Is(err, engine.ErrCorrectionsChanged) {
		t.Errorf("the late answer: err = %v, want ErrCorrectionsChanged", err)
	}
	if p, _ := procedures.Get(ctx, "P-1"); !p.AwaitingApplicant() {
		t.Error("the request made after the answer was read is not pending any more")
	}
}

func TestConcurrentAnswersAndARequestNeverAnswerTheRequestUnseen(t *testing.T) {
	for round := range 50 {
		ctx := context.Background()
		service, procedures := confirmationFixture(t)
		if _, err := service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", fileCorrection()); err != nil {
			t.Fatal(err)
		}
		if _, err := service.EnterValue(ctx, titular, "P-1", "subjects", submission.Value{Text: "61.03"}); err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() { _, _ = service.AnswerCorrection(ctx, titular, "P-1") })
		}
		wg.Go(func() { _, _ = service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", fileCorrection()) })
		wg.Wait()

		history, _ := procedures.Events(ctx, "P-1")
		answers, lastRequest, lastAnswer := 0, -1, -1
		for i, e := range history {
			if e.CorrectionAnswer != nil {
				answers++
				lastAnswer = i
			}
			if e.CorrectionRequest != nil {
				lastRequest = i
			}
		}
		p, _ := procedures.Get(ctx, "P-1")
		if answers > 1 {
			t.Fatalf("round %d: %d answers recorded, want at most one", round, answers)
		}
		if lastRequest > lastAnswer && !p.AwaitingApplicant() {
			t.Fatalf("round %d: the request after the answer is not pending", round)
		}
	}
}

func mustDefinition(t *testing.T, s *ProcedureService, p proceduredomain.Procedure) definition.ProcedureTypeDefinition {
	t.Helper()
	c, err := s.content(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	return c.Definition
}
