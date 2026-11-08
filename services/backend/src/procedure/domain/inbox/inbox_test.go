package inbox_test

import (
	"testing"
	"time"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
)

var (
	now      = time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	office   = definition.Area{ID: "office", Name: "Oficina", Kind: definition.AreaOffice}
	external = definition.Area{ID: "committee", Name: "Comisión", Kind: definition.AreaExternal}
	titular  = definition.Area{ID: "applicant", Name: "Titular", Kind: definition.AreaApplicant}
)

func TestCanAct(t *testing.T) {
	tests := []struct {
		name   string
		viewer definition.Profile
		step   inbox.Summary
		want   bool
	}{
		{"operative on an operative step", definition.ProfileStaffOperative, inbox.Summary{Area: office, Profile: definition.ProfileStaffOperative}, true},
		{"operative on an approval step", definition.ProfileStaffOperative, inbox.Summary{Area: office, Profile: definition.ProfileStaffApprover}, false},
		{"approver on an operative step", definition.ProfileStaffApprover, inbox.Summary{Area: office, Profile: definition.ProfileStaffOperative}, true},
		{"approver on an approval step", definition.ProfileStaffApprover, inbox.Summary{Area: office, Profile: definition.ProfileStaffApprover}, true},
		{"nobody acts on an external step", definition.ProfileStaffApprover, inbox.Summary{Area: external}, false},
		{"the staff does not do the applicant's step", definition.ProfileStaffApprover, inbox.Summary{Area: titular}, false},
		{"an unknown profile acts on nothing", "applicant", inbox.Summary{Area: office, Profile: definition.ProfileStaffOperative}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inbox.CanAct(tt.viewer, tt.step); got != tt.want {
				t.Errorf("CanAct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildCountsWholeDaysAndNeverNegative(t *testing.T) {
	summaries := []inbox.Summary{
		{ID: "a", Area: external, StartedAt: now.Add(-47 * time.Hour), LastMovedAt: now.Add(-25 * time.Hour)},
		// A clock that moved backwards must not show a negative age.
		{ID: "b", Area: external, StartedAt: now.Add(time.Hour), LastMovedAt: now.Add(time.Hour)},
	}

	list := inbox.Build(summaries, definition.ProfileStaffOperative, inbox.Query{View: inbox.ViewAll}, now)

	got := map[string][2]int{}
	for _, item := range list.Items {
		got[item.ID] = [2]int{item.AgeDays, *item.DaysReferred}
	}
	if got["a"] != [2]int{1, 1} || got["b"] != [2]int{0, 0} {
		t.Errorf("ages and days referred = %v", got)
	}
}

func TestBuildReturnsAnEmptyListAndOptionsOfTheView(t *testing.T) {
	summaries := []inbox.Summary{
		{ID: "a", ProcedureType: 2, ProcedureTypeName: "B", Area: external},
		{ID: "b", ProcedureType: 1, ProcedureTypeName: "A", Area: office, Profile: definition.ProfileStaffOperative},
		{ID: "c", ProcedureType: 1, ProcedureTypeName: "A", Area: office, Profile: definition.ProfileStaffOperative},
	}

	mine := inbox.Build(summaries, definition.ProfileStaffOperative, inbox.Query{View: inbox.ViewMine, ProcedureType: 9}, now)
	if mine.Items == nil || len(mine.Items) != 0 {
		t.Errorf("items = %#v, want an empty list", mine.Items)
	}
	// Only the procedure types of the view, each once and by name.
	if len(mine.ProcedureTypes) != 1 || mine.ProcedureTypes[0].Name != "A" {
		t.Errorf("type options of mine = %v", mine.ProcedureTypes)
	}
	all := inbox.Build(summaries, definition.ProfileStaffOperative, inbox.Query{View: inbox.ViewAll}, now)
	if len(all.ProcedureTypes) != 2 || all.ProcedureTypes[0].Name != "A" || len(all.Areas) != 2 || all.Areas[0].Name != "Comisión" {
		t.Errorf("options of all = %v, %v", all.ProcedureTypes, all.Areas)
	}
}

func TestBuildBreaksTiesByID(t *testing.T) {
	summaries := []inbox.Summary{
		{ID: "b", Area: titular, StartedAt: now, LastMovedAt: now},
		{ID: "a", Area: titular, StartedAt: now, LastMovedAt: now},
	}
	for _, sort := range []inbox.Sort{inbox.SortOldest, inbox.SortLastMoved} {
		list := inbox.Build(summaries, definition.ProfileStaffOperative, inbox.Query{View: inbox.ViewAll, Sort: sort}, now)
		if list.Items[0].ID != "a" {
			t.Errorf("sort %q puts %q first, want %q", sort, list.Items[0].ID, "a")
		}
	}
}
