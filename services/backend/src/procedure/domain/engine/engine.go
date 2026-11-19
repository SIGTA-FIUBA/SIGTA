// Package engine moves a procedure forward when its open step is confirmed.
// Who may confirm is checked by the use case.
package engine

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
)

var (
	ErrInvalidConfirmation = errors.New("invalid confirmation")
	ErrStepNotOpen         = fmt.Errorf("%w: the step is not open", ErrInvalidConfirmation)
	ErrNoObservation       = fmt.Errorf("%w: an observation is required", ErrInvalidConfirmation)
	ErrInvalidOutcome      = fmt.Errorf("%w: the outcome is not one of the step", ErrInvalidConfirmation)
	ErrNotAReferral        = fmt.Errorf("%w: an external file number goes only with a referral", ErrInvalidConfirmation)
	ErrApplicantStep       = fmt.Errorf("%w: the applicant completes their steps, staff confirms theirs", ErrInvalidConfirmation)
	ErrMissingDocuments    = fmt.Errorf("%w: the step asks for documents that are not attached", ErrInvalidConfirmation)
)

// MissingDocumentsError names the mandatory documents of a step that have no
// file attached to it, in the order the step asks for them.
type MissingDocumentsError struct {
	Documents []definition.Document
}

func (e *MissingDocumentsError) Error() string {
	ids := make([]string, 0, len(e.Documents))
	for _, d := range e.Documents {
		ids = append(ids, d.ID)
	}
	return fmt.Sprintf("%v: %s", ErrMissingDocuments, strings.Join(ids, ", "))
}

func (e *MissingDocumentsError) Unwrap() error { return ErrMissingDocuments }

type Confirmation struct {
	Step               string
	By                 string
	At                 time.Time
	Observation        string
	Outcome            string
	ExternalFileNumber string
	// Attached holds the documents of the procedure type that have a file
	// attached to the step being confirmed.
	Attached []string
}

func Confirm(p domain.Procedure, d definition.ProcedureTypeDefinition, c Confirmation) (domain.Procedure, domain.Event, error) {
	if p.Closed() || c.Step != p.OpenStep {
		return p, domain.Event{}, ErrStepNotOpen
	}
	step, ok := find(d.Steps, c.Step)
	if !ok {
		return p, domain.Event{}, fmt.Errorf("procedure %s: open step %q is not in its version", p.ID, c.Step)
	}
	if step.Role == definition.RoleApplicant {
		return p, domain.Event{}, ErrApplicantStep
	}
	if strings.TrimSpace(c.Observation) == "" {
		return p, domain.Event{}, ErrNoObservation
	}
	if err := MissingDocuments(d, step, c.Attached); err != nil {
		return p, domain.Event{}, err
	}
	nextID, err := next(step, c.Outcome)
	if err != nil {
		return p, domain.Event{}, err
	}

	event := domain.Event{
		ProcedureID: p.ID,
		Step:        step.ID,
		Action:      step.Action,
		By:          c.By,
		At:          c.At,
		Observation: c.Observation,
		Outcome:     c.Outcome,
		Declared:    step.Action == definition.ActionConfirmExternalFact,
		Next:        nextID,
	}
	moved := p
	moved.OpenStep = nextID
	moved.MovedAt = c.At
	if nextID == "" {
		// Closing ends the referral too: a closed procedure is nowhere.
		moved.Referral = nil
	} else {
		nextStep, ok := find(d.Steps, nextID)
		if !ok {
			return p, domain.Event{}, fmt.Errorf("procedure %s: step %q leads to %q, which is not in its version", p.ID, step.ID, nextID)
		}
		moved.Referral, event.Referral, event.Returned = refer(p.Referral, step.Area, nextStep.Area, c)
	}
	if c.ExternalFileNumber != "" && event.Referral == nil {
		return p, domain.Event{}, ErrNotAReferral
	}
	return moved, event, nil
}

// Open records the opening steps, which the applicant completes by submitting
// the procedure.
func Open(p domain.Procedure, opening []definition.Step, by string, at time.Time) []domain.Event {
	events := make([]domain.Event, 0, len(opening))
	for _, s := range opening {
		events = append(events, domain.Event{ProcedureID: p.ID, Step: s.ID, Action: s.Action, By: by, At: at, Next: s.Next})
	}
	return events
}

// MissingDocuments fails with a MissingDocumentsError if a mandatory document
// of the step is not among the attached ones. Optional documents do not count.
// Submitting a procedure checks the same for the opening steps of the
// applicant (submission.MissingDocuments).
func MissingDocuments(d definition.ProcedureTypeDefinition, step definition.Step, attached []string) error {
	var missing []definition.Document
	for _, id := range step.Documents {
		if slices.Contains(attached, id) {
			continue
		}
		doc := definition.Document{ID: id}
		if i := slices.IndexFunc(d.Documents, func(candidate definition.Document) bool { return candidate.ID == id }); i >= 0 {
			doc = d.Documents[i]
		}
		missing = append(missing, doc)
	}
	if len(missing) > 0 {
		return &MissingDocumentsError{Documents: missing}
	}
	return nil
}

func find(steps []definition.Step, id string) (definition.Step, bool) {
	i := slices.IndexFunc(steps, func(s definition.Step) bool { return s.ID == id })
	if i < 0 {
		return definition.Step{}, false
	}
	return steps[i], true
}

func next(step definition.Step, outcome string) (string, error) {
	if step.Action != definition.ActionDecide {
		if outcome != "" {
			return "", ErrInvalidOutcome
		}
		return step.Next, nil
	}
	i := slices.IndexFunc(step.Outcomes, func(o definition.Outcome) bool { return o.ID == outcome })
	if i < 0 {
		return "", ErrInvalidOutcome
	}
	return step.Outcomes[i].Next, nil
}

// refer follows ADR-023: moving to a step of another area refers the
// procedure, and moving back to the area it first left returns it. Applicant
// steps have no area, so they neither refer nor return.
func refer(current *domain.Referral, fromArea, toArea string, c Confirmation) (*domain.Referral, *domain.ReferralRecord, bool) {
	if fromArea == "" || toArea == "" || fromArea == toArea {
		return current, nil, false
	}
	if current != nil && toArea == current.From {
		return nil, nil, true
	}
	record := &domain.ReferralRecord{From: fromArea, To: toArea, ExternalFileNumber: c.ExternalFileNumber}
	if current != nil {
		return &domain.Referral{From: current.From, To: toArea, Since: current.Since}, record, false
	}
	return &domain.Referral{From: fromArea, To: toArea, Since: c.At}, record, false
}
