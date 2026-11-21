// Package submission is what the applicant submits to start a procedure: the
// fields and documents of the opening steps, the steps of the applicant before
// the first step of the staff.
package submission

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"backend/src/procedure/domain/definition"
)

var ErrInvalid = errors.New("invalid submission")

// StepFieldRequirement is a field an opening step asks for, with that step.
type StepFieldRequirement struct {
	definition.FieldRequirement
	Step string
}

// StepDocumentRequirement is a document an opening step asks for, with that
// step.
type StepDocumentRequirement struct {
	definition.DocumentRequirement
	Step string
}

// OpeningSteps follows the steps from the first one while they are of the
// applicant. Steps of the applicant do not decide, so they have one next step
// at most, and a valid definition has no loops.
func OpeningSteps(def definition.ProcedureTypeDefinition) []definition.Step {
	if len(def.Steps) == 0 {
		return nil
	}
	byID := make(map[string]definition.Step, len(def.Steps))
	for _, s := range def.Steps {
		byID[s.ID] = s
	}
	var opening []definition.Step
	// The bound on the length only matters for a definition with a loop.
	for s, ok := def.Steps[0], true; ok && s.Role == definition.RoleApplicant && len(opening) < len(def.Steps); s, ok = byID[s.Next] {
		opening = append(opening, s)
	}
	return opening
}

// OpenStep is the step a procedure waits on once submitted: the one after the
// opening steps, or the first step if the applicant does not open it. It is
// empty if the opening steps are all the procedure has.
func OpenStep(def definition.ProcedureTypeDefinition) string {
	opening := OpeningSteps(def)
	if len(opening) == 0 {
		if len(def.Steps) == 0 {
			return ""
		}
		return def.Steps[0].ID
	}
	return opening[len(opening)-1].Next
}

// Requirements returns what the opening steps ask for, in the order the
// definition declares it. Something mandatory in one opening step is
// mandatory, even if another makes it optional, and it belongs to the first
// step that asks for it.
func Requirements(def definition.ProcedureTypeDefinition) ([]StepFieldRequirement, []StepDocumentRequirement) {
	fields, documents := askedBy{}, askedBy{}
	for _, s := range OpeningSteps(def) {
		fields.add(s.ID, s.Fields, s.OptionalFields)
		documents.add(s.ID, s.Documents, s.OptionalDocuments)
	}
	var fieldRequirements []StepFieldRequirement
	for _, f := range def.Fields {
		if a, ok := fields[f.ID]; ok {
			fieldRequirements = append(fieldRequirements, StepFieldRequirement{
				FieldRequirement: definition.FieldRequirement{Field: f, Required: a.required},
				Step:             a.step,
			})
		}
	}
	var documentRequirements []StepDocumentRequirement
	for _, d := range def.Documents {
		if a, ok := documents[d.ID]; ok {
			documentRequirements = append(documentRequirements, StepDocumentRequirement{
				DocumentRequirement: definition.DocumentRequirement{Document: d, Required: a.required},
				Step:                a.step,
			})
		}
	}
	return fieldRequirements, documentRequirements
}

type asked struct {
	step     string
	required bool
}

type askedBy map[string]asked

func (a askedBy) add(step string, mandatory, optional []string) {
	for _, id := range slices.Concat(mandatory, optional) {
		if _, ok := a[id]; !ok {
			a[id] = asked{step: step}
		}
	}
	for _, id := range mandatory {
		a[id] = asked{step: a[id].step, required: true}
	}
}

// ProblemCode names a kind of problem. Codes are stable: the interface shows
// its own text for each one, next to the field or document it names.
type ProblemCode string

const (
	CodeMissing         ProblemCode = "missing"
	CodeInvalid         ProblemCode = "invalid"
	CodeNotAsked        ProblemCode = "not_asked"
	CodeRepeated        ProblemCode = "repeated"
	CodeTypeNotAccepted ProblemCode = "type_not_accepted"
	CodeTooLarge        ProblemCode = "too_large"
	CodeEmpty           ProblemCode = "empty"
)

// Problem is one thing wrong in a submission, about a field or a document.
type Problem struct {
	Code ProblemCode
	Kind definition.ElementKind
	ID   string
}

func (p Problem) String() string {
	return fmt.Sprintf("%s %q: %s", p.Kind, p.ID, p.Code)
}

func FieldProblem(code ProblemCode, id string) Problem {
	return Problem{Code: code, Kind: definition.KindField, ID: id}
}

func DocumentProblem(code ProblemCode, id string) Problem {
	return Problem{Code: code, Kind: definition.KindDocument, ID: id}
}

// Error is what a rejected submission returns: errors.Is matches ErrInvalid.
type Error struct {
	Problems []Problem
}

func (e *Error) Error() string {
	texts := make([]string, len(e.Problems))
	for i, p := range e.Problems {
		texts[i] = p.String()
	}
	return ErrInvalid.Error() + ": " + strings.Join(texts, "; ")
}

func (e *Error) Unwrap() error { return ErrInvalid }

// Reject returns nil when there are no problems.
func Reject(problems []Problem) error {
	if len(problems) == 0 {
		return nil
	}
	return &Error{Problems: problems}
}

// MissingDocuments returns a problem for each mandatory document that was not
// sent.
func MissingDocuments(documents []StepDocumentRequirement, sent map[string]bool) []Problem {
	var problems []Problem
	for _, d := range documents {
		if d.Required && !sent[d.Document.ID] {
			problems = append(problems, DocumentProblem(CodeMissing, d.Document.ID))
		}
	}
	return problems
}

// Fields leaves out the step of each requirement.
func Fields(requirements []StepFieldRequirement) []definition.FieldRequirement {
	fields := make([]definition.FieldRequirement, len(requirements))
	for i, r := range requirements {
		fields[i] = r.FieldRequirement
	}
	return fields
}

// Documents leaves out the step of each requirement.
func Documents(requirements []StepDocumentRequirement) []definition.DocumentRequirement {
	documents := make([]definition.DocumentRequirement, len(requirements))
	for i, r := range requirements {
		documents[i] = r.DocumentRequirement
	}
	return documents
}
