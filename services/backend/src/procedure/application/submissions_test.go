package application

import (
	"slices"
	"testing"
	"time"

	"backend/src/procedure/adapters/memory"
	procdomain "backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/submission"
	"backend/src/procedure/ports/input"
)

func TestSubmitRecordsTheOpeningStepsAsDoneByTheApplicant(t *testing.T) {
	ctx := t.Context()
	d := testDefinition()
	d.Fields = append(d.Fields, definition.Field{ID: "reason", Name: "Motivo", Type: definition.FieldLongText})
	d.Steps = []definition.Step{
		{ID: "request", Name: "Solicitud", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Fields: []string{"national_id"}, Next: "detail"},
		{ID: "detail", Name: "Detalle", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Fields: []string{"reason"}, Next: "review"},
		{ID: "review", Name: "Revisión", Role: definition.RoleStaff, Area: office.ID, Permissions: []string{"review"}, Action: definition.ActionConfirmExternalFact},
	}
	types := newService(memory.NewProcedureVersions(), office)
	if _, err := types.Publish(ctx, publication(1, 0, d)); err != nil {
		t.Fatal(err)
	}
	procedures := memory.NewProcedureRepository()
	at := time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC)
	service := NewSubmissionService(
		NewProcedureService(types, procedures, applicants{}, nil, time.Now),
		memory.NewDocumentRepository(), nil, func() string { return "P-1" }, func() time.Time { return at },
	)

	started, err := service.Submit(ctx, "ana", input.Submission{
		ProcedureType: 1,
		Values:        map[string]submission.Value{"national_id": {Text: "30111222"}, "reason": {Text: "Cursé afuera"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err := procedures.Events(ctx, started.Procedure.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := []procdomain.Event{
		{ProcedureID: "P-1", Step: "request", Action: definition.ActionEnterData, By: "ana", At: at, Next: "detail"},
		{ProcedureID: "P-1", Step: "detail", Action: definition.ActionEnterData, By: "ana", At: at, Next: "review"},
	}
	if !slices.Equal(events, want) {
		t.Errorf("events = %+v, want %+v", events, want)
	}
}
