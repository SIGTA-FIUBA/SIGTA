// Package definitiontest holds a valid definition and its areas for the tests
// of every package that works with procedure type definitions.
package definitiontest

import "backend/src/procedure/domain/definition"

// Areas returns two staff areas: the applicant has none.
func Areas() []definition.Area {
	return []definition.Area{
		{ID: "office", Name: "Dirección de Alumnos"},
		{ID: "commission", Name: "Comisión Curricular"},
	}
}

// Definition returns a valid definition with one step of each action. Tests
// change it by index, so the order of its elements is part of what it offers.
func Definition() definition.ProcedureTypeDefinition {
	return definition.ProcedureTypeDefinition{
		Name:        "Trámite de prueba",
		Description: "Un trámite para las pruebas",
		Fields:      []definition.Field{{ID: "national_id", Name: "DNI", Type: definition.FieldText}},
		Documents:   []definition.Document{{ID: "letter", Name: "Carta", AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: 1024}},
		Steps: []definition.Step{
			{ID: "request", Name: "Solicitud", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Fields: []string{"national_id"}, Next: "letter"},
			{ID: "letter", Name: "Carta", Role: definition.RoleApplicant, Action: definition.ActionUploadDocument, Documents: []string{"letter"}, Next: "review"},
			{ID: "review", Name: "Revisión", Role: definition.RoleStaff, Area: "office", Profile: definition.ProfileStaffOperative, Action: definition.ActionDecide, Outcomes: []definition.Outcome{
				{ID: "complete", Name: "Completa", Next: "closed"},
				{ID: "incomplete", Name: "Incompleta", Next: "letter"},
			}},
			{ID: "closed", Name: "Archivado", Role: definition.RoleStaff, Area: "office", Profile: definition.ProfileStaffOperative, Action: definition.ActionConfirmExternalFact},
		},
	}
}
