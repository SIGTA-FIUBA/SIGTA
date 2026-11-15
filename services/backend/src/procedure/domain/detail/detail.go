package detail

import (
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

// Step is a step of a procedure as its detail shows it: who does it, what
// documents it asks for and the files that cover them.
type Step struct {
	Step definition.Step
	// AreaName is empty on the steps of the applicant.
	AreaName  string
	Documents []definition.DocumentRequirement
	// Files holds the current file of each document that has one, in the
	// order of Documents.
	Files []domain.Document
}

// Steps returns the steps of a version in the order it declares them. Each
// step lists its documents in the order the definition declares them; a
// document the step asks for both ways is mandatory.
func Steps(c version.Content) []Step {
	areas := map[string]string{}
	for _, a := range c.Areas {
		areas[a.ID] = a.Name
	}
	steps := make([]Step, 0, len(c.Definition.Steps))
	for _, s := range c.Definition.Steps {
		step := Step{Step: s}
		if s.Area != "" {
			step.AreaName = areas[s.Area]
			if step.AreaName == "" {
				step.AreaName = s.Area
			}
		}
		required := map[string]bool{}
		for _, id := range s.OptionalDocuments {
			required[id] = false
		}
		for _, id := range s.Documents {
			required[id] = true
		}
		for _, d := range c.Definition.Documents {
			if r, asked := required[d.ID]; asked {
				step.Documents = append(step.Documents, definition.DocumentRequirement{Document: d, Required: r})
			}
		}
		steps = append(steps, step)
	}
	return steps
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
