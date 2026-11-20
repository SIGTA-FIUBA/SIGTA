package acceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/cucumber/godog"

	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/input"
)

type correctionWorld struct {
	*confirmationWorld
	err error
}

func registerCorrectionRequestSteps(sc *godog.ScenarioContext, base *confirmationWorld) {
	w := &correctionWorld{confirmationWorld: base}

	sc.Given(`^"([^"]*)" requested a correction of "([^"]*)" because "([^"]*)"$`, w.givenRequested)

	sc.When(`^"([^"]*)" requests a correction of "([^"]*)" because "([^"]*)"$`, w.request)

	sc.Then(`^the correction request is (recorded|denied|rejected)$`, w.thenRequest)
	sc.Then(`^the applicant is asked to correct "([^"]*)" because "([^"]*)"$`, w.thenAsked)
	sc.Then(`^the procedure leaves the inbox of "([^"]*)"$`, w.thenLeavesInbox)
}

func (w *correctionWorld) givenRequested(user, requirements, reason string) error {
	return w.send(user, requirements, reason)
}

func (w *correctionWorld) request(user, requirements, reason string) {
	w.err = w.send(user, requirements, reason)
}

// send asks for the open step; each requirement is a field of the procedure
// type or, if it is not one, a document.
func (w *correctionWorld) send(user, requirements, reason string) error {
	p, err := w.procedures.Get(context.Background(), w.current)
	if err != nil {
		return err
	}
	r := input.CorrectionRequest{Step: p.OpenStep, Reason: reason}
	for _, name := range list(requirements) {
		id := idOf(name)
		if slices.ContainsFunc(w.definition.Documents, func(d definition.Document) bool { return d.ID == id }) {
			r.Documents = append(r.Documents, id)
		} else {
			r.Fields = append(r.Fields, id)
		}
	}
	_, err = w.service.RequestCorrection(context.Background(), user, w.current, r)
	return err
}

func (w *correctionWorld) thenRequest(result string) error {
	var want error
	switch result {
	case "recorded":
		if w.err != nil {
			return fmt.Errorf("the correction request failed: %w", w.err)
		}
		return nil
	case "denied":
		want = domain.ErrAccessDenied
	case "rejected":
		want = engine.ErrInvalidCorrectionRequest
	}
	if !errors.Is(w.err, want) {
		return fmt.Errorf("got %w, want %w", w.err, want)
	}
	return nil
}

// thenAsked reads the detail with the account of the applicant, who is the one
// that has to see what is asked.
func (w *correctionWorld) thenAsked(requirements, reason string) error {
	p, err := w.procedures.Get(context.Background(), w.current)
	if err != nil {
		return err
	}
	details := application.NewProcedureDetailService(w.service, w.documents)
	shown, err := details.Detail(context.Background(), p.Applicant, w.current)
	if err != nil {
		return err
	}
	want := list(requirements)
	for _, r := range shown.CorrectionRequests {
		if r.Reason == reason && slices.Equal(slices.Concat(r.Fields, r.Documents), want) {
			return nil
		}
	}
	return fmt.Errorf("the applicant sees the correction requests %+v, want one of %v because %q", shown.CorrectionRequests, want, reason)
}

func (w *correctionWorld) thenLeavesInbox(user string) error {
	list, err := application.NewInboxService(w.service).List(context.Background(), user, inbox.Query{View: inbox.ViewMine})
	if err != nil {
		return err
	}
	if slices.ContainsFunc(list.Items, func(item inbox.Item) bool { return item.ID == w.current }) {
		return fmt.Errorf("the inbox of %s still lists %s", user, w.current)
	}
	return nil
}
