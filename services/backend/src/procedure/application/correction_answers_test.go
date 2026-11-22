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

// interruptedAttach is the memory repository that, right before an attachment
// is recorded, lets meanwhile change the procedure, once.
type interruptedAttach struct {
	*memory.ProcedureRepository
	meanwhile func()
}

func (r *interruptedAttach) Attach(ctx context.Context, d proceduredomain.Document, e proceduredomain.Event) error {
	if r.meanwhile != nil {
		meanwhile := r.meanwhile
		r.meanwhile = nil
		meanwhile()
	}
	return r.ProcedureRepository.Attach(ctx, d, e)
}

// attachingOn is an attachment service over a procedure on openStep whose
// attachments can be interrupted.
func attachingOn(t *testing.T, openStep string) (*ProcedureService, *DocumentService, *interruptedAttach, *memory.DocumentRepository) {
	t.Helper()
	procedures := &interruptedAttach{ProcedureRepository: memory.NewProcedureRepository()}
	service := confirmationFixtureOver(t, procedures, procedures.ProcedureRepository, openStep)
	documents := memory.NewDocumentRepository(procedures.ProcedureRepository)
	files := &memoryFiles{files: map[string][]byte{}, stored: map[string]time.Time{}, at: confirmedAt}
	return service, NewDocumentService(service, documents, files, func() string { return "D-1" }, func() time.Time { return confirmedAt }), procedures, documents
}

func proofUpload(name string, size int) input.Upload {
	return input.Upload{ProcedureID: "P-1", Requirement: "proof", FileName: name, Content: bytes.NewReader(pdf(size))}
}

// An answer sent while a document is being attached for it leaves the
// document out: it is refused, and not recorded after the answer.
func TestAnAttachmentAfterTheAnswerIsRefused(t *testing.T) {
	ctx := context.Background()
	service, attach, procedures, documents := attachingOn(t, "file")
	proof := input.CorrectionRequest{Step: "file", Reason: "falta la constancia", Fields: []string{"subjects"}, Documents: []string{"proof"}}
	if _, err := service.RequestCorrection(ctx, "operative@fi.uba.ar", "P-1", proof); err != nil {
		t.Fatal(err)
	}
	if _, err := service.EnterValue(ctx, titular, "P-1", "subjects", submission.Value{Text: "61.03"}); err != nil {
		t.Fatal(err)
	}
	if _, err := attach.Attach(ctx, titular, proofUpload("uno.pdf", 100)); err != nil {
		t.Fatal(err)
	}
	procedures.meanwhile = func() {
		if _, err := service.AnswerCorrection(ctx, titular, "P-1"); err != nil {
			t.Error(err)
		}
	}

	_, err := attach.Attach(ctx, titular, proofUpload("dos.pdf", 120))

	if !errors.Is(err, proceduredomain.ErrAccessDenied) {
		t.Errorf("err = %v, want ErrAccessDenied", err)
	}
	// Losing the race is refused as an attachment after the answer is.
	if _, after := attach.Attach(ctx, titular, proofUpload("tres.pdf", 140)); !errors.Is(after, proceduredomain.ErrAccessDenied) || after.Error() != err.Error() {
		t.Errorf("after the answer: err = %v, want %v", after, err)
	}
	if recorded, _ := documents.List(ctx, "P-1"); len(recorded) != 1 || recorded[0].FileName != "uno.pdf" {
		t.Errorf("documents = %+v, want only the one attached before the answer", recorded)
	}
}

// A procedure closed after the answer, while the document was being attached
// for it, refuses it as a closed procedure refuses any attachment.
func TestAnAttachmentAfterTheAnswerAndTheCloseIsRefusedAsOnAClosedProcedure(t *testing.T) {
	ctx := context.Background()
	service, attach, procedures, _ := attachingOn(t, "evaluate")
	proof := input.CorrectionRequest{Step: "evaluate", Reason: "falta la constancia", Documents: []string{"proof"}}
	if _, err := service.RequestCorrection(ctx, "commission@fi.uba.ar", "P-1", proof); err != nil {
		t.Fatal(err)
	}
	if _, err := attach.Attach(ctx, titular, proofUpload("uno.pdf", 100)); err != nil {
		t.Fatal(err)
	}
	procedures.meanwhile = func() {
		if _, err := service.AnswerCorrection(ctx, titular, "P-1"); err != nil {
			t.Error(err)
		}
		if _, err := service.Confirm(ctx, "commission@fi.uba.ar", "P-1", input.StepConfirmation{Step: "evaluate", Observation: "evaluated"}); err != nil {
			t.Error(err)
		}
	}

	_, err := attach.Attach(ctx, titular, proofUpload("dos.pdf", 120))

	if !errors.Is(err, engine.ErrStepNotOpen) {
		t.Errorf("err = %v, want ErrStepNotOpen", err)
	}
	if _, after := attach.Attach(ctx, titular, proofUpload("tres.pdf", 140)); !errors.Is(after, engine.ErrStepNotOpen) {
		t.Errorf("after the close: err = %v, want ErrStepNotOpen", after)
	}
}

// An attachment to the open step that loses the race against its
// confirmation is refused because the step is not open any more, as
// before: only an attachment in answer is answered again.
func TestAnAttachmentToTheOpenStepAfterItsConfirmationFindsItClosed(t *testing.T) {
	ctx := context.Background()
	_, attach, procedures, documents := attachingOn(t, "proofs")
	procedures.meanwhile = func() {
		// The applicant sends the step from another tab.
		moved := proceduredomain.Procedure{ID: "P-1", ProcedureType: 1, Version: 1, Applicant: titular, OpenStep: "file"}
		if err := procedures.Advance(ctx, moved, proceduredomain.Event{ProcedureID: "P-1", Step: "proofs", By: titular, At: confirmedAt}); err != nil {
			t.Error(err)
		}
	}

	_, err := attach.Attach(ctx, titular, proofUpload("uno.pdf", 100))

	if !errors.Is(err, engine.ErrStepNotOpen) {
		t.Errorf("err = %v, want ErrStepNotOpen", err)
	}
	if recorded, _ := documents.List(ctx, "P-1"); len(recorded) != 0 {
		t.Errorf("documents = %+v, want none", recorded)
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
