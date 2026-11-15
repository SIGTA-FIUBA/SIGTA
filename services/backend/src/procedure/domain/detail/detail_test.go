package detail_test

import (
	"fmt"
	"slices"
	"testing"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/domain/version"
)

func TestSteps(t *testing.T) {
	content := version.Content{
		Definition: definition.ProcedureTypeDefinition{
			Documents: []definition.Document{{ID: "annex"}, {ID: "transcript"}, {ID: "ruling"}},
			Steps: []definition.Step{
				{ID: "documents", Role: definition.RoleApplicant, Documents: []string{"transcript"}, OptionalDocuments: []string{"annex", "transcript"}, Next: "review"},
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
	var documents []string
	for _, d := range steps[0].Documents {
		documents = append(documents, fmt.Sprintf("%s %t", d.Document.ID, d.Required))
	}
	if want := []string{"annex false", "transcript true"}; !slices.Equal(documents, want) {
		t.Errorf("documents of the first step = %v, want %v", documents, want)
	}
	if steps[0].AreaName != "" {
		t.Errorf("the step of the applicant has the area %q", steps[0].AreaName)
	}
	if steps[1].AreaName != "Dirección de Alumnos" || len(steps[1].Documents) != 0 {
		t.Errorf("second step = %+v", steps[1])
	}
	// A version keeps the areas its steps use, so this only guards the display.
	if steps[2].AreaName != "missing" {
		t.Errorf("an area the version does not keep shows as %q, want its id", steps[2].AreaName)
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
