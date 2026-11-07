package catalog

import (
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

type FieldRequirement struct {
	Field    definition.Field
	Required bool
}

type DocumentRequirement struct {
	Document definition.Document
	Required bool
}

// ApplicantRequirements returns what the applicant has to provide, in the order
// the definition declares it. What staff steps record is left out. Something
// mandatory in any applicant step is mandatory, even if another makes it
// optional.
func ApplicantRequirements(c version.Content) ([]FieldRequirement, []DocumentRequirement) {
	applicantAreas := map[string]bool{}
	for _, area := range c.Areas {
		applicantAreas[area.ID] = area.Kind == definition.AreaApplicant
	}
	// Fields and documents keep separate ids, so they are tracked apart.
	fields, documents := requirementSet{}, requirementSet{}
	for _, step := range c.Definition.Steps {
		if applicantAreas[step.Area] {
			fields.add(step.Fields, step.OptionalFields)
			documents.add(step.Documents, step.OptionalDocuments)
		}
	}

	var fieldRequirements []FieldRequirement
	for _, field := range c.Definition.Fields {
		if required, asked := fields[field.ID]; asked {
			fieldRequirements = append(fieldRequirements, FieldRequirement{Field: field, Required: required})
		}
	}
	var documentRequirements []DocumentRequirement
	for _, document := range c.Definition.Documents {
		if required, asked := documents[document.ID]; asked {
			documentRequirements = append(documentRequirements, DocumentRequirement{Document: document, Required: required})
		}
	}
	return fieldRequirements, documentRequirements
}

// requirementSet tells, for each id asked for, whether it is mandatory.
type requirementSet map[string]bool

func (requirements requirementSet) add(mandatory, optional []string) {
	for _, id := range optional {
		if _, asked := requirements[id]; !asked {
			requirements[id] = false
		}
	}
	for _, id := range mandatory {
		requirements[id] = true
	}
}
