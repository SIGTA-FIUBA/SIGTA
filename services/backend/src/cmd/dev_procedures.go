package main

import (
	"fmt"
	"slices"
	"time"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
)

// devProcedures are made-up procedures, in progress and a few closed, so that
// the inbox has something to show in local development until procedures can
// be started and moved. The people and the file numbers are invented
// (ADR-005), and the dates are relative to now so that ages stay the same
// whenever the server starts.
func devProcedures(now time.Time) []inbox.Summary {
	applicant := definition.Area{ID: "applicant", Name: "Titular", Kind: definition.AreaApplicant}
	office := definition.Area{ID: "students_office", Name: "Dirección de Alumnos", Kind: definition.AreaOffice}
	director := definition.Area{ID: "students_director", Name: "Dirección General de Alumnos", Kind: definition.AreaOffice}
	commission := definition.Area{ID: "curricular_commission", Name: "Comisión Curricular", Kind: definition.AreaExternal}
	council := definition.Area{ID: "council_office", Name: "Despacho del Consejo Directivo", Kind: definition.AreaExternal}

	const equivalence, recognition = 2, 3
	names := map[int64]string{equivalence: "Equivalencia entre asignaturas", recognition: "Reconocimiento de créditos"}
	daysAgo := func(days int) time.Time { return now.Add(-time.Duration(days) * 24 * time.Hour) }
	// Both types share these steps, in this order; the file number exists
	// once "Registrar expediente" is done.
	steps := []string{
		"Solicitud", "Documentación pedida", "Registrar expediente", "Dictamen de la comisión",
		"Proyecto de resolución", "Resolución del Consejo Directivo", "Firma de la resolución",
		"Carga en el SIU", "Visto bueno del cierre",
	}
	const recorded = 3
	programs := []string{
		"Ingeniería en Informática", "Ingeniería Civil", "Ingeniería Mecánica",
		"Ingeniería Electrónica", "Ingeniería Química", "",
	}

	rows := []struct {
		procedureType  int64
		applicant      string
		step           string
		area           definition.Area
		profile        definition.Profile
		started, moved int
		// closed is how many days ago it was closed; negative while in
		// progress.
		closed int
	}{
		{equivalence, "Lucía Ejemplo", "Registrar expediente", office, definition.ProfileStaffOperative, 2, 2, -1},
		{equivalence, "Martín Prueba", "Registrar expediente", office, definition.ProfileStaffOperative, 6, 5, -1},
		{recognition, "Sofía Inventada", "Registrar expediente", office, definition.ProfileStaffOperative, 1, 1, -1},
		{equivalence, "Tomás Ficticio", "Dictamen de la comisión", commission, "", 34, 27, -1},
		{recognition, "Valentina Muestra", "Dictamen de la comisión", commission, "", 19, 15, -1},
		{equivalence, "Joaquín Ejemplo", "Proyecto de resolución", office, definition.ProfileStaffOperative, 41, 3, -1},
		{equivalence, "Camila Prueba", "Resolución del Consejo Directivo", council, "", 58, 12, -1},
		{recognition, "Mateo Inventado", "Firma de la resolución", office, definition.ProfileStaffApprover, 25, 4, -1},
		{equivalence, "Julieta Ficticia", "Carga en el SIU", office, definition.ProfileStaffOperative, 73, 6, -1},
		{recognition, "Bruno Muestra", "Visto bueno del cierre", director, definition.ProfileStaffApprover, 30, 1, -1},
		{equivalence, "Renata Ejemplo", "Documentación pedida", applicant, "", 9, 4, -1},
		{recognition, "Lautaro Prueba", "Solicitud", applicant, "", 0, 0, -1},
		{equivalence, "Agustina Ejemplo", "Visto bueno del cierre", director, definition.ProfileStaffApprover, 64, 3, 2},
		{recognition, "Federico Prueba", "Visto bueno del cierre", director, definition.ProfileStaffApprover, 40, 6, 5},
		{equivalence, "Paula Inventada", "Visto bueno del cierre", director, definition.ProfileStaffApprover, 90, 45, 44},
	}

	summaries := make([]inbox.Summary, 0, len(rows))
	for i, row := range rows {
		number := slices.Index(steps, row.step) + 1
		fileNumber := ""
		if number > recorded {
			fileNumber = fmt.Sprintf("EX-%d-%08d", now.Year(), 41000+i*137)
		}
		var closedAt *time.Time
		if row.closed >= 0 {
			at := daysAgo(row.closed)
			closedAt = &at
		}
		summaries = append(summaries, inbox.Summary{
			ID:                fmt.Sprintf("dev-%02d", i+1),
			ProcedureType:     row.procedureType,
			ProcedureTypeName: names[row.procedureType],
			ApplicantID:       fmt.Sprintf("dev-applicant-%02d", i+1),
			ApplicantName:     row.applicant,
			ApplicantProgram:  programs[i%len(programs)],
			FileNumber:        fileNumber,
			StepName:          row.step,
			StepNumber:        number,
			StepCount:         len(steps),
			Area:              row.area,
			Profile:           row.profile,
			StartedAt:         daysAgo(row.started),
			LastMovedAt:       daysAgo(row.moved),
			ClosedAt:          closedAt,
		})
	}
	return summaries
}
