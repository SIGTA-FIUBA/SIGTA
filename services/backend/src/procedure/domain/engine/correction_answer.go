package engine

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
)

var (
	ErrNoCorrectionPending = errors.New("the applicant has no correction to make")
	ErrNotRequested        = errors.New("no pending correction asks for it")
	ErrCorrectionsChanged  = errors.New("the pending corrections changed while answering them")
)

// MissingAnswerError names what the pending corrections ask for and the
// applicant has not loaded since they were asked.
type MissingAnswerError struct {
	Fields    []string
	Documents []string
}

func (e *MissingAnswerError) Error() string {
	return fmt.Sprintf("the answer misses fields %v and documents %v", e.Fields, e.Documents)
}

// CorrectionField returns the step of the applicant that asks for a field a
// pending correction names: what they enter in answer lands there, where they
// see it (ADR-026).
func CorrectionField(p domain.Procedure, d definition.ProcedureTypeDefinition, field string) (definition.Step, error) {
	return correctionStep(p, d, field,
		func(r domain.CorrectionRequest) []string { return r.Fields },
		func(s definition.Step) []string { return slices.Concat(s.Fields, s.OptionalFields) })
}

// CorrectionDocument is CorrectionField for a document.
func CorrectionDocument(p domain.Procedure, d definition.ProcedureTypeDefinition, document string) (definition.Step, error) {
	return correctionStep(p, d, document,
		func(r domain.CorrectionRequest) []string { return r.Documents },
		func(s definition.Step) []string { return slices.Concat(s.Documents, s.OptionalDocuments) })
}

func correctionStep(p domain.Procedure, d definition.ProcedureTypeDefinition, id string, asked func(domain.CorrectionRequest) []string, asks func(definition.Step) []string) (definition.Step, error) {
	if !p.AwaitingApplicant() {
		return definition.Step{}, ErrNoCorrectionPending
	}
	if !slices.ContainsFunc(p.CorrectionRequests, func(r domain.CorrectionRequest) bool { return slices.Contains(asked(r), id) }) {
		return definition.Step{}, ErrNotRequested
	}
	for _, s := range d.Steps {
		if s.Role == definition.RoleApplicant && slices.Contains(asks(s), id) {
			return s, nil
		}
	}
	// RequestCorrection only asks for what a step of the applicant asks for.
	return definition.Step{}, fmt.Errorf("procedure %s: no step of the applicant asks for %q", p.ID, id)
}

// AnswerCorrection sends what the pending corrections asked for. Each field
// and document they name has to have been loaded again by the applicant after
// the last request that asks for it, as the history records it: the order of the history, not
// the time, says what came after. The procedure goes back to staff on the same
// open step, without pending corrections.
func AnswerCorrection(p domain.Procedure, history []domain.Event, by string, at time.Time) (domain.Procedure, domain.Event, error) {
	if !p.AwaitingApplicant() {
		return p, domain.Event{}, ErrNoCorrectionPending
	}
	asked := domain.CorrectionRecord{}
	for _, r := range p.CorrectionRequests {
		asked.Fields = appendNew(asked.Fields, r.Fields...)
		asked.Documents = appendNew(asked.Documents, r.Documents...)
	}
	missing := &MissingAnswerError{}
	for _, f := range asked.Fields {
		if !loadedAfterRequest(history, func(e domain.Event) bool { return e.By == p.Applicant && e.Value != nil && e.Value.Field == f },
			func(r *domain.CorrectionRecord) bool { return slices.Contains(r.Fields, f) }) {
			missing.Fields = append(missing.Fields, f)
		}
	}
	for _, doc := range asked.Documents {
		if !loadedAfterRequest(history, func(e domain.Event) bool {
			return e.By == p.Applicant && e.Document != nil && e.Document.Requirement == doc
		},
			func(r *domain.CorrectionRecord) bool { return slices.Contains(r.Documents, doc) }) {
			missing.Documents = append(missing.Documents, doc)
		}
	}
	if len(missing.Fields)+len(missing.Documents) > 0 {
		return p, domain.Event{}, missing
	}
	event := domain.Event{ProcedureID: p.ID, Step: p.OpenStep, By: by, At: at, CorrectionAnswer: &asked}
	p.CorrectionRequests = nil
	return p, event, nil
}

// loadedAfterRequest tells whether an event that loads the requirement comes
// after the last request that asks for it.
func loadedAfterRequest(history []domain.Event, loads func(domain.Event) bool, asks func(*domain.CorrectionRecord) bool) bool {
	loaded := false
	for _, e := range history {
		switch {
		case e.CorrectionRequest != nil && asks(e.CorrectionRequest):
			loaded = false
		case loads(e):
			loaded = true
		}
	}
	return loaded
}

func appendNew(list []string, ids ...string) []string {
	for _, id := range ids {
		if !slices.Contains(list, id) {
			list = append(list, id)
		}
	}
	return list
}
