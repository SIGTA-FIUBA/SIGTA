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
	ErrNoCorrectionPending = errors.New("the applicant has no correction to make")
	ErrNotRequested        = errors.New("no pending correction asks for it")
	ErrCorrectionsChanged  = errors.New("the pending corrections changed while answering them")
	ErrMissingAnswer       = errors.New("the answer misses what a correction asks for")
)

// MissingAnswerError names what the pending corrections ask for and the
// applicant has not loaded since they were asked.
type MissingAnswerError struct {
	Fields    []definition.Field
	Documents []definition.Document
}

func (e *MissingAnswerError) Error() string {
	ids := make([]string, 0, len(e.Fields)+len(e.Documents))
	for _, f := range e.Fields {
		ids = append(ids, f.ID)
	}
	for _, d := range e.Documents {
		ids = append(ids, d.ID)
	}
	return fmt.Sprintf("%v: %s", ErrMissingAnswer, strings.Join(ids, ", "))
}

func (e *MissingAnswerError) Unwrap() error { return ErrMissingAnswer }

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
func AnswerCorrection(p domain.Procedure, d definition.ProcedureTypeDefinition, history []domain.Event, by string, at time.Time) (domain.Procedure, domain.Event, error) {
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
			missing.Fields = append(missing.Fields, declared(d.Fields, f, func(field definition.Field) string { return field.ID }, definition.Field{ID: f}))
		}
	}
	for _, doc := range asked.Documents {
		if !loadedAfterRequest(history, func(e domain.Event) bool {
			return e.By == p.Applicant && e.Document != nil && e.Document.Requirement == doc
		},
			func(r *domain.CorrectionRecord) bool { return slices.Contains(r.Documents, doc) }) {
			missing.Documents = append(missing.Documents, declared(d.Documents, doc, func(document definition.Document) string { return document.ID }, definition.Document{ID: doc}))
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

// declared returns the field or document of the definition with the id, or
// fallback, which names it only by its id.
func declared[T any](all []T, id string, idOf func(T) string, fallback T) T {
	if i := slices.IndexFunc(all, func(item T) bool { return idOf(item) == id }); i >= 0 {
		return all[i]
	}
	return fallback
}

// AttachedSinceRequest counts the files attached for a document after the last
// correction request that asks for it.
func AttachedSinceRequest(history []domain.Event, document string) int {
	n := 0
	for _, e := range history {
		switch {
		case e.CorrectionRequest != nil && slices.Contains(e.CorrectionRequest.Documents, document):
			n = 0
		case e.Document != nil && e.Document.Requirement == document:
			n++
		}
	}
	return n
}

// CanRecordAttachment says whether a document can be recorded on a procedure
// as it is stored, by its open step, its applicant and whether a correction
// is pending: on the open step or, attached by the applicant while a
// correction is pending, on a step of theirs (ADR-026). Otherwise the step is
// not open for it, also when an answer sent meanwhile left nothing pending.
// Adapters check it when storing.
func CanRecordAttachment(openStep, applicant string, awaiting bool, d domain.Document) error {
	if openStep != "" && d.Step == openStep || awaiting && d.UploadedBy == applicant {
		return nil
	}
	return ErrStepNotOpen
}

// CorrectionRequestsIn counts the correction requests a history records. It
// only grows, so a different count means a request came in between.
func CorrectionRequestsIn(history []domain.Event) int {
	n := 0
	for _, e := range history {
		if e.CorrectionRequest != nil {
			n++
		}
	}
	return n
}

func appendNew(list []string, ids ...string) []string {
	for _, id := range ids {
		if !slices.Contains(list, id) {
			list = append(list, id)
		}
	}
	return list
}
