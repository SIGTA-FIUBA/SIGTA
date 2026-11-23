package definition

import (
	"maps"
	"slices"
	"strings"
)

// ProcedureTypeDefinition describes a procedure type. A procedure starts at
// the first step, and a step without next or outcomes closes it. Staff areas
// are shared by every procedure type, while fields and documents belong to the one
// that declares them.
type ProcedureTypeDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Deadlines and Admissibility are text for the applicant to read: nothing
	// enforces them yet.
	Deadlines     []string   `json:"deadlines,omitempty"`
	Admissibility []string   `json:"admissibility,omitempty"`
	Fields        []Field    `json:"fields,omitempty"`
	Documents     []Document `json:"documents,omitempty"`
	Steps         []Step     `json:"steps"`
}

func (d ProcedureTypeDefinition) Validate(areas []Area) error {
	if problems := validateAreas(areas); len(problems) > 0 {
		return &ValidationError{Cause: ErrInvalidAreas, Problems: problems}
	}
	if problems := d.problems(areas); len(problems) > 0 {
		return &ValidationError{Cause: ErrInvalidDefinition, Problems: problems}
	}
	return nil
}

// validateTexts checks what the catalog shows besides the name.
func validateTexts(d ProcedureTypeDefinition) []ValidationProblem {
	var problems []ValidationProblem
	if strings.TrimSpace(d.Description) == "" {
		problems = append(problems, ValidationProblem{Code: CodeNoDescription, Path: []Element{{Kind: KindDefinition}}})
	}
	for i, text := range d.Deadlines {
		if strings.TrimSpace(text) == "" {
			problems = append(problems, ValidationProblem{Code: CodeBlankText, Path: []Element{{Kind: KindDeadline, Position: i + 1}}})
		}
	}
	for i, text := range d.Admissibility {
		if strings.TrimSpace(text) == "" {
			problems = append(problems, ValidationProblem{Code: CodeBlankText, Path: []Element{{Kind: KindAdmissibility, Position: i + 1}}})
		}
	}
	return problems
}

func (d ProcedureTypeDefinition) problems(areas []Area) []ValidationProblem {
	problems := validateName(d.Name, Element{Kind: KindDefinition})
	problems = append(problems, validateTexts(d)...)
	fields, documents := map[string]bool{}, map[string]bool{}
	problems = append(problems, validateFields(d.Fields, fields)...)
	problems = append(problems, validateDocuments(d.Documents, documents)...)
	if len(d.Steps) == 0 {
		return append(problems, ValidationProblem{Code: CodeNoSteps, Path: []Element{{Kind: KindDefinition}}})
	}

	areasByID := map[string]Area{}
	for _, a := range areas {
		areasByID[a.ID] = a
	}

	steps := map[string]Step{}
	seen := map[string]bool{}
	for _, s := range d.Steps {
		problems = append(problems, validateEntry(KindStep, s.ID, s.Name, seen)...)
		steps[s.ID] = s
	}

	validator := stepValidator{areas: areasByID, fields: fields, documents: documents, steps: steps}
	for _, s := range d.Steps {
		problems = append(problems, within(Element{Kind: KindStep, ID: s.ID}, validator.validate(s))...)
	}
	problems = append(problems, unused(d)...)
	problems = append(problems, resolutionsFromApplicant(d)...)
	if len(problems) > 0 {
		return problems
	}
	return validateFlow(d.Steps)
}

// unused reports fields and documents that no step asks for: they would be
// stored without ever being shown.
func unused(d ProcedureTypeDefinition) []ValidationProblem {
	usedFields := map[string]bool{}
	usedDocuments := map[string]bool{}
	for _, s := range d.Steps {
		for _, id := range slices.Concat(s.Fields, s.OptionalFields) {
			usedFields[id] = true
		}
		for _, id := range slices.Concat(s.Documents, s.OptionalDocuments) {
			usedDocuments[id] = true
		}
	}
	var problems []ValidationProblem
	for _, f := range d.Fields {
		if !usedFields[f.ID] {
			problems = append(problems, ValidationProblem{Code: CodeUnusedField, Path: []Element{{Kind: KindField, ID: f.ID}}})
		}
	}
	for _, doc := range d.Documents {
		if !usedDocuments[doc.ID] {
			problems = append(problems, ValidationProblem{Code: CodeUnusedDocument, Path: []Element{{Kind: KindDocument, ID: doc.ID}}})
		}
	}
	return problems
}

// resolutionsFromApplicant reports a resolution that a step of the applicant
// asks for: the resolution is what staff attaches for the applicant to
// download once the procedure closes.
func resolutionsFromApplicant(d ProcedureTypeDefinition) []ValidationProblem {
	var problems []ValidationProblem
	for _, s := range d.Steps {
		if s.Role != RoleApplicant {
			continue
		}
		for _, id := range slices.Concat(s.Documents, s.OptionalDocuments) {
			if slices.ContainsFunc(d.Documents, func(doc Document) bool { return doc.ID == id && doc.Resolution }) {
				problems = append(problems, ValidationProblem{Code: CodeResolutionFromApplicant, Path: []Element{{Kind: KindStep, ID: s.ID}, {Kind: KindDocument, ID: id}}})
			}
		}
	}
	return problems
}

func validateFlow(steps []Step) []ValidationProblem {
	next := map[string][]string{}
	previous := map[string][]string{}
	var closing []string
	for _, s := range steps {
		successors := s.successors()
		next[s.ID] = successors
		for _, n := range successors {
			previous[n] = append(previous[n], s.ID)
		}
		if len(successors) == 0 {
			closing = append(closing, s.ID)
		}
	}

	reached := walk([]string{steps[0].ID}, next)
	closes := walk(closing, previous)
	var problems []ValidationProblem
	for _, s := range steps {
		if !reached[s.ID] {
			problems = append(problems, ValidationProblem{Code: CodeUnreachableStep, Path: []Element{{Kind: KindStep, ID: s.ID}}, Value: steps[0].ID})
		}
		if !closes[s.ID] {
			problems = append(problems, ValidationProblem{Code: CodeStepNeverCloses, Path: []Element{{Kind: KindStep, ID: s.ID}}})
		}
		if backwards := backwardsFrom(s, steps[0].ID, next); backwards != nil {
			for _, o := range s.Outcomes {
				if o.SendBack && !backwards(o.Next) {
					problems = append(problems, ValidationProblem{Code: CodeSendBackForward, Path: []Element{{Kind: KindStep, ID: s.ID}, {Kind: KindOutcome, ID: o.ID}}})
				}
			}
		}
	}
	return problems
}

// backwardsFrom tells whether a step is earlier than s: reached from the first
// step without going through s, and leading back to s. It is nil if s has no
// send backs.
func backwardsFrom(s Step, first string, next map[string][]string) func(id string) bool {
	if !slices.ContainsFunc(s.Outcomes, func(o Outcome) bool { return o.SendBack }) {
		return nil
	}
	around := maps.Clone(next)
	around[s.ID] = nil
	before := walk([]string{first}, around)
	return func(id string) bool {
		return id != s.ID && before[id] && walk([]string{id}, next)[s.ID]
	}
}

func walk(from []string, edges map[string][]string) map[string]bool {
	seen := map[string]bool{}
	pending := slices.Clone(from)
	for len(pending) > 0 {
		id := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if seen[id] {
			continue
		}
		seen[id] = true
		pending = append(pending, edges[id]...)
	}
	return seen
}
