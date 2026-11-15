package domain_test

import (
	"testing"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
)

func TestStateOf(t *testing.T) {
	def := definition.ProcedureTypeDefinition{Steps: []definition.Step{
		{ID: "documents", Role: definition.RoleApplicant},
		{ID: "review", Role: definition.RoleStaff, Area: "office"},
	}}
	for _, c := range []struct {
		openStep string
		want     domain.State
	}{
		{"documents", domain.StateWithApplicant},
		{"review", domain.StateInOffice},
		{"", domain.StateClosed},
	} {
		got, err := domain.StateOf(domain.Procedure{ID: "P-1", OpenStep: c.openStep}, def)
		if err != nil || got != c.want {
			t.Errorf("open step %q: state = %q, %v, want %q", c.openStep, got, err, c.want)
		}
	}
	if _, err := domain.StateOf(domain.Procedure{ID: "P-1", OpenStep: "gone"}, def); err == nil {
		t.Error("an open step that is not in the definition has a state")
	}
}
