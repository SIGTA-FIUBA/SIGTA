package main

import (
	"fmt"
	"time"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
)

// devProcedures are made-up procedures in progress, so that the inbox has
// something to show in local development until procedures can be started and
// moved. The people are invented (ADR-005), and the dates are relative to now
// so that ages stay the same whenever the server starts.
func devProcedures(now time.Time) []inbox.Summary {
	applicant := definition.Area{ID: "applicant", Name: "Titular", Kind: definition.AreaApplicant}
	office := definition.Area{ID: "students_office", Name: "Dirección de Alumnos", Kind: definition.AreaOffice}
	director := definition.Area{ID: "students_director", Name: "Dirección General de Alumnos", Kind: definition.AreaOffice}
	commission := definition.Area{ID: "curricular_commission", Name: "Comisión Curricular", Kind: definition.AreaExternal}
	council := definition.Area{ID: "council_office", Name: "Despacho del Consejo Directivo", Kind: definition.AreaExternal}

	const equivalence, recognition = 2, 3
	names := map[int64]string{equivalence: "Equivalencia entre asignaturas", recognition: "Reconocimiento de créditos"}
	daysAgo := func(days int) time.Time { return now.Add(-time.Duration(days) * 24 * time.Hour) }

	rows := []struct {
		procedureType  int64
		applicant      string
		step           string
		area           definition.Area
		profile        definition.Profile
		started, moved int
	}{
		{equivalence, "Lucía Ejemplo", "Registrar expediente", office, definition.ProfileStaffOperative, 2, 2},
		{equivalence, "Martín Prueba", "Registrar expediente", office, definition.ProfileStaffOperative, 6, 5},
		{recognition, "Sofía Inventada", "Registrar expediente", office, definition.ProfileStaffOperative, 1, 1},
		{equivalence, "Tomás Ficticio", "Dictamen de la comisión", commission, "", 34, 27},
		{recognition, "Valentina Muestra", "Dictamen de la comisión", commission, "", 19, 15},
		{equivalence, "Joaquín Ejemplo", "Proyecto de resolución", office, definition.ProfileStaffOperative, 41, 3},
		{equivalence, "Camila Prueba", "Resolución del Consejo Directivo", council, "", 58, 12},
		{recognition, "Mateo Inventado", "Firma de la resolución", office, definition.ProfileStaffApprover, 25, 4},
		{equivalence, "Julieta Ficticia", "Carga en el SIU", office, definition.ProfileStaffOperative, 73, 6},
		{recognition, "Bruno Muestra", "Visto bueno del cierre", director, definition.ProfileStaffApprover, 30, 1},
		{equivalence, "Renata Ejemplo", "Documentación pedida", applicant, "", 9, 4},
		{recognition, "Lautaro Prueba", "Solicitud", applicant, "", 0, 0},
	}

	summaries := make([]inbox.Summary, 0, len(rows))
	for i, row := range rows {
		summaries = append(summaries, inbox.Summary{
			ID:                fmt.Sprintf("dev-%02d", i+1),
			ProcedureType:     row.procedureType,
			ProcedureTypeName: names[row.procedureType],
			ApplicantID:       fmt.Sprintf("dev-applicant-%02d", i+1),
			ApplicantName:     row.applicant,
			StepName:          row.step,
			Area:              row.area,
			Profile:           row.profile,
			StartedAt:         daysAgo(row.started),
			LastMovedAt:       daysAgo(row.moved),
		})
	}
	return summaries
}
