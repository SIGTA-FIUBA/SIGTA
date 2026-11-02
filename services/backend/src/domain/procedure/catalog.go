package procedure

type FieldRequirement struct {
	Field    Field
	Required bool
}

type DocumentRequirement struct {
	Document Document
	Required bool
}

// ApplicantRequirements returns what the applicant has to provide, in the order
// the definition declares it. What staff steps record is left out. Something
// mandatory in any applicant step is mandatory, even if another makes it
// optional.
func (c Content) ApplicantRequirements() ([]FieldRequirement, []DocumentRequirement) {
	applicantAreas := map[string]bool{}
	for _, a := range c.Areas {
		applicantAreas[a.ID] = a.Kind == AreaApplicant
	}
	// Fields and documents keep separate ids, so they are tracked apart.
	fields, documents := requirementSet{}, requirementSet{}
	for _, s := range c.Definition.Steps {
		if applicantAreas[s.Area] {
			fields.add(s.Fields, s.OptionalFields)
			documents.add(s.Documents, s.OptionalDocuments)
		}
	}

	var fieldRequirements []FieldRequirement
	for _, f := range c.Definition.Fields {
		if required, asked := fields[f.ID]; asked {
			fieldRequirements = append(fieldRequirements, FieldRequirement{Field: f, Required: required})
		}
	}
	var documentRequirements []DocumentRequirement
	for _, d := range c.Definition.Documents {
		if required, asked := documents[d.ID]; asked {
			documentRequirements = append(documentRequirements, DocumentRequirement{Document: d, Required: required})
		}
	}
	return fieldRequirements, documentRequirements
}

// requirementSet tells, for each id asked for, whether it is mandatory.
type requirementSet map[string]bool

func (r requirementSet) add(mandatory, optional []string) {
	for _, id := range optional {
		if _, asked := r[id]; !asked {
			r[id] = false
		}
	}
	for _, id := range mandatory {
		r[id] = true
	}
}
