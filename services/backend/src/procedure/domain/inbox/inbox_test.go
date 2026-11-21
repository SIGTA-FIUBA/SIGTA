package inbox_test

import (
	"testing"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/inbox"
)

var (
	now       = time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	office    = &inbox.Option[string]{ID: "office", Name: "Oficina"}
	committee = &inbox.Option[string]{ID: "committee", Name: "Comisión"}
)

func ids(list inbox.List) []string {
	got := []string{}
	for _, item := range list.Items {
		got = append(got, item.ID)
	}
	return got
}

func TestBuildCountsWholeDaysAndNeverNegative(t *testing.T) {
	summaries := []inbox.Summary{
		{ID: "a", StartedAt: now.Add(-47 * time.Hour)},
		// A clock that moved backwards must not show a negative age.
		{ID: "b", StartedAt: now.Add(time.Hour)},
	}
	list := inbox.Build(summaries, inbox.Query{View: inbox.ViewAll}, now)

	got := map[string]int{}
	for _, item := range list.Items {
		got[item.ID] = item.AgeDays
	}
	if got["a"] != 1 || got["b"] != 0 {
		t.Errorf("ages = %v, want a: 1, b: 0", got)
	}
}

func TestBuildShowsInMineOnlyWhatThePersonCanDo(t *testing.T) {
	summaries := []inbox.Summary{
		{ID: "can", Area: office, State: domain.StateInOffice, Actionable: true},
		{ID: "cannot", Area: office, State: domain.StateInOffice},
		{ID: "applicant", State: domain.StateWithApplicant},
	}
	if got := ids(inbox.Build(summaries, inbox.Query{View: inbox.ViewMine}, now)); len(got) != 1 || got[0] != "can" {
		t.Errorf("mine = %v, want only can", got)
	}
	if got := ids(inbox.Build(summaries, inbox.Query{View: inbox.ViewAll}, now)); len(got) != 3 {
		t.Errorf("all = %v, want the three", got)
	}
}

func TestBuildReturnsAnEmptyListAndOptionsOfTheView(t *testing.T) {
	summaries := []inbox.Summary{
		{ID: "a", ProcedureType: 2, ProcedureTypeName: "B", Area: committee},
		{ID: "b", ProcedureType: 1, ProcedureTypeName: "A", Area: office, Actionable: true},
		{ID: "c", ProcedureType: 1, ProcedureTypeName: "A", Area: office, Actionable: true},
		{ID: "d", ProcedureType: 1, ProcedureTypeName: "A", State: domain.StateWithApplicant},
	}

	mine := inbox.Build(summaries, inbox.Query{View: inbox.ViewMine, ProcedureType: 9}, now)
	if mine.Items == nil || len(mine.Items) != 0 {
		t.Errorf("items = %#v, want an empty list", mine.Items)
	}
	// Only the procedure types of the view, each once and by name.
	if len(mine.ProcedureTypes) != 1 || mine.ProcedureTypes[0].Name != "A" {
		t.Errorf("type options of mine = %v", mine.ProcedureTypes)
	}
	// A step of the applicant has no area, so it adds no option.
	all := inbox.Build(summaries, inbox.Query{View: inbox.ViewAll}, now)
	if len(all.ProcedureTypes) != 2 || all.ProcedureTypes[0].Name != "A" || len(all.Areas) != 2 || all.Areas[0].Name != "Comisión" {
		t.Errorf("options of all = %v, %v", all.ProcedureTypes, all.Areas)
	}
}

func TestBuildFiltersByTypeStateAndArea(t *testing.T) {
	summaries := []inbox.Summary{
		{ID: "office-1", ProcedureType: 1, Area: office, State: domain.StateInOffice},
		{ID: "committee-1", ProcedureType: 1, Area: committee, State: domain.StateInOffice},
		{ID: "applicant-2", ProcedureType: 2, State: domain.StateWithApplicant},
	}
	tests := []struct {
		name string
		q    inbox.Query
		want []string
	}{
		{"by type", inbox.Query{ProcedureType: 2}, []string{"applicant-2"}},
		{"by state", inbox.Query{State: domain.StateInOffice}, []string{"committee-1", "office-1"}},
		{"by area, which a step of the applicant does not have", inbox.Query{Area: "office"}, []string{"office-1"}},
		{"combined", inbox.Query{ProcedureType: 1, Area: "committee"}, []string{"committee-1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.q.View = inbox.ViewAll
			got := ids(inbox.Build(summaries, tt.q, now))
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestBuildSorts(t *testing.T) {
	summaries := []inbox.Summary{
		{ID: "new-but-still", StartedAt: now.Add(-time.Hour), LastMovedAt: now.Add(-time.Hour)},
		{ID: "old-but-moved", StartedAt: now.Add(-72 * time.Hour), LastMovedAt: now.Add(-time.Minute)},
	}
	if got := ids(inbox.Build(summaries, inbox.Query{View: inbox.ViewAll, Sort: inbox.SortOldest}, now)); got[0] != "old-but-moved" {
		t.Errorf("oldest first = %v", got)
	}
	if got := ids(inbox.Build(summaries, inbox.Query{View: inbox.ViewAll, Sort: inbox.SortLastMoved}, now)); got[0] != "old-but-moved" || got[1] != "new-but-still" {
		t.Errorf("last moved first = %v", got)
	}
}

func TestBuildBreaksTiesByID(t *testing.T) {
	summaries := []inbox.Summary{
		{ID: "b", StartedAt: now, LastMovedAt: now},
		{ID: "a", StartedAt: now, LastMovedAt: now},
	}
	for _, sort := range []inbox.Sort{inbox.SortOldest, inbox.SortLastMoved} {
		list := inbox.Build(summaries, inbox.Query{View: inbox.ViewAll, Sort: sort}, now)
		if list.Items[0].ID != "a" {
			t.Errorf("sort %q puts %q first, want %q", sort, list.Items[0].ID, "a")
		}
	}
}
