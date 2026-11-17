package detail

import (
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

// Step is a step of a procedure as its detail shows it: who does it, what
// fields and documents it asks for and the values and files that cover them.
// Its outcomes are those of Step.
type Step struct {
	Step definition.Step
	// Area is nil on the steps of the applicant.
	Area      *Area
	Fields    []definition.FieldRequirement
	Documents []definition.DocumentRequirement
	// Values holds the last value entered for each field that has one, in
	// the order of Fields.
	Values []domain.FieldValue
	// Files holds the current file of each document that has one, in the
	// order of Documents.
	Files []domain.Document
}

type Area struct {
	ID   string
	Name string
}

// Steps returns the steps of a version in the order it declares them, which
// is not the path a procedure follows when a step goes back to an earlier
// one. Each step lists its fields and documents in the order the definition
// declares them; one the step asks for both ways is mandatory.
func Steps(c version.Content) []Step {
	steps := make([]Step, 0, len(c.Definition.Steps))
	for _, s := range c.Definition.Steps {
		step := Step{Step: s}
		if s.Area != "" {
			area := areaOf(c, s.Area)
			step.Area = &area
		}
		fields := asked(s.Fields, s.OptionalFields)
		for _, f := range c.Definition.Fields {
			if r, ok := fields[f.ID]; ok {
				step.Fields = append(step.Fields, definition.FieldRequirement{Field: f, Required: r})
			}
		}
		documents := asked(s.Documents, s.OptionalDocuments)
		for _, d := range c.Definition.Documents {
			if r, ok := documents[d.ID]; ok {
				step.Documents = append(step.Documents, definition.DocumentRequirement{Document: d, Required: r})
			}
		}
		steps = append(steps, step)
	}
	return steps
}

// asked says, for each id a step asks for, whether it is mandatory.
func asked(mandatory, optional []string) map[string]bool {
	required := map[string]bool{}
	for _, id := range optional {
		required[id] = false
	}
	for _, id := range mandatory {
		required[id] = true
	}
	return required
}

// WithFiles puts next to each document of the steps the file that covers it
// now, out of the documents of the procedure in the order they were attached.
func WithFiles(steps []Step, attached []domain.Document) []Step {
	current := map[string]domain.Document{}
	for _, d := range domain.Current(attached) {
		current[d.Requirement] = d
	}
	for i := range steps {
		steps[i].Files = nil
		for _, r := range steps[i].Documents {
			if d, ok := current[r.Document.ID]; ok {
				steps[i].Files = append(steps[i].Files, d)
			}
		}
	}
	return steps
}

// WithValues puts next to each field of the steps the last value entered for
// it, out of the values of the procedure in the order they were recorded.
func WithValues(steps []Step, entered []domain.FieldValue) []Step {
	last := map[string]domain.FieldValue{}
	for _, v := range entered {
		last[v.Field] = v
	}
	for i := range steps {
		steps[i].Values = nil
		for _, r := range steps[i].Fields {
			if v, ok := last[r.Field.ID]; ok {
				steps[i].Values = append(steps[i].Values, v)
			}
		}
	}
	return steps
}

// areaOf names an area as the version keeps it, or by its ID if the version
// does not have it.
func areaOf(c version.Content, id string) Area {
	for _, a := range c.Areas {
		if a.ID == id && a.Name != "" {
			return Area{ID: id, Name: a.Name}
		}
	}
	return Area{ID: id, Name: id}
}

type Referral struct {
	Area              Area
	From              Area
	DaysWithoutReturn int
}

// ReferralOf is nil when the procedure is not referred.
func ReferralOf(p domain.Procedure, c version.Content, now time.Time) *Referral {
	days, referred := p.DaysWithoutReturn(now)
	if !referred {
		return nil
	}
	return &Referral{Area: areaOf(c, p.Referral.To), From: areaOf(c, p.Referral.From), DaysWithoutReturn: days}
}
