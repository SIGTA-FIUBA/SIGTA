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
)

type Confirmation struct {
	Step               string
	By                 string
	At                 time.Time
	Observation        string
	Outcome            string
	ExternalFileNumber string
}

type Event struct {
	ProcedureID string
	Step        string
	Action      definition.Action
	By          string
	At          time.Time
	Observation string
	Outcome     string
	Declared    bool
	Next        string
	Referral    *ReferralRecord
	Returned    bool
}

type ReferralRecord struct {
	From               string
	To                 string
	ExternalFileNumber string
}

func Confirm(p domain.Procedure, d definition.ProcedureTypeDefinition, c Confirmation) (domain.Procedure, Event, error) {
	if p.Closed() || c.Step != p.OpenStep {
		return p, Event{}, ErrStepNotOpen
	}
	step, ok := find(d.Steps, c.Step)
	if !ok {
		return p, Event{}, fmt.Errorf("procedure %s: open step %q is not in its version", p.ID, c.Step)
	}
	if step.Role == definition.RoleApplicant {
		return p, Event{}, ErrApplicantStep
	}
	if strings.TrimSpace(c.Observation) == "" {
		return p, Event{}, ErrNoObservation
	}
	nextID, err := next(step, c.Outcome)
	if err != nil {
		return p, Event{}, err
	}

	event := Event{
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
	if nextID == "" {
		// Closing ends the referral too: a closed procedure is nowhere.
		moved.Referral = nil
	} else {
		nextStep, ok := find(d.Steps, nextID)
		if !ok {
			return p, Event{}, fmt.Errorf("procedure %s: step %q leads to %q, which is not in its version", p.ID, step.ID, nextID)
		}
		moved.Referral, event.Referral, event.Returned = refer(p.Referral, step.Area, nextStep.Area, c)
	}
	if c.ExternalFileNumber != "" && event.Referral == nil {
		return p, Event{}, ErrNotAReferral
	}
	return moved, event, nil
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

// refer follows ADR-020: moving to a step of another area refers the
// procedure, and moving back to the area it first left returns it. Applicant
// steps have no area, so they neither refer nor return.
func refer(current *domain.Referral, fromArea, toArea string, c Confirmation) (*domain.Referral, *ReferralRecord, bool) {
	if fromArea == "" || toArea == "" || fromArea == toArea {
		return current, nil, false
	}
	if current != nil && toArea == current.From {
		return nil, nil, true
	}
	record := &ReferralRecord{From: fromArea, To: toArea, ExternalFileNumber: c.ExternalFileNumber}
	if current != nil {
		return &domain.Referral{From: current.From, To: toArea, Since: current.Since}, record, false
	}
	return &domain.Referral{From: fromArea, To: toArea, Since: c.At}, record, false
}
