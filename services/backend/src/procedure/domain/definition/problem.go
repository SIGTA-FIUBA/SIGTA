package definition

import (
	"fmt"
	"strings"
)

// ProblemCode names a kind of problem. Codes are stable: an editor shows its
// own text for each one, in any language, and places the problem by its Path.
type ProblemCode string

const (
	CodeNoID                   ProblemCode = "no_id"
	CodeInvalidID              ProblemCode = "invalid_id"
	CodeNoName                 ProblemCode = "no_name"
	CodeNameWithSpaces         ProblemCode = "name_with_spaces"
	CodeDeclaredTwice          ProblemCode = "declared_twice"
	CodeRepeatedName           ProblemCode = "repeated_name"
	CodeNameInUse              ProblemCode = "name_in_use"
	CodeNoSteps                ProblemCode = "no_steps"
	CodeUnusedField            ProblemCode = "unused_field"
	CodeUnusedDocument         ProblemCode = "unused_document"
	CodeUnreachableStep        ProblemCode = "unreachable_step"
	CodeStepNeverCloses        ProblemCode = "step_never_closes"
	CodeUnknownArea            ProblemCode = "unknown_area"
	CodeUnknownAreaKind        ProblemCode = "unknown_area_kind"
	CodeApplicantAreas         ProblemCode = "applicant_areas"
	CodeApplicantWithProfile   ProblemCode = "applicant_with_profile"
	CodeApplicantAction        ProblemCode = "applicant_action"
	CodeInvalidProfile         ProblemCode = "invalid_profile"
	CodeUnknownAction          ProblemCode = "unknown_action"
	CodeUnknownField           ProblemCode = "unknown_field"
	CodeUnknownDocument        ProblemCode = "unknown_document"
	CodeRepeatedField          ProblemCode = "repeated_field"
	CodeRepeatedDocument       ProblemCode = "repeated_document"
	CodeEnterDataNoFields      ProblemCode = "enter_data_no_fields"
	CodeEnterDataWithDocuments ProblemCode = "enter_data_with_documents"
	CodeUploadNoDocuments      ProblemCode = "upload_no_documents"
	CodeUploadWithFields       ProblemCode = "upload_with_fields"
	CodeConfirmWithDocuments   ProblemCode = "confirm_with_documents"
	CodeDecideWithNext         ProblemCode = "decide_with_next"
	CodeDecideFewOutcomes      ProblemCode = "decide_few_outcomes"
	CodeOutcomesOutsideDecide  ProblemCode = "outcomes_outside_decide"
	CodeUnknownNextStep        ProblemCode = "unknown_next_step"
	CodeOutcomeLoops           ProblemCode = "outcome_loops"
	CodeUnknownFieldType       ProblemCode = "unknown_field_type"
	CodeOptionsOutsideOption   ProblemCode = "options_outside_option"
	CodeOptionFieldNoOptions   ProblemCode = "option_field_no_options"
	CodeNoDescription          ProblemCode = "no_description"
	CodeBlankText              ProblemCode = "blank_text"
	CodeNoAcceptedTypes        ProblemCode = "no_accepted_types"
	CodeInvalidAcceptedType    ProblemCode = "invalid_accepted_type"
	CodeRepeatedAcceptedType   ProblemCode = "repeated_accepted_type"
	CodeNoMaxSize              ProblemCode = "no_max_size"
)

type ElementKind string

const (
	KindDefinition    ElementKind = "definition"
	KindArea          ElementKind = "area"
	KindField         ElementKind = "field"
	KindOption        ElementKind = "option"
	KindDocument      ElementKind = "document"
	KindStep          ElementKind = "step"
	KindOutcome       ElementKind = "outcome"
	KindDeadline      ElementKind = "deadline"
	KindAdmissibility ElementKind = "admissibility"
)

// Element points at one element of a definition. An element that has no id, or
// whose id is not valid, has an empty ID. Deadlines and admissibility
// conditions have no id, so Position says which one it is, counting from 1.
type Element struct {
	Kind     ElementKind
	ID       string
	Position int
}

func (e Element) String() string {
	switch {
	case e.ID != "":
		return fmt.Sprintf("%s %q", e.Kind, e.ID)
	case e.Position > 0:
		return fmt.Sprintf("%s %d", e.Kind, e.Position)
	default:
		return string(e.Kind)
	}
}

// ValidationProblem is one thing wrong in a definition: what (Code), where
// (Path, from the outermost element to the innermost) and the values involved.
type ValidationProblem struct {
	Code  ProblemCode
	Path  []Element
	Value string
	Other string
}

// String is the English text of the problem, for logs and tests.
func (p ValidationProblem) String() string {
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
func (p ValidationProblem) text() (text string, aboutLast bool) {
	var last Element
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
	case CodeNoDescription:
		return fmt.Sprintf("%s has no description", last), true
	case CodeBlankText:
		return fmt.Sprintf("%s has no text", last), true
	case CodeNoAcceptedTypes:
		return fmt.Sprintf("%s accepts no file types", last), true
	case CodeInvalidAcceptedType:
		return fmt.Sprintf("%s accepts %q, which is not a media type", last, p.Value), true
	case CodeRepeatedAcceptedType:
		return fmt.Sprintf("%s accepts %q twice", last, p.Value), true
	case CodeNoMaxSize:
		return fmt.Sprintf("%s has no maximum size", last), true
	}
	return "unknown problem " + string(p.Code), false
}

func within(parent Element, problems []ValidationProblem) []ValidationProblem {
	for i := range problems {
		problems[i].Path = append([]Element{parent}, problems[i].Path...)
	}
	return problems
}
