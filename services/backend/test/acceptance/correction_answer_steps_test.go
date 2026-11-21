package acceptance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/cucumber/godog"

	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/domain/submission"
	"backend/src/procedure/ports/input"
)

// correctionAnswerWorld loads and sends what a correction asks for, over the
// world of the detail of a procedure, which already attaches files.
type correctionAnswerWorld struct {
	*myProcedureDetailWorld
	loadErr   error
	answerErr error
}

func registerCorrectionAnswerSteps(sc *godog.ScenarioContext, details *myProcedureDetailWorld) {
	w := &correctionAnswerWorld{myProcedureDetailWorld: details}

	sc.Given(`^"([^"]*)" entered "([^"]*)" as the "([^"]*)"$`, w.givenEntered)

	sc.When(`^"([^"]*)" enters "([^"]*)" as the "([^"]*)"$`, w.givenEntered)
	sc.When(`^"([^"]*)" attaches "([^"]*)" as the "([^"]*)"$`, w.givenAttachedFile)
	sc.When(`^"([^"]*)" tries to attach "([^"]*)" as the "([^"]*)"$`, w.tryAttach)
	sc.When(`^"([^"]*)" sends the answer to the correction$`, w.answer)

	sc.Then(`^they are asked for "([^"]*)" because "([^"]*)", by "([^"]*)"$`, w.thenAskedBy)
	sc.Then(`^the "([^"]*)" of the "([^"]*)" step is "([^"]*)"$`, w.thenLoaded)
	sc.Then(`^the attachment is rejected because (no pending correction asks for it|it is not up to the applicant)$`, w.thenAttachmentRejected)
	sc.Then(`^the answer is (sent|denied)$`, w.thenAnswer)
	sc.Then(`^the answer is rejected because "([^"]*)" (?:is|are) missing$`, w.thenMissing)
	sc.Then(`^"([^"]*)" has the procedure in their inbox$`, w.thenInInbox)
	sc.Then(`^"([^"]*)" sees in the history the answer of "([^"]*)" on "([^"]*)", on the "([^"]*)" step$`, w.thenAnswerInHistory)
	sc.Then(`^the procedure shows a pending correction$`, w.thenPendingInMine)
}

func (w *correctionAnswerWorld) givenEntered(user, text, field string) error {
	_, err := w.service.EnterValue(context.Background(), user, w.current, idOf(field), submission.Value{Text: text})
	return err
}

func (w *correctionAnswerWorld) tryAttach(user, fileName, requirement string) error {
	documents, err := w.documentService()
	if err != nil {
		return err
	}
	_, w.loadErr = documents.Attach(context.Background(), user, input.Upload{
		ProcedureID: w.current,
		Requirement: idOf(requirement),
		FileName:    fileName,
		Content:     bytes.NewReader([]byte("%PDF-1.7\n" + fileName)),
	})
	return nil
}

func (w *correctionAnswerWorld) answer(user string) {
	_, w.answerErr = w.service.AnswerCorrection(context.Background(), user, w.current)
}

// thenAskedBy reads the detail as the applicant, who sees the area of the
// step the correction was asked on, not who asked it.
func (w *correctionAnswerWorld) thenAskedBy(requirements, reason, area string) error {
	if w.readErr != nil {
		return w.readErr
	}
	for _, r := range w.detail.CorrectionRequests {
		if r.Reason != reason || !slices.Equal(slices.Concat(r.Fields, r.Documents), idsOf(requirements)) {
			continue
		}
		if r.By != "" {
			return fmt.Errorf("the applicant sees who asked: %q", r.By)
		}
		for _, s := range w.detail.Steps {
			if s.Step.ID == r.Step && s.Area != nil && s.Area.Name == area {
				return nil
			}
		}
		return fmt.Errorf("the correction was asked on %q, which is not in %q", r.Step, area)
	}
	return fmt.Errorf("the applicant sees %+v, want %q because %q", w.detail.CorrectionRequests, requirements, reason)
}

// thenLoaded reads the detail as the applicant: what they loaded has to be on
// one of their steps, where they see it.
func (w *correctionAnswerWorld) thenLoaded(requirement, step, want string) error {
	applicant, err := w.applicantOf()
	if err != nil {
		return err
	}
	shown, err := application.NewProcedureDetailService(w.service, w.documents).Detail(context.Background(), applicant, w.current)
	if err != nil {
		return err
	}
	for _, s := range shown.Steps {
		if s.Step.ID != step {
			continue
		}
		for _, v := range s.Values {
			if v.Field == idOf(requirement) && v.Text == want {
				return nil
			}
		}
		for _, d := range s.Files {
			if d.Requirement == idOf(requirement) && d.FileName == want {
				return nil
			}
		}
		return fmt.Errorf("step %s has values %+v and files %+v, want %q as the %s", step, s.Values, s.Files, want, requirement)
	}
	return fmt.Errorf("the applicant does not see the step %q", step)
}

func (w *correctionAnswerWorld) thenAttachmentRejected(reason string) error {
	want := engine.ErrNotRequested
	if reason == "it is not up to the applicant" {
		want = domain.ErrAccessDenied
	}
	if !errors.Is(w.loadErr, want) {
		return fmt.Errorf("got %w, want the attachment rejected because %s", w.loadErr, reason)
	}
	return nil
}

func (w *correctionAnswerWorld) thenAnswer(result string) error {
	switch result {
	case "sent":
		if w.answerErr != nil {
			return fmt.Errorf("the answer failed: %w", w.answerErr)
		}
	case "denied":
		if !errors.Is(w.answerErr, domain.ErrAccessDenied) {
			return fmt.Errorf("got %w, want the answer denied", w.answerErr)
		}
	}
	return nil
}

func (w *correctionAnswerWorld) thenMissing(requirements string) error {
	var missing *engine.MissingAnswerError
	if !errors.As(w.answerErr, &missing) {
		return fmt.Errorf("got %w, want the answer rejected for what is missing", w.answerErr)
	}
	var got []string
	for _, f := range missing.Fields {
		got = append(got, f.ID)
	}
	for _, d := range missing.Documents {
		got = append(got, d.ID)
	}
	if !slices.Equal(got, idsOf(requirements)) {
		return fmt.Errorf("missing %v, want %v", got, idsOf(requirements))
	}
	return nil
}

func (w *correctionAnswerWorld) thenInInbox(user string) error {
	list, err := application.NewInboxService(w.service).List(context.Background(), user, inbox.Query{View: inbox.ViewMine})
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(list.Items, func(item inbox.Item) bool { return item.ID == w.current && item.Actionable }) {
		return fmt.Errorf("the inbox of %s does not have %s to act on: %+v", user, w.current, list.Items)
	}
	return nil
}

func (w *correctionAnswerWorld) thenAnswerInHistory(reader, applicant, on, step string) error {
	shown, err := application.NewProcedureDetailService(w.service, w.documents).Detail(context.Background(), reader, w.current)
	if err != nil {
		return err
	}
	if len(shown.History) == 0 {
		return errors.New("the history is empty")
	}
	last := shown.History[len(shown.History)-1]
	if last.CorrectionAnswer == nil || last.By != applicant || last.At.Format(time.DateOnly) != on || last.Step != step {
		return fmt.Errorf("the last entry is %+v, want the answer of %s on %s, on %s", last, applicant, on, step)
	}
	return nil
}

func (w *correctionAnswerWorld) thenPendingInMine() error {
	applicant, err := w.applicantOf()
	if err != nil {
		return err
	}
	mine, err := application.NewProcedureDetailService(w.service, w.documents).Mine(context.Background(), applicant)
	if err != nil {
		return err
	}
	for _, m := range mine {
		if m.Procedure.ID == w.current && m.Procedure.AwaitingApplicant() {
			return nil
		}
	}
	return fmt.Errorf("my procedures are %+v, want %s with a pending correction", mine, w.current)
}

func (w *correctionAnswerWorld) applicantOf() (string, error) {
	p, err := w.procedures.Get(context.Background(), w.current)
	return p.Applicant, err
}

func idsOf(names string) []string {
	var ids []string
	for _, name := range list(names) {
		ids = append(ids, idOf(name))
	}
	return ids
}
