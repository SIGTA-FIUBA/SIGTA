package version

import (
	"cmp"
	"slices"

	"backend/src/procedure/domain/definition"
)

// legacyContent is the shape of formats 1 and 2, which are never rewritten:
// the applicant was an area of kind "applicant", every other kind ("office",
// or "external" before ADR-020) was staff, and a staff step asked for a fixed
// profile instead of permissions of its area.
type legacyContent struct {
	Format     int              `json:"format"`
	Definition legacyDefinition `json:"definition"`
	Areas      []legacyArea     `json:"areas,omitempty"`
}

// legacyDefinition and legacyStep keep the keys and their order as formats 1
// and 2 stored them.
type legacyDefinition struct {
	Name          string                `json:"name"`
	Description   string                `json:"description,omitempty"`
	Deadlines     []string              `json:"deadlines,omitempty"`
	Admissibility []string              `json:"admissibility,omitempty"`
	Fields        []definition.Field    `json:"fields,omitempty"`
	Documents     []definition.Document `json:"documents,omitempty"`
	Steps         []legacyStep          `json:"steps"`
}

type legacyStep struct {
	ID                string               `json:"id"`
	Name              string               `json:"name"`
	Area              string               `json:"area"`
	Profile           string               `json:"profile,omitempty"`
	Action            definition.Action    `json:"action"`
	Fields            []string             `json:"fields,omitempty"`
	OptionalFields    []string             `json:"optional_fields,omitempty"`
	Documents         []string             `json:"documents,omitempty"`
	OptionalDocuments []string             `json:"optional_documents,omitempty"`
	Outcomes          []definition.Outcome `json:"outcomes,omitempty"`
	Next              string               `json:"next,omitempty"`
}

type legacyArea struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

const legacyApplicantKind = "applicant"

// legacyProfileNames name the permission that each fixed profile becomes. The
// permission keeps the profile as its ID, so that someone can still do the
// step once the area has a permission with that ID.
var legacyProfileNames = map[string]string{
	"staff_operative": "Operativo",
	"staff_approver":  "Aprobación",
}

// convert gives each step its role, turns the profile of a staff step into a
// permission of its area, and keeps only the staff areas. The result is
// current content, so that publishing the same definition again does not add
// a version that changes nothing.
func (l legacyContent) convert() Content {
	applicant := map[string]bool{}
	asked := map[string]map[string]bool{}
	d := definition.ProcedureTypeDefinition{
		Name:          l.Definition.Name,
		Description:   l.Definition.Description,
		Deadlines:     l.Definition.Deadlines,
		Admissibility: l.Definition.Admissibility,
		Fields:        l.Definition.Fields,
		Documents:     l.Definition.Documents,
	}
	for _, a := range l.Areas {
		if a.Kind == legacyApplicantKind {
			applicant[a.ID] = true
		}
	}
	for _, s := range l.Definition.Steps {
		step := definition.Step{
			ID: s.ID, Name: s.Name, Action: s.Action,
			Fields: s.Fields, OptionalFields: s.OptionalFields,
			Documents: s.Documents, OptionalDocuments: s.OptionalDocuments,
			Outcomes: s.Outcomes, Next: s.Next,
		}
		if applicant[s.Area] {
			step.Role = definition.RoleApplicant
		} else {
			step.Role, step.Area = definition.RoleStaff, s.Area
			if s.Profile != "" {
				step.Permissions = []string{s.Profile}
				if asked[s.Area] == nil {
					asked[s.Area] = map[string]bool{}
				}
				asked[s.Area][s.Profile] = true
			}
		}
		d.Steps = append(d.Steps, step)
	}
	c := Content{Format: ContentFormat, Definition: d}
	for _, a := range l.Areas {
		if applicant[a.ID] {
			continue
		}
		area := definition.Area{ID: a.ID, Name: a.Name}
		for profile := range asked[a.ID] {
			name := legacyProfileNames[profile]
			if name == "" {
				name = profile
			}
			area.Permissions = append(area.Permissions, definition.Permission{ID: profile, Name: name})
		}
		slices.SortFunc(area.Permissions, func(a, b definition.Permission) int { return cmp.Compare(a.ID, b.ID) })
		c.Areas = append(c.Areas, area)
	}
	return c
}
