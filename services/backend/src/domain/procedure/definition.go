package procedure

import (
	"slices"
	"strings"
)

// A procedure starts at the first step, and a step without next or outcomes
// closes it. Areas are shared by every procedure type, while fields and
// documents belong to the one that declares them.
type Definition struct {
	Name string `json:"name"`
	// Description explains the procedure type in plain language.
	Description string `json:"description,omitempty"`
	// Deadlines and Admissibility are text for the applicant to read: nothing
	// enforces them yet.
	Deadlines     []string   `json:"deadlines,omitempty"`
	Admissibility []string   `json:"admissibility,omitempty"`
	Fields        []Field    `json:"fields,omitempty"`
	Documents     []Document `json:"documents,omitempty"`
	Steps         []Step     `json:"steps"`
}

func (d Definition) Validate(areas []Area) error {
	if problems := validateAreas(areas); len(problems) > 0 {
		return &InvalidError{Kind: ErrInvalidAreas, Problems: problems}
	}
	if problems := d.problems(areas); len(problems) > 0 {
		return &InvalidError{Kind: ErrInvalidDefinition, Problems: problems}
	}
	return nil
}

// validateTexts checks what the catalog shows besides the name.
func validateTexts(d Definition) []Problem {
	var problems []Problem
	if strings.TrimSpace(d.Description) == "" {
		problems = append(problems, Problem{Code: CodeNoDescription, Path: []Ref{{Kind: KindDefinition}}})
	}
	for _, text := range d.Deadlines {
		if strings.TrimSpace(text) == "" {
			problems = append(problems, Problem{Code: CodeBlankText, Path: []Ref{{Kind: KindDeadline}}})
		}
	}
	for _, text := range d.Admissibility {
		if strings.TrimSpace(text) == "" {
			problems = append(problems, Problem{Code: CodeBlankText, Path: []Ref{{Kind: KindAdmissibility}}})
		}
	}
	return problems
}

func (d Definition) problems(areas []Area) []Problem {
	errs := validateName(d.Name, Ref{Kind: KindDefinition})
	errs = append(errs, validateTexts(d)...)
	fields, documents := map[string]bool{}, map[string]bool{}
	errs = append(errs, validateFields(d.Fields, fields)...)
	errs = append(errs, validateDocuments(d.Documents, documents)...)
	if len(d.Steps) == 0 {
		return append(errs, Problem{Code: CodeNoSteps, Path: []Ref{{Kind: KindDefinition}}})
	}

	areasByID := map[string]Area{}
	for _, a := range areas {
		areasByID[a.ID] = a
	}

	steps := map[string]Step{}
	seen := map[string]bool{}
	for _, s := range d.Steps {
		errs = append(errs, validateEntry(KindStep, s.ID, s.Name, seen)...)
		steps[s.ID] = s
	}

	validator := stepValidator{areas: areasByID, fields: fields, documents: documents, steps: steps}
	for _, s := range d.Steps {
		errs = append(errs, within(Ref{Kind: KindStep, ID: s.ID}, validator.validate(s))...)
	}
	errs = append(errs, unused(d)...)
	if len(errs) > 0 {
		return errs
	}
	return validateFlow(d.Steps)
}

// unused reports fields and documents that no step asks for: they would be
// stored without ever being shown.
func unused(d Definition) []Problem {
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
	var errs []Problem
	for _, f := range d.Fields {
		if !usedFields[f.ID] {
			errs = append(errs, Problem{Code: CodeUnusedField, Path: []Ref{{Kind: KindField, ID: f.ID}}})
		}
	}
	for _, doc := range d.Documents {
		if !usedDocuments[doc.ID] {
			errs = append(errs, Problem{Code: CodeUnusedDocument, Path: []Ref{{Kind: KindDocument, ID: doc.ID}}})
		}
	}
	return errs
}

func validateFlow(steps []Step) []Problem {
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
	var errs []Problem
	for _, s := range steps {
		if !reached[s.ID] {
			errs = append(errs, Problem{Code: CodeUnreachableStep, Path: []Ref{{Kind: KindStep, ID: s.ID}}, Value: steps[0].ID})
		}
		if !closes[s.ID] {
			errs = append(errs, Problem{Code: CodeStepNeverCloses, Path: []Ref{{Kind: KindStep, ID: s.ID}}})
		}
	}
	return errs
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
