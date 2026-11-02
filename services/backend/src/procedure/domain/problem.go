package domain

import (
	"fmt"
	"strings"
)

// Code names a kind of problem. Codes are stable: an editor shows its own text
// for each one, in any language, and places the problem by its Path.
type Code string

const (
	CodeNoID                   Code = "no_id"
	CodeInvalidID              Code = "invalid_id"
	CodeNoName                 Code = "no_name"
	CodeNameWithSpaces         Code = "name_with_spaces"
	CodeDeclaredTwice          Code = "declared_twice"
	CodeRepeatedName           Code = "repeated_name"
	CodeNameInUse              Code = "name_in_use"
	CodeNoSteps                Code = "no_steps"
	CodeUnusedField            Code = "unused_field"
	CodeUnusedDocument         Code = "unused_document"
	CodeUnreachableStep        Code = "unreachable_step"
	CodeStepNeverCloses        Code = "step_never_closes"
	CodeUnknownArea            Code = "unknown_area"
	CodeUnknownAreaKind        Code = "unknown_area_kind"
	CodeApplicantAreas         Code = "applicant_areas"
	CodeApplicantWithProfile   Code = "applicant_with_profile"
	CodeApplicantAction        Code = "applicant_action"
	CodeInvalidProfile         Code = "invalid_profile"
	CodeUnknownAction          Code = "unknown_action"
	CodeUnknownField           Code = "unknown_field"
	CodeUnknownDocument        Code = "unknown_document"
	CodeRepeatedField          Code = "repeated_field"
	CodeRepeatedDocument       Code = "repeated_document"
	CodeEnterDataNoFields      Code = "enter_data_no_fields"
	CodeEnterDataWithDocuments Code = "enter_data_with_documents"
	CodeUploadNoDocuments      Code = "upload_no_documents"
	CodeUploadWithFields       Code = "upload_with_fields"
	CodeConfirmWithDocuments   Code = "confirm_with_documents"
	CodeDecideWithNext         Code = "decide_with_next"
	CodeDecideFewOutcomes      Code = "decide_few_outcomes"
	CodeOutcomesOutsideDecide  Code = "outcomes_outside_decide"
	CodeUnknownNextStep        Code = "unknown_next_step"
	CodeOutcomeLoops           Code = "outcome_loops"
	CodeUnknownFieldType       Code = "unknown_field_type"
	CodeOptionsOutsideOption   Code = "options_outside_option"
	CodeOptionFieldNoOptions   Code = "option_field_no_options"
)

type Kind string

const (
	KindDefinition Kind = "definition"
	KindArea       Kind = "area"
	KindField      Kind = "field"
	KindOption     Kind = "option"
	KindDocument   Kind = "document"
	KindStep       Kind = "step"
	KindOutcome    Kind = "outcome"
)

// Ref points at one element of a definition. An element that has no id, or
// whose id is not valid, has an empty ID.
type Ref struct {
	Kind Kind
	ID   string
}

func (r Ref) String() string {
	if r.ID == "" {
		return string(r.Kind)
	}
	return fmt.Sprintf("%s %q", r.Kind, r.ID)
}

// Problem is one thing wrong in a definition: what (Code), where (Path, from
// the outermost element to the innermost) and the values involved.
type Problem struct {
	Code  Code
	Path  []Ref
	Value string
	Other string
}

// String is the English text of the problem, for logs and tests.
func (p Problem) String() string {
	text, aboutLast := p.text()
	path := p.Path
	if aboutLast && len(path) > 0 {
		path = path[:len(path)-1]
	}
	parts := make([]string, 0, len(path)+1)
	for _, ref := range path {
		parts = append(parts, ref.String())
	}
	return strings.Join(append(parts, text), ": ")
}

// text says what is wrong. When aboutLast is true the text already names the
// last element of the path, so String leaves it out of the prefix.
func (p Problem) text() (text string, aboutLast bool) {
	var last Ref
	if len(p.Path) > 0 {
		last = p.Path[len(p.Path)-1]
	}
	switch p.Code {
	case CodeNoID:
		return fmt.Sprintf("%s has no id", last), true
	case CodeInvalidID:
		return fmt.Sprintf("%s id %q must be lowercase ASCII letters, digits and underscores, starting with a letter", last, p.Value), true
	case CodeNoName:
		return fmt.Sprintf("%s has no name", last), true
	case CodeNameWithSpaces:
		return fmt.Sprintf("%s name %q starts or ends with spaces", last, p.Value), true
	case CodeDeclaredTwice:
		return fmt.Sprintf("%s is declared twice", last), true
	case CodeRepeatedName:
		return fmt.Sprintf("%s name %q is listed twice", last, p.Value), true
	case CodeNoSteps:
		return fmt.Sprintf("%s has no steps", last), true
	case CodeUnusedField, CodeUnusedDocument:
		return fmt.Sprintf("%s is not used by any step", last), true
	case CodeUnreachableStep:
		return fmt.Sprintf("%s is unreachable from %q", last, p.Value), true
	case CodeStepNeverCloses:
		return fmt.Sprintf("%s never reaches a closing step", last), true
	case CodeOutcomeLoops:
		return fmt.Sprintf("%s goes back to its own step", last), true
	case CodeNameInUse:
		return fmt.Sprintf("name %q is already used by area %q", p.Value, p.Other), false
	case CodeUnknownArea:
		return fmt.Sprintf("unknown area %q", p.Value), false
	case CodeUnknownAreaKind:
		return fmt.Sprintf("unknown kind %q", p.Value), false
	case CodeApplicantAreas:
		return fmt.Sprintf("areas must have one applicant area, found %s", p.Value), false
	case CodeApplicantWithProfile:
		return "steps of the applicant have no profile", false
	case CodeApplicantAction:
		return "the applicant only enters data or uploads documents", false
	case CodeInvalidProfile:
		return fmt.Sprintf("profile %q must be staff_operative or staff_approver", p.Value), false
	case CodeUnknownAction:
		return fmt.Sprintf("unknown action %q", p.Value), false
	case CodeUnknownField:
		return fmt.Sprintf("unknown field %q", p.Value), false
	case CodeUnknownDocument:
		return fmt.Sprintf("unknown document %q", p.Value), false
	case CodeRepeatedField:
		return fmt.Sprintf("field %q is listed twice", p.Value), false
	case CodeRepeatedDocument:
		return fmt.Sprintf("document %q is listed twice", p.Value), false
	case CodeEnterDataNoFields:
		return "enter_data needs at least one field", false
	case CodeEnterDataWithDocuments:
		return "enter_data takes no documents", false
	case CodeUploadNoDocuments:
		return "upload_document needs at least one document", false
	case CodeUploadWithFields:
		return "upload_document takes no fields", false
	case CodeConfirmWithDocuments:
		return "confirm_external_fact takes no documents", false
	case CodeDecideWithNext:
		return "decide goes on through its outcomes, not next", false
	case CodeDecideFewOutcomes:
		return "decide needs at least two outcomes", false
	case CodeOutcomesOutsideDecide:
		return "only decide has outcomes", false
	case CodeUnknownNextStep:
		return fmt.Sprintf("unknown next step %q", p.Value), false
	case CodeUnknownFieldType:
		return fmt.Sprintf("unknown type %q", p.Value), false
	case CodeOptionsOutsideOption:
		return "only option fields have options", false
	case CodeOptionFieldNoOptions:
		return "option fields need at least one option", false
	}
	return "unknown problem " + string(p.Code), false
}

func within(parent Ref, problems []Problem) []Problem {
	for i := range problems {
		problems[i].Path = append([]Ref{parent}, problems[i].Path...)
	}
	return problems
}
