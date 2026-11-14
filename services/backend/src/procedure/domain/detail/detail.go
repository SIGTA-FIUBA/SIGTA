package detail

import (
	"backend/src/procedure/domain/catalog"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

// Step is a step of a procedure as its detail shows it: who does it and what
// documents it asks for.
type Step struct {
	Step definition.Step
	// AreaName is empty on the steps of the applicant.
	AreaName  string
	Documents []catalog.DocumentRequirement
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
				step.Documents = append(step.Documents, catalog.DocumentRequirement{Document: d, Required: r})
			}
		}
		steps = append(steps, step)
	}
	return steps
}
