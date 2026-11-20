package postgres_test

import (
	"testing"

	"backend/src/auth/adapters/postgres"
	"backend/src/auth/domain"
	"backend/src/migrations/dbtest"
)

// A change made by hand to a system area stays after a save of a copy read
// before it, so that the server sees it instead of covering it up.
func TestASaveKeepsWhatWasChangedByHand(t *testing.T) {
	schema := dbtest.NewSchema(t)
	areas := postgres.NewAreaRepository(schema.Pool(t, dbtest.AppRole, dbtest.ReadCommitted))
	area := domain.Area{ID: "office", Name: "Mesa", System: true, Permissions: []domain.Permission{
		{ID: "edit", Name: "Editar", System: true},
	}}
	if err := areas.Save(t.Context(), area); err != nil {
		t.Fatal(err)
	}
	if _, err := schema.Admin(t).Exec(t.Context(), `UPDATE areas SET system = false;
		UPDATE area_permissions SET system = false, deleted_at = now(), deleted_by = 'by hand'`); err != nil {
		t.Fatal(err)
	}
	if err := areas.Save(t.Context(), area); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	got, err := areas.Get(t.Context(), "office")
	if err != nil {
		t.Fatal(err)
	}
	if got.System || got.Permissions[0].System || !got.Permissions[0].Deleted() {
		t.Errorf("Get() = %+v, want the area and its permission as they were left by hand", got)
	}
}
