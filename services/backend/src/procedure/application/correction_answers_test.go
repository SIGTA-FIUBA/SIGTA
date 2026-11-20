package application

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"backend/src/procedure/adapters/memory"
	proceduredomain "backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/submission"
	"backend/src/procedure/ports/input"
)

const titular = "applicant@fi.uba.ar"

func TestEnterValueIsOnlyForTheApplicantAndOnlyValid(t *testing.T) {
	ctx := context.Background()
	service, procedures := confirmationFixture(t, "file")
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
	service, procedures := confirmationFixture(t, "file")
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
	for round := range 200 {
		ctx := context.Background()
		service, procedures := confirmationFixture(t, "file")
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
		another := fileCorrection()
		another.Fields = []string{"plan"}
		wg.Go(func() { _, _ = service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", another) })
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

// requestingOnHistory is the memory repository that, the first time the
// history is read, lets staff request another correction just before.
type requestingOnHistory struct {
	*memory.ProcedureRepository
	request func()
}

func (r *requestingOnHistory) Events(ctx context.Context, procedureID string) ([]proceduredomain.Event, error) {
	if r.request != nil {
		request := r.request
		r.request = nil
		request()
	}
	return r.ProcedureRepository.Events(ctx, procedureID)
}

// A request that comes in between reading the procedure and reading its
// history stays pending: the answer does not store the procedure as it was
// before it.
func TestARequestBetweenTheReadsOfAnAnswerStaysPending(t *testing.T) {
	ctx := context.Background()
	procedures := &requestingOnHistory{ProcedureRepository: memory.NewProcedureRepository()}
	service := confirmationFixtureOver(t, procedures, procedures.ProcedureRepository, "file")
	if _, err := service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", fileCorrection()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.EnterValue(ctx, titular, "P-1", "subjects", submission.Value{Text: "61.03"}); err != nil {
		t.Fatal(err)
	}
	// The new request asks for something the first one did not.
	another := fileCorrection()
	another.Fields = []string{"plan"}
	procedures.request = func() {
		if _, err := service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", another); err != nil {
			t.Error(err)
		}
	}

	_, err := service.AnswerCorrection(ctx, titular, "P-1")

	var missing *engine.MissingAnswerError
	if !errors.As(err, &missing) {
		t.Errorf("err = %v, want the answer to miss what the new request asks for", err)
	}
	if p, _ := procedures.Get(ctx, "P-1"); len(p.CorrectionRequests) != 2 {
		t.Errorf("pending = %d, want both requests", len(p.CorrectionRequests))
	}
}

// answeringOnRecord is the memory repository that, right before an attachment
// in answer is recorded, lets the applicant send their answer.
type answeringOnRecord struct {
	*memory.ProcedureRepository
	answer func()
}

func (r *answeringOnRecord) Attach(ctx context.Context, d proceduredomain.Document, e proceduredomain.Event) error {
	if r.answer != nil {
		answer := r.answer
		r.answer = nil
		answer()
	}
	return r.ProcedureRepository.Attach(ctx, d, e)
}

// An answer sent while a document is being attached for it leaves the
// document out: it is refused, and not recorded after the answer.
func TestAnAttachmentAfterTheAnswerIsRefused(t *testing.T) {
	ctx := context.Background()
	procedures := &answeringOnRecord{ProcedureRepository: memory.NewProcedureRepository()}
	service := confirmationFixtureOver(t, procedures, procedures.ProcedureRepository, "file")
	documents := memory.NewDocumentRepository(procedures.ProcedureRepository)
	files := &memoryFiles{files: map[string][]byte{}, stored: map[string]time.Time{}, at: confirmedAt}
	attach := NewDocumentService(service, documents, files, func() string { return "D-1" }, func() time.Time { return confirmedAt })
	proof := input.CorrectionRequest{Step: "file", Reason: "falta la constancia", Fields: []string{"subjects"}, Documents: []string{"proof"}}
	if _, err := service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", proof); err != nil {
		t.Fatal(err)
	}
	if _, err := service.EnterValue(ctx, titular, "P-1", "subjects", submission.Value{Text: "61.03"}); err != nil {
		t.Fatal(err)
	}
	if _, err := attach.Attach(ctx, titular, input.Upload{ProcedureID: "P-1", Requirement: "proof", FileName: "uno.pdf", Content: bytes.NewReader(pdf(100))}); err != nil {
		t.Fatal(err)
	}
	procedures.answer = func() {
		if _, err := service.AnswerCorrection(ctx, titular, "P-1"); err != nil {
			t.Error(err)
		}
	}

	_, err := attach.Attach(ctx, titular, input.Upload{ProcedureID: "P-1", Requirement: "proof", FileName: "dos.pdf", Content: bytes.NewReader(pdf(120))})

	if !errors.Is(err, engine.ErrNoCorrectionPending) {
		t.Errorf("err = %v, want ErrNoCorrectionPending", err)
	}
	if recorded, _ := documents.List(ctx, "P-1"); len(recorded) != 1 || recorded[0].FileName != "uno.pdf" {
		t.Errorf("documents = %+v, want only the one attached before the answer", recorded)
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
