package memory_test

import (
	"testing"
	"time"

	"backend/src/procedure/adapters/memory"
)

func TestProcedureViewsKeepsTheLastViewOfEachPerson(t *testing.T) {
	views := memory.NewProcedureViews()
	first := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	later := first.Add(time.Hour)

	for _, at := range []time.Time{first, later} {
		if err := views.See(t.Context(), "u1", "p-1", at); err != nil {
			t.Fatal(err)
		}
	}
	seen, err := views.SeenBy(t.Context(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if !seen["p-1"].Equal(later) {
		t.Errorf("seen at %v, want %v", seen["p-1"], later)
	}
	// What one person saw says nothing of another.
	other, err := views.SeenBy(t.Context(), "u2")
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 0 {
		t.Errorf("u2 saw %v", other)
	}
	// The caller cannot change what is kept.
	seen["p-2"] = later
	if again, _ := views.SeenBy(t.Context(), "u1"); len(again) != 1 {
		t.Errorf("seen = %v", again)
	}
}
