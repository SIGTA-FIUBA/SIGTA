package acceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	flow       []definition.Step
	definition definition.ProcedureTypeDefinition
	err        error
}

type failingAdvance struct {
	*memory.ProcedureRepository
}

func (failingAdvance) Advance(context.Context, domain.Procedure, domain.Event) error {
	return errRecording
}

func registerConfirmationSteps(sc *godog.ScenarioContext, base *authorizationWorld) *confirmationWorld {
	w := &confirmationWorld{authorizationWorld: base}

	sc.Given(`^a published "([^"]*)" procedure type with this flow:$`, w.givenFlow)
	sc.Given(`^"([^"]*)" confirmed the "([^"]*)" step(?: choosing "([^"]*)")?$`, w.givenConfirmed)
	sc.Given(`^recording confirmations fails$`, w.givenRecordingFails)
	sc.Given(`^it is "([^"]*)"$`, w.itIs)
	sc.Given(`^the "([^"]*)" documents? of the "([^"]*)" step (?:is|are) attached$`, w.givenAttached)

	sc.When(`^"([^"]*)" confirms the "([^"]*)" step with the observation "([^"]*)"$`, w.confirmWithObservation)
	sc.When(`^"([^"]*)" confirms the "([^"]*)" step(?: choosing "([^"]*)")?(?: with the external file number "([^"]*)")?$`, w.confirm)
	sc.When(`^"([^"]*)" and "([^"]*)" confirm the "([^"]*)" step at the same time$`, w.confirmAtTheSameTime)

	sc.Then(`^the confirmation is (recorded|denied|rejected|not recorded)$`, w.thenConfirmation)
	sc.Then(`^the confirmation is rejected because the step is not open$`, w.thenNotOpen)
	sc.Then(`^the confirmation is rejected because "([^"]*)" (?:is|are) missing$`, w.thenMissing)
	sc.Then(`^the "([^"]*)" step is confirmed by "([^"]*)" on "([^"]*)" with the observation "([^"]*)"$`, w.thenConfirmedBy)
	sc.Then(`^the "([^"]*)" step has (\d+) confirmations?$`, w.thenConfirmations)
	sc.Then(`^the "([^"]*)" step is recorded as declared by "([^"]*)", not verified by SIGTA$`, w.thenDeclaredBy)
	sc.Then(`^the procedure moves to the "([^"]*)" step$`, w.thenStaysOn)
	sc.Then(`^no step is left open$`, w.thenClosed)
	sc.Then(`^the procedure is referred to "([^"]*)"$`, w.thenReferredTo)
	sc.Then(`^the procedure is referred to "([^"]*)" for (\d+) days without return$`, w.thenReferredFor)
	sc.Then(`^the procedure is not referred$`, w.thenNotReferred)
	sc.Then(`^its state is (with the applicant|in an office|referred|closed)$`, w.thenState)
	sc.Then(`^its detail shows it in "([^"]*)", from "([^"]*)", for (\d+) days without return$`, w.thenDetailShowsReferral)
	sc.Then(`^the referral is recorded by "([^"]*)" on "([^"]*)" to "([^"]*)" with the external file number "([^"]*)"$`, w.thenReferralRecorded)
	sc.Then(`^the referral is recorded without an external file number$`, w.thenReferralWithoutNumber)
	return w
}

// givenFlow reads "done by" as the applicant or a staff area, and "next" as
// a step or, for a decision, as "outcome: step" pairs. The optional "asks for"
// column names the field, or the documents of an upload_document step, that
// the step asks for; a document that ends in "optional" is optional. Without
// it, a step of the applicant asks for a reason.
func (w *confirmationWorld) givenFlow(name string, table *godog.Table) error {
	d := definition.ProcedureTypeDefinition{Name: name, Description: "Descripción de " + name}
	reason := definition.Field{ID: "reason", Name: "Motivo", Type: definition.FieldLongText}
	for _, row := range tableRows(table) {
		id, doneBy, action, next := row["step"], row["done by"], definition.Action(row["action"]), row["next"]
		step := definition.Step{ID: id, Name: id, Action: action}
		if doneBy == "applicant" {
			step.Role = definition.RoleApplicant
		} else {
			step.Role = definition.RoleStaff
			step.Area = areaID(doneBy)
			step.Permissions = list(row["permissions"])
		}
		switch asked := row["asks for"]; {
		case asked != "" && action == definition.ActionUploadDocument:
			for _, name := range list(asked) {
				name, optional := strings.CutSuffix(name, " optional")
				d.Documents = append(d.Documents, definition.Document{ID: areaID(name), Name: name, AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: 1024})
				if optional {
					step.OptionalDocuments = append(step.OptionalDocuments, areaID(name))
				} else {
					step.Documents = append(step.Documents, areaID(name))
				}
			}
		case asked != "":
			d.Fields = append(d.Fields, definition.Field{ID: areaID(asked), Name: asked, Type: definition.FieldText})
			step.Fields = []string{areaID(asked)}
		case step.Role == definition.RoleApplicant:
			if !slices.ContainsFunc(d.Fields, func(f definition.Field) bool { return f.ID == reason.ID }) {
				d.Fields = append(d.Fields, reason)
			}
			step.Fields = []string{reason.ID}
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
		d.Steps = append(d.Steps, step)
	}
	w.steps[name] = d.Steps
	w.flow = d.Steps
	w.definition = d

	typeID, known := w.typeIDs[name]
	if !known {
		typeID = int64(len(w.typeIDs) + 1)
		w.typeIDs[name] = typeID
	}
	_, err := w.types.Publish(context.Background(), input.Publication{
		ProcedureType: typeID,
		Base:          w.latestVersion(name),
		Definition:    d,
		Authorship:    defaultAuthors,
	})
	return err
}

// tableRows reads each row after the header as a map from header to value.
func tableRows(table *godog.Table) []map[string]string {
	var rows []map[string]string
	for _, row := range table.Rows[1:] {
		values := map[string]string{}
		for i, cell := range row.Cells {
			values[table.Rows[0].Cells[i].Value] = cell.Value
		}
		rows = append(rows, values)
	}
	return rows
}

func (w *confirmationWorld) givenConfirmed(user, step, outcome string) error {
	return w.send(user, step, input.StepConfirmation{Observation: defaultObservation, Outcome: areaID(outcome)})
}

func (w *confirmationWorld) givenRecordingFails() error {
	clock := func() time.Time { return w.now }
	w.service = application.NewProcedureService(w.types, failingAdvance{w.procedures}, w.documents, w.memberships, w.newID, clock)
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

// givenAttached records a file for each document, as attaching it to the
// step would, without going through the upload.
func (w *confirmationWorld) givenAttached(documents, step string) error {
	for _, name := range list(documents) {
		err := w.documents.Add(context.Background(), domain.Document{
			ID:          w.newID(),
			Procedure:   w.current,
			Requirement: areaID(name),
			Step:        step,
			FileName:    areaID(name) + ".pdf",
			MediaType:   "application/pdf",
			Size:        1,
			Digest:      "sha256:" + areaID(name),
			UploadedBy:  "staff@fi.uba.ar",
			UploadedAt:  w.now,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *confirmationWorld) thenMissing(documents string) error {
	var missing *engine.MissingDocumentsError
	if !errors.As(w.err, &missing) {
		return fmt.Errorf("got %w, want the documents %s missing", w.err, documents)
	}
	var got []string
	for _, d := range missing.Documents {
		got = append(got, d.Name)
	}
	if want := list(documents); !slices.Equal(got, want) {
		return fmt.Errorf("missing %v, want %v", got, want)
	}
	return nil
}

func (w *confirmationWorld) thenNotOpen() error {
	if !errors.Is(w.err, engine.ErrStepNotOpen) {
		return fmt.Errorf("got %w, want %w", w.err, engine.ErrStepNotOpen)
	}
	return nil
}

func (w *confirmationWorld) eventsOf(step string) ([]domain.Event, error) {
	all, err := w.procedures.Events(context.Background(), w.current)
	if err != nil {
		return nil, err
	}
	var events []domain.Event
	for _, e := range all {
		if e.Step == step {
			events = append(events, e)
		}
	}
	return events, nil
}

func (w *confirmationWorld) onlyEventOf(step string) (domain.Event, error) {
	events, err := w.eventsOf(step)
	if err != nil {
		return domain.Event{}, err
	}
	if len(events) != 1 {
		return domain.Event{}, fmt.Errorf("the %q step has %d confirmations, want 1", step, len(events))
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

func (w *confirmationWorld) lastReferral() (domain.Event, error) {
	all, err := w.procedures.Events(context.Background(), w.current)
	if err != nil {
		return domain.Event{}, err
	}
	for i := len(all) - 1; i >= 0; i-- {
		if all[i].Referral != nil {
			return all[i], nil
		}
	}
	return domain.Event{}, errors.New("no referral was recorded")
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

func (w *confirmationWorld) thenState(state string) error {
	want := map[string]domain.State{
		"with the applicant": domain.StateWithApplicant,
		"in an office":       domain.StateInOffice,
		"referred":           domain.StateReferred,
		"closed":             domain.StateClosed,
	}[state]
	p, err := w.procedure()
	if err != nil {
		return err
	}
	got, err := domain.StateOf(p, definition.ProcedureTypeDefinition{Steps: w.flow})
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("state = %s, want %s", got, want)
	}
	return nil
}

func (w *confirmationWorld) thenDetailShowsReferral(area, from string, days int) error {
	details := application.NewProcedureDetailService(w.service, memory.NewDocumentRepository())
	shown, err := details.Detail(context.Background(), "applicant@fi.uba.ar", w.current)
	if err != nil {
		return err
	}
	r := shown.Referral
	if r == nil {
		return errors.New("the detail shows no referral")
	}
	if r.Area.Name != area || r.From.Name != from || r.DaysWithoutReturn != days {
		return fmt.Errorf("the detail shows it in %q, from %q, for %d days; want %q, %q, %d", r.Area.Name, r.From.Name, r.DaysWithoutReturn, area, from, days)
	}
	return nil
}
