package domain_test

import (
	"testing"
	"time"

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

func TestDaysWithoutReturn(t *testing.T) {
	since := time.Date(2026, 11, 16, 18, 0, 0, 0, time.UTC)
	p := domain.Procedure{Referral: &domain.Referral{From: "affairs", To: "commission", Since: since}}

	tests := []struct {
		name string
		now  time.Time
		want int
	}{
		{"same day", time.Date(2026, 11, 16, 23, 0, 0, 0, time.UTC), 0},
		{"next morning counts a day", time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC), 1},
		{"four days later", time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC), 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, referred := p.DaysWithoutReturn(tt.now)
			if !referred || got != tt.want {
				t.Errorf("DaysWithoutReturn = %d, %v; want %d, true", got, referred, tt.want)
			}
		})
	}
}

func TestDaysWithoutReturnOfAProcedureNotReferred(t *testing.T) {
	if _, referred := (domain.Procedure{OpenStep: "check"}).DaysWithoutReturn(time.Now()); referred {
		t.Error("a procedure without referral is reported as referred")
	}
}
