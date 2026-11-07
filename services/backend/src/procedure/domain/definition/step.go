package definition

import "slices"

type Action string

const (
	ActionEnterData           Action = "enter_data"
	ActionUploadDocument      Action = "upload_document"
	ActionConfirmExternalFact Action = "confirm_external_fact"
	ActionDecide              Action = "decide"
)

type Profile string

const (
	ProfileStaffOperative Profile = "staff_operative"
	ProfileStaffApprover  Profile = "staff_approver"
)

type Step struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Area              string    `json:"area"`
	Profile           Profile   `json:"profile,omitempty"`
	Action            Action    `json:"action"`
	Fields            []string  `json:"fields,omitempty"`
	OptionalFields    []string  `json:"optional_fields,omitempty"`
	Documents         []string  `json:"documents,omitempty"`
	OptionalDocuments []string  `json:"optional_documents,omitempty"`
	Outcomes          []Outcome `json:"outcomes,omitempty"`
	Next              string    `json:"next,omitempty"`
}

type Outcome struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Next string `json:"next"`
}

type stepValidator struct {
	areas     map[string]Area
	fields    map[string]bool
	documents map[string]bool
	steps     map[string]Step
}

func (v stepValidator) validate(s Step) []ValidationProblem {
	problems := v.owner(s)
	problems = append(problems, v.contents(s)...)
	if s.Action == ActionDecide {
		return append(problems, v.outcomes(s)...)
	}
	return append(problems, v.next(s)...)
}

func (v stepValidator) owner(s Step) []ValidationProblem {
	area, ok := v.areas[s.Area]
	switch {
	case !ok:
		return []ValidationProblem{{Code: CodeUnknownArea, Value: s.Area}}
	case area.Kind == AreaApplicant && s.Profile != "":
		return []ValidationProblem{{Code: CodeApplicantWithProfile}}
	case area.Kind == AreaApplicant && s.Action != ActionEnterData && s.Action != ActionUploadDocument:
		return []ValidationProblem{{Code: CodeApplicantAction}}
	case area.Kind != AreaApplicant && s.Profile != ProfileStaffOperative && s.Profile != ProfileStaffApprover:
		return []ValidationProblem{{Code: CodeInvalidProfile, Value: string(s.Profile)}}
	}
	return nil
}

func (v stepValidator) contents(s Step) []ValidationProblem {
	problems := validateReferences(CodeUnknownField, CodeRepeatedField, slices.Concat(s.Fields, s.OptionalFields), v.fields)
	problems = append(problems, validateReferences(CodeUnknownDocument, CodeRepeatedDocument, slices.Concat(s.Documents, s.OptionalDocuments), v.documents)...)
	hasFields := len(s.Fields)+len(s.OptionalFields) > 0
	hasDocuments := len(s.Documents)+len(s.OptionalDocuments) > 0

	switch s.Action {
	case ActionEnterData:
		if !hasFields {
			problems = append(problems, ValidationProblem{Code: CodeEnterDataNoFields})
		}
		if hasDocuments {
			problems = append(problems, ValidationProblem{Code: CodeEnterDataWithDocuments})
		}
	case ActionUploadDocument:
		if !hasDocuments {
			problems = append(problems, ValidationProblem{Code: CodeUploadNoDocuments})
		}
		if hasFields {
			problems = append(problems, ValidationProblem{Code: CodeUploadWithFields})
		}
	case ActionConfirmExternalFact:
		if hasDocuments {
			problems = append(problems, ValidationProblem{Code: CodeConfirmWithDocuments})
		}
	case ActionDecide:
		// The fields and documents of a decision are its grounds, recorded
		// together with the outcome.
	default:
		problems = append(problems, ValidationProblem{Code: CodeUnknownAction, Value: string(s.Action)})
	}
	return problems
}

func (v stepValidator) outcomes(s Step) []ValidationProblem {
	var problems []ValidationProblem
	if s.Next != "" {
		problems = append(problems, ValidationProblem{Code: CodeDecideWithNext})
	}
	if len(s.Outcomes) < 2 {
		problems = append(problems, ValidationProblem{Code: CodeDecideFewOutcomes})
	}
	seenIDs := map[string]bool{}
	seenNames := map[string]bool{}
	for _, o := range s.Outcomes {
		problems = append(problems, validateEntry(KindOutcome, o.ID, o.Name, seenIDs)...)
		problems = append(problems, repeatedName(KindOutcome, o.Name, seenNames)...)
		outcome := Element{Kind: KindOutcome, ID: o.ID}
		if _, ok := v.steps[o.Next]; !ok {
			problems = append(problems, within(outcome, []ValidationProblem{{Code: CodeUnknownNextStep, Value: o.Next}})...)
		}
		if o.Next != "" && o.Next == s.ID {
			problems = append(problems, ValidationProblem{Code: CodeOutcomeLoops, Path: []Element{outcome}})
		}
	}
	return problems
}

func (v stepValidator) next(s Step) []ValidationProblem {
	var problems []ValidationProblem
	if len(s.Outcomes) > 0 {
		problems = append(problems, ValidationProblem{Code: CodeOutcomesOutsideDecide})
	}
	if _, ok := v.steps[s.Next]; s.Next != "" && !ok {
		problems = append(problems, ValidationProblem{Code: CodeUnknownNextStep, Value: s.Next})
	}
	return problems
}

func validateReferences(unknown, repeated ProblemCode, ids []string, declared map[string]bool) []ValidationProblem {
	var problems []ValidationProblem
	seen := map[string]bool{}
	for _, id := range ids {
		if !declared[id] {
			problems = append(problems, ValidationProblem{Code: unknown, Value: id})
		}
		if seen[id] {
			problems = append(problems, ValidationProblem{Code: repeated, Value: id})
		}
		seen[id] = true
	}
	return problems
}

func (s Step) successors() []string {
	if s.Action == ActionDecide {
		ids := make([]string, 0, len(s.Outcomes))
		for _, o := range s.Outcomes {
			ids = append(ids, o.Next)
		}
		return ids
	}
	if s.Next == "" {
		return nil
	}
	return []string{s.Next}
}
