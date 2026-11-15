package acceptance

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"

	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/input"
)

const defaultObservation = "done"

var errRecording = errors.New("recording confirmations fails")

type confirmationWorld struct {
	*authorizationWorld
	err error
}

type failingAdvance struct {
	*memory.ProcedureRepository
}

func (failingAdvance) Advance(context.Context, domain.Procedure, engine.Event) error {
	return errRecording
}

func registerConfirmationSteps(sc *godog.ScenarioContext, base *authorizationWorld) {
	w := &confirmationWorld{authorizationWorld: base}

	sc.Given(`^a published "([^"]*)" procedure type with this flow:$`, w.givenFlow)
	sc.Given(`^"([^"]*)" confirmed the "([^"]*)" step(?: choosing "([^"]*)")?$`, w.givenConfirmed)
	sc.Given(`^recording confirmations fails$`, w.givenRecordingFails)
	sc.Given(`^it is "([^"]*)"$`, w.itIs)

	sc.When(`^"([^"]*)" confirms the "([^"]*)" step with the observation "([^"]*)"$`, w.confirmWithObservation)
	sc.When(`^"([^"]*)" confirms the "([^"]*)" step(?: choosing "([^"]*)")?(?: with the external file number "([^"]*)")?$`, w.confirm)
	sc.When(`^"([^"]*)" and "([^"]*)" confirm the "([^"]*)" step at the same time$`, w.confirmAtTheSameTime)

	sc.Then(`^the confirmation is (recorded|denied|rejected|not recorded)$`, w.thenConfirmation)
	sc.Then(`^the confirmation is rejected because the step is not open$`, w.thenNotOpen)
	sc.Then(`^the "([^"]*)" step is confirmed by "([^"]*)" on "([^"]*)" with the observation "([^"]*)"$`, w.thenConfirmedBy)
	sc.Then(`^the "([^"]*)" step has (\d+) confirmations?$`, w.thenConfirmations)
	sc.Then(`^the "([^"]*)" step is recorded as declared by "([^"]*)", not verified by SIGTA$`, w.thenDeclaredBy)
	sc.Then(`^the procedure moves to the "([^"]*)" step$`, w.thenStaysOn)
	sc.Then(`^no step is left open$`, w.thenClosed)
	sc.Then(`^the procedure is referred to "([^"]*)"$`, w.thenReferredTo)
	sc.Then(`^the procedure is referred to "([^"]*)" for (\d+) days without return$`, w.thenReferredFor)
	sc.Then(`^the procedure is not referred$`, w.thenNotReferred)
	sc.Then(`^the referral is recorded by "([^"]*)" on "([^"]*)" to "([^"]*)" with the external file number "([^"]*)"$`, w.thenReferralRecorded)
	sc.Then(`^the referral is recorded without an external file number$`, w.thenReferralWithoutNumber)
}

// givenFlow reads "done by" as the applicant or a staff area, and "next" as
// a step or, for a decision, as "outcome: step" pairs.
func (w *confirmationWorld) givenFlow(name string, table *godog.Table) error {
	var steps []definition.Step
	for _, row := range table.Rows[1:] {
		id, doneBy, action, next := row.Cells[0].Value, row.Cells[1].Value, definition.Action(row.Cells[3].Value), row.Cells[4].Value
		step := definition.Step{ID: id, Name: id, Action: action}
		if doneBy == "applicant" {
			step.Role = definition.RoleApplicant
			step.Fields = []string{"reason"}
		} else {
			step.Role = definition.RoleStaff
			step.Area = areaID(doneBy)
			step.Permissions = list(row.Cells[2].Value)
		}
		if action == definition.ActionDecide {
			for _, pair := range list(next) {
				outcome, to, _ := strings.Cut(pair, ":")
				outcome = strings.TrimSpace(outcome)
				step.Outcomes = append(step.Outcomes, definition.Outcome{ID: areaID(outcome), Name: outcome, Next: strings.TrimSpace(to)})
			}
		} else {
			step.Next = next
		}
		steps = append(steps, step)
	}
	w.steps[name] = steps

	typeID, known := w.typeIDs[name]
	if !known {
		typeID = int64(len(w.typeIDs) + 1)
		w.typeIDs[name] = typeID
	}
	_, err := w.types.Publish(context.Background(), input.Publication{
		ProcedureType: typeID,
		Base:          w.latestVersion(name),
		Definition: definition.ProcedureTypeDefinition{
			Name:        name,
			Description: "Descripción de " + name,
			Fields:      []definition.Field{{ID: "reason", Name: "Motivo", Type: definition.FieldLongText}},
			Steps:       steps,
		},
		Authorship: defaultAuthors,
	})
	return err
}

func (w *confirmationWorld) givenConfirmed(user, step, outcome string) error {
	return w.send(user, step, input.StepConfirmation{Observation: defaultObservation, Outcome: areaID(outcome)})
}

func (w *confirmationWorld) givenRecordingFails() error {
	clock := func() time.Time { return w.now }
	w.service = application.NewProcedureService(w.types, failingAdvance{w.procedures}, w.memberships, w.newID, clock)
	return nil
}

func (w *confirmationWorld) send(user, step string, c input.StepConfirmation) error {
	c.Step = step
	_, err := w.service.Confirm(context.Background(), user, w.current, c)
	return err
}

func (w *confirmationWorld) confirmWithObservation(user, step, observation string) {
	w.err = w.send(user, step, input.StepConfirmation{Observation: observation})
}

func (w *confirmationWorld) confirm(user, step, outcome, fileNumber string) {
	w.err = w.send(user, step, input.StepConfirmation{
		Observation:        defaultObservation,
		Outcome:            areaID(outcome),
		ExternalFileNumber: fileNumber,
	})
}

func (w *confirmationWorld) confirmAtTheSameTime(first, second, step string) {
	var wg sync.WaitGroup
	for _, user := range []string{first, second} {
		wg.Go(func() {
			_ = w.send(user, step, input.StepConfirmation{Observation: defaultObservation})
		})
	}
	wg.Wait()
}

func (w *confirmationWorld) thenConfirmation(result string) error {
	var want error
	switch result {
	case "recorded":
		if w.err != nil {
			return fmt.Errorf("the confirmation failed: %w", w.err)
		}
		return nil
	case "denied":
		want = domain.ErrAccessDenied
	case "rejected":
		want = engine.ErrInvalidConfirmation
	case "not recorded":
		want = errRecording
	}
	if !errors.Is(w.err, want) {
		return fmt.Errorf("got %w, want %w", w.err, want)
	}
	return nil
}

func (w *confirmationWorld) thenNotOpen() error {
	if !errors.Is(w.err, engine.ErrStepNotOpen) {
		return fmt.Errorf("got %w, want %w", w.err, engine.ErrStepNotOpen)
	}
	return nil
}

func (w *confirmationWorld) eventsOf(step string) ([]engine.Event, error) {
	all, err := w.procedures.Events(context.Background(), w.current)
	if err != nil {
		return nil, err
	}
	var events []engine.Event
	for _, e := range all {
		if e.Step == step {
			events = append(events, e)
		}
	}
	return events, nil
}

func (w *confirmationWorld) onlyEventOf(step string) (engine.Event, error) {
	events, err := w.eventsOf(step)
	if err != nil {
		return engine.Event{}, err
	}
	if len(events) != 1 {
		return engine.Event{}, fmt.Errorf("the %q step has %d confirmations, want 1", step, len(events))
	}
	return events[0], nil
}

func (w *confirmationWorld) thenConfirmedBy(step, user, date, observation string) error {
	e, err := w.onlyEventOf(step)
	if err != nil {
		return err
	}
	if e.By != user || e.At.Format(time.DateOnly) != date || e.Observation != observation {
		return fmt.Errorf("confirmed by %s on %s with %q, want %s on %s with %q", e.By, e.At.Format(time.DateOnly), e.Observation, user, date, observation)
	}
	return nil
}

func (w *confirmationWorld) thenConfirmations(step string, n int) error {
	events, err := w.eventsOf(step)
	if err != nil {
		return err
	}
	if len(events) != n {
		return fmt.Errorf("the %q step has %d confirmations, want %d", step, len(events), n)
	}
	return nil
}

func (w *confirmationWorld) thenDeclaredBy(step, user string) error {
	e, err := w.onlyEventOf(step)
	if err != nil {
		return err
	}
	if !e.Declared || e.By != user {
		return fmt.Errorf("declared = %v by %s, want declared by %s", e.Declared, e.By, user)
	}
	return nil
}

func (w *confirmationWorld) procedure() (domain.Procedure, error) {
	return w.procedures.Get(context.Background(), w.current)
}

func (w *confirmationWorld) thenClosed() error {
	p, err := w.procedure()
	if err != nil {
		return err
	}
	if !p.Closed() {
		return fmt.Errorf("the procedure is on %q, want it closed", p.OpenStep)
	}
	return nil
}

func (w *confirmationWorld) thenReferredTo(area string) error {
	p, err := w.procedure()
	if err != nil {
		return err
	}
	if p.Referral == nil || p.Referral.To != areaID(area) {
		return fmt.Errorf("referral = %+v, want it referred to %s", p.Referral, areaID(area))
	}
	return nil
}

func (w *confirmationWorld) thenReferredFor(area string, days int) error {
	if err := w.thenReferredTo(area); err != nil {
		return err
	}
	p, err := w.procedure()
	if err != nil {
		return err
	}
	if got, _ := p.DaysWithoutReturn(w.now); got != days {
		return fmt.Errorf("%d days without return, want %d", got, days)
	}
	return nil
}

func (w *confirmationWorld) thenNotReferred() error {
	p, err := w.procedure()
	if err != nil {
		return err
	}
	if p.Referral != nil {
		return fmt.Errorf("referral = %+v, want none", p.Referral)
	}
	return nil
}

func (w *confirmationWorld) lastReferral() (engine.Event, error) {
	all, err := w.procedures.Events(context.Background(), w.current)
	if err != nil {
		return engine.Event{}, err
	}
	for i := len(all) - 1; i >= 0; i-- {
		if all[i].Referral != nil {
			return all[i], nil
		}
	}
	return engine.Event{}, errors.New("no referral was recorded")
}

func (w *confirmationWorld) thenReferralRecorded(user, date, area, fileNumber string) error {
	e, err := w.lastReferral()
	if err != nil {
		return err
	}
	if e.By != user || e.At.Format(time.DateOnly) != date || e.Referral.To != areaID(area) || e.Referral.ExternalFileNumber != fileNumber {
		return fmt.Errorf("referral by %s on %s to %s with %q, want %s on %s to %s with %q",
			e.By, e.At.Format(time.DateOnly), e.Referral.To, e.Referral.ExternalFileNumber, user, date, areaID(area), fileNumber)
	}
	return nil
}

func (w *confirmationWorld) thenReferralWithoutNumber() error {
	e, err := w.lastReferral()
	if err != nil {
		return err
	}
	if e.Referral.ExternalFileNumber != "" {
		return fmt.Errorf("external file number = %q, want none", e.Referral.ExternalFileNumber)
	}
	return nil
}
