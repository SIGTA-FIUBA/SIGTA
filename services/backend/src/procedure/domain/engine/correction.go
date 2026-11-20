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
	ErrInvalidCorrectionRequest  = errors.New("invalid correction request")
	ErrNoReason                  = fmt.Errorf("%w: a reason is required", ErrInvalidCorrectionRequest)
	ErrNothingRequested          = fmt.Errorf("%w: it asks for no field or document", ErrInvalidCorrectionRequest)
	ErrNotAnApplicantRequirement = fmt.Errorf("%w: it asks for a field or document no step of the applicant asks for", ErrInvalidCorrectionRequest)
	ErrRepeatedRequirement       = fmt.Errorf("%w: it asks for the same field or document twice", ErrInvalidCorrectionRequest)
	ErrCorrectionOnApplicant     = fmt.Errorf("%w: the open step is the applicant's", ErrInvalidCorrectionRequest)
)

// RequestCorrection asks the applicant, on the open step of staff, for some
// fields and documents that their own steps ask for, and says why. The procedure
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
	fields, documents := applicantRequirements(d)
	if err := requirements(r.Fields, fields); err != nil {
		return domain.CorrectionRequest{}, domain.Event{}, err
	}
	if err := requirements(r.Documents, documents); err != nil {
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

// applicantRequirements returns the ids of the fields and documents the steps
// of the applicant ask for, so that what they send back lands on one of their
// steps.
func applicantRequirements(d definition.ProcedureTypeDefinition) (fields, documents []string) {
	for _, s := range d.Steps {
		if s.Role == definition.RoleApplicant {
			fields = slices.Concat(fields, s.Fields, s.OptionalFields)
			documents = slices.Concat(documents, s.Documents, s.OptionalDocuments)
		}
	}
	return fields, documents
}

// requirements checks that each id asked for is one of the applicant's, once.
func requirements(asked, applicant []string) error {
	seen := map[string]bool{}
	for _, a := range asked {
		if !slices.Contains(applicant, a) {
			return ErrNotAnApplicantRequirement
		}
		if seen[a] {
			return ErrRepeatedRequirement
		}
		seen[a] = true
	}
	return nil
}
