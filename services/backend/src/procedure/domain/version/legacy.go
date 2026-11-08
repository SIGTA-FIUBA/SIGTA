package version

import (
	"fmt"

	"backend/src/procedure/domain/definition"
)

// legacyContent is the shape of formats 1 and 2, which are never rewritten:
// the applicant was an area of kind "applicant", and every other kind
// ("office", or "external" before ADR-020) was staff.
type legacyContent struct {
	Format     int                                `json:"format"`
	Definition definition.ProcedureTypeDefinition `json:"definition"`
	Areas      []legacyArea                       `json:"areas,omitempty"`
}

type legacyArea struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

const legacyApplicantKind = "applicant"

// convert gives each step its role and keeps only the staff areas. The result
// is current content, so that publishing the same definition again does not
// add a version that changes nothing.
func (l legacyContent) convert() (Content, error) {
	applicant := map[string]bool{}
	c := Content{Format: ContentFormat, Definition: l.Definition}
	for _, a := range l.Areas {
		if a.Kind == legacyApplicantKind {
			applicant[a.ID] = true
			continue
		}
		c.Areas = append(c.Areas, definition.Area{ID: a.ID, Name: a.Name})
	}
	steps := make([]definition.Step, len(l.Definition.Steps))
	for i, s := range l.Definition.Steps {
		if s.Role != "" {
			return Content{}, fmt.Errorf("decode content: format %d has no role, found %q in step %q", l.Format, s.Role, s.ID)
		}
		if applicant[s.Area] {
			s.Role, s.Area = definition.RoleApplicant, ""
		} else {
			s.Role = definition.RoleStaff
		}
		steps[i] = s
	}
	c.Definition.Steps = steps
	return c, nil
}
