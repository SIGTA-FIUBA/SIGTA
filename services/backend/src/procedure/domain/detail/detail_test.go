package detail_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/domain/version"
)

func TestSteps(t *testing.T) {
	content := version.Content{
		Definition: definition.ProcedureTypeDefinition{
			Fields:    []definition.Field{{ID: "reason"}, {ID: "phone"}},
			Documents: []definition.Document{{ID: "annex"}, {ID: "transcript"}, {ID: "ruling"}},
			Steps: []definition.Step{
				{ID: "documents", Role: definition.RoleApplicant, Fields: []string{"reason"}, OptionalFields: []string{"phone", "reason"}, Documents: []string{"transcript"}, OptionalDocuments: []string{"annex", "transcript"}, Next: "review"},
				{ID: "review", Role: definition.RoleStaff, Area: "office", Next: "ruling"},
				{ID: "ruling", Role: definition.RoleStaff, Area: "missing", Documents: []string{"ruling"}},
			},
		},
		Areas: []definition.Area{{ID: "office", Name: "Dirección de Alumnos"}},
	}

	steps := detail.Steps(content)

	if len(steps) != 3 {
		t.Fatalf("got %d steps, want 3", len(steps))
	}
	for i, want := range []string{"documents", "review", "ruling"} {
		if steps[i].Step.ID != want {
			t.Errorf("step %d is %s, want %s", i, steps[i].Step.ID, want)
		}
	}
	var fields []string
	for _, f := range steps[0].Fields {
		fields = append(fields, fmt.Sprintf("%s %t", f.Field.ID, f.Required))
	}
	if want := []string{"reason true", "phone false"}; !slices.Equal(fields, want) {
		t.Errorf("fields of the first step = %v, want %v", fields, want)
	}
	var documents []string
	for _, d := range steps[0].Documents {
		documents = append(documents, fmt.Sprintf("%s %t", d.Document.ID, d.Required))
	}
	if want := []string{"annex false", "transcript true"}; !slices.Equal(documents, want) {
		t.Errorf("documents of the first step = %v, want %v", documents, want)
	}
	if steps[0].Area != nil {
		t.Errorf("the step of the applicant has the area %+v", *steps[0].Area)
	}
	if a := steps[1].Area; a == nil || *a != (detail.Area{ID: "office", Name: "Dirección de Alumnos"}) || len(steps[1].Fields) != 0 || len(steps[1].Documents) != 0 {
		t.Errorf("second step = %+v", steps[1])
	}
	// A version keeps the areas its steps use, so this only guards the display.
	if a := steps[2].Area; a == nil || a.Name != "missing" {
		t.Errorf("an area the version does not keep shows as %+v, want its id", a)
	}
}

func TestWithFiles(t *testing.T) {
	asks := func(ids ...string) []definition.DocumentRequirement {
		var r []definition.DocumentRequirement
		for _, id := range ids {
			r = append(r, definition.DocumentRequirement{Document: definition.Document{ID: id}})
		}
		return r
	}
	steps := []detail.Step{
		{Step: definition.Step{ID: "documents"}, Documents: asks("annex", "transcript")},
		{Step: definition.Step{ID: "review"}},
		{Step: definition.Step{ID: "fix"}, Documents: asks("transcript", "ruling")},
	}
	attached := []domain.Document{
		{ID: "D-1", Requirement: "transcript"},
		{ID: "D-2", Requirement: "annex"},
		{ID: "D-3", Requirement: "transcript"},
	}

	steps = detail.WithFiles(steps, attached)

	got := make([]string, 0, len(steps))
	for _, s := range steps {
		var files []string
		for _, d := range s.Files {
			files = append(files, d.Requirement+" "+d.ID)
		}
		got = append(got, fmt.Sprintf("%s %v", s.Step.ID, files))
	}
	// A document that two steps ask for shows its current file in both, and a
	// document with no file shows none.
	want := []string{"documents [annex D-2 transcript D-3]", "review []", "fix [transcript D-3]"}
	if !slices.Equal(got, want) {
		t.Errorf("files = %v, want %v", got, want)
	}
}

func TestReferralOf(t *testing.T) {
	content := version.Content{Areas: []definition.Area{{ID: "affairs", Name: "Dirección de Alumnos"}, {ID: "commission", Name: "Comisión Curricular"}}}
	since := time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC)
	now := time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)

	if got := detail.ReferralOf(domain.Procedure{OpenStep: "check"}, content, now); got != nil {
		t.Errorf("referral of a procedure not referred = %+v, want nil", got)
	}

	p := domain.Procedure{OpenStep: "evaluate", Referral: &domain.Referral{From: "affairs", To: "commission", Since: since}}
	want := detail.Referral{
		Area:              detail.Area{ID: "commission", Name: "Comisión Curricular"},
		From:              detail.Area{ID: "affairs", Name: "Dirección de Alumnos"},
		DaysWithoutReturn: 4,
	}
	if got := detail.ReferralOf(p, content, now); got == nil || *got != want {
		t.Errorf("referral = %+v, want %+v", got, want)
	}
}
