package engine

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
)

var (
	ErrInvalidCorrectionRequest = errors.New("invalid correction request")
	ErrNoReason                 = fmt.Errorf("%w: a reason is required", ErrInvalidCorrectionRequest)
	ErrNothingRequested         = fmt.Errorf("%w: it asks for no field or document", ErrInvalidCorrectionRequest)
	ErrNotARequirement          = fmt.Errorf("%w: it asks for a field or document the procedure type does not have", ErrInvalidCorrectionRequest)
	ErrRepeatedRequirement      = fmt.Errorf("%w: it asks for the same field or document twice", ErrInvalidCorrectionRequest)
	ErrCorrectionOnApplicant    = fmt.Errorf("%w: the open step is the applicant's", ErrInvalidCorrectionRequest)
)

// RequestCorrection asks the applicant, on the open step of staff, for some
// fields and documents of the procedure type and says why. The procedure
// stays on its step, waiting for the applicant. Who may ask is checked by the
// use case, as for a confirmation.
func RequestCorrection(p domain.Procedure, d definition.ProcedureTypeDefinition, r domain.CorrectionRequest) (domain.CorrectionRequest, domain.Event, error) {
	if p.Closed() || r.Step != p.OpenStep {
		return domain.CorrectionRequest{}, domain.Event{}, ErrStepNotOpen
	}
	step, ok := find(d.Steps, r.Step)
	if !ok {
		return domain.CorrectionRequest{}, domain.Event{}, fmt.Errorf("procedure %s: open step %q is not in its version", p.ID, r.Step)
	}
	if step.Role == definition.RoleApplicant {
		return domain.CorrectionRequest{}, domain.Event{}, ErrCorrectionOnApplicant
	}
	if strings.TrimSpace(r.Reason) == "" {
		return domain.CorrectionRequest{}, domain.Event{}, ErrNoReason
	}
	if len(r.Fields)+len(r.Documents) == 0 {
		return domain.CorrectionRequest{}, domain.Event{}, ErrNothingRequested
	}
	if err := requirements(r.Fields, d.Fields, func(f definition.Field) string { return f.ID }); err != nil {
		return domain.CorrectionRequest{}, domain.Event{}, err
	}
	if err := requirements(r.Documents, d.Documents, func(doc definition.Document) string { return doc.ID }); err != nil {
		return domain.CorrectionRequest{}, domain.Event{}, err
	}
	event := domain.Event{
		ProcedureID:       p.ID,
		Step:              step.ID,
		By:                r.By,
		At:                r.At,
		Observation:       r.Reason,
		CorrectionRequest: &domain.CorrectionRecord{Fields: slices.Clone(r.Fields), Documents: slices.Clone(r.Documents)},
	}
	return r, event, nil
}

// requirements checks that each id asked for is declared once in the
// procedure type.
func requirements[T any](asked []string, declared []T, id func(T) string) error {
	seen := map[string]bool{}
	for _, a := range asked {
		if !slices.ContainsFunc(declared, func(item T) bool { return id(item) == a }) {
			return ErrNotARequirement
		}
		if seen[a] {
			return ErrRepeatedRequirement
		}
		seen[a] = true
	}
	return nil
}
