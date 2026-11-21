package outputtest

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

// TestAreaRepository runs the contract of output.AreaRepository.
// newRepository returns an empty repository each time it is called.
func TestAreaRepository(t *testing.T, newRepository func(t *testing.T) output.AreaRepository) {
	t.Run("round trip", func(t *testing.T) { areaRoundTrip(t, newRepository(t)) })
	t.Run("list in the order added, without the deleted", func(t *testing.T) { listAreas(t, newRepository(t)) })
	t.Run("changes keep what was there", func(t *testing.T) { changeArea(t, newRepository(t)) })
	t.Run("an area read before another save does not undo it", func(t *testing.T) { saveOlderCopy(t, newRepository(t)) })
	t.Run("concurrent new areas", func(t *testing.T) { concurrentNewAreas(t, newRepository(t)) })
}

// Area builds an area with a deleted permission and a deleted role, so that
// every field has a value.
func Area(id string) domain.Area {
	return domain.Area{
		ID:   id,
		Name: "Área " + id,
		Permissions: []domain.Permission{
			{ID: "review", Name: "Revisar"},
			{ID: "approve", Name: "Aprobar", System: true},
			{ID: "old", Name: "Viejo", DeletedAt: at.Add(time.Hour), DeletedBy: "admin"},
		},
		Roles: []domain.Role{
			{ID: "reviewer", Name: "Revisor", Revisions: []domain.RoleRevision{
				{Permissions: []string{"review", "old"}, At: at, By: "admin"},
				{Permissions: []string{"review"}, At: at.Add(time.Hour), By: "otro"},
			}},
			{ID: "nobody", Name: "Nadie", Revisions: []domain.RoleRevision{{At: at, By: "admin"}}, DeletedAt: at.Add(2 * time.Hour), DeletedBy: "admin"},
		},
	}
}

func saveAreas(t *testing.T, r output.AreaRepository, areas ...domain.Area) {
	t.Helper()
	for _, a := range areas {
		if err := r.Save(t.Context(), a); err != nil {
			t.Fatalf("Save(%s) = %v", a.ID, err)
		}
	}
}

func getArea(t *testing.T, r output.AreaRepository, id string) domain.Area {
	t.Helper()
	a, err := r.Get(t.Context(), id)
	if err != nil {
		t.Fatalf("Get(%s) = %v", id, err)
	}
	return a
}

func areaRoundTrip(t *testing.T, r output.AreaRepository) {
	if _, err := r.Get(t.Context(), "office"); !errors.Is(err, domain.ErrAreaNotFound) {
		t.Errorf("Get() of an empty repository = %v, want ErrAreaNotFound", err)
	}
	full := Area("office")
	bare := domain.Area{ID: "secretariat", Name: "Secretaría", System: true}
	saveAreas(t, r, full, bare)
	for _, want := range []domain.Area{full, bare} {
		if got := getArea(t, r, want.ID); !reflect.DeepEqual(got, want) {
			t.Errorf("Get(%s) = %+v, want %+v", want.ID, got, want)
		}
	}

	got := getArea(t, r, "office")
	got.Permissions[0].Name = "Cambiado"
	got.Roles[0].Revisions[0].Permissions[0] = "Cambiado"
	if again := getArea(t, r, "office"); !reflect.DeepEqual(again, full) {
		t.Error("the repository shares the area with the caller")
	}
}

func listAreas(t *testing.T, r output.AreaRepository) {
	deleted := Area("archive")
	deleted.DeletedAt, deleted.DeletedBy = at, "admin"
	saveAreas(t, r, Area("office"), deleted, Area("committee"))
	list, err := r.List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if want := []domain.Area{Area("office"), Area("committee")}; !reflect.DeepEqual(list, want) {
		t.Errorf("List() = %+v, want %+v", list, want)
	}
	if got := getArea(t, r, "archive"); !reflect.DeepEqual(got, deleted) {
		t.Errorf("Get() of a deleted area = %+v, want %+v", got, deleted)
	}
}

// changeArea makes the changes that the application makes, on an area saved
// before: each one is a new Save of the whole area.
func changeArea(t *testing.T, r output.AreaRepository) {
	saveAreas(t, r, domain.Area{ID: "office", Name: "Mesa"})
	steps := []func(a *domain.Area){
		func(a *domain.Area) {
			a.Permissions = append(a.Permissions, domain.Permission{ID: "review", Name: "Revisar"})
		},
		func(a *domain.Area) {
			a.Roles = append(a.Roles, domain.Role{ID: "reviewer", Name: "Revisor", Revisions: []domain.RoleRevision{{Permissions: []string{"review"}, At: at, By: "admin"}}})
		},
		func(a *domain.Area) {
			a.Roles[0].Revisions = append(a.Roles[0].Revisions, domain.RoleRevision{At: at.Add(time.Hour), By: "admin"})
		},
		func(a *domain.Area) { a.Permissions[0].DeletedAt, a.Permissions[0].DeletedBy = at, "admin" },
		func(a *domain.Area) {
			a.Permissions[0].Name = "Revisar casos"
		},
		func(a *domain.Area) { a.Roles[0].DeletedAt, a.Roles[0].DeletedBy = at, "admin" },
		func(a *domain.Area) { a.Name = "Mesa de entradas" },
		func(a *domain.Area) { a.DeletedAt, a.DeletedBy = at, "admin" },
	}
	for i, change := range steps {
		want := getArea(t, r, "office")
		change(&want)
		saveAreas(t, r, want)
		if got := getArea(t, r, "office"); !reflect.DeepEqual(got, want) {
			t.Fatalf("Get() after change %d = %+v, want %+v", i+1, got, want)
		}
	}
	if list, err := r.List(t.Context()); err != nil || len(list) != 0 {
		t.Errorf("List() after deleting the only area = %+v, %v; want none", list, err)
	}
}

// saveOlderCopy saves two changes made on the same reading of an area, as
// two administrators editing it at once do.
func saveOlderCopy(t *testing.T, r output.AreaRepository) {
	initial := []domain.RoleRevision{{Permissions: []string{"review"}, At: at, By: "admin"}}
	saveAreas(t, r, domain.Area{
		ID: "office", Name: "Mesa",
		Permissions: []domain.Permission{{ID: "review", Name: "Revisar"}},
		Roles:       []domain.Role{{ID: "reviewer", Name: "Revisor", Revisions: initial}},
	})
	first, second := getArea(t, r, "office"), getArea(t, r, "office")
	byFirst := domain.RoleRevision{At: at.Add(time.Hour), By: "admin"}
	first.Permissions = append(first.Permissions, domain.Permission{ID: "approve", Name: "Aprobar"})
	first.Permissions[0].DeletedAt, first.Permissions[0].DeletedBy = at, "admin"
	first.Roles[0].Revisions = append(first.Roles[0].Revisions, byFirst)
	bySecond := domain.RoleRevision{Permissions: []string{"review"}, At: at.Add(2 * time.Hour), By: "otro"}
	second.Name = "Mesa de entradas"
	second.Permissions = append(second.Permissions, domain.Permission{ID: "sign", Name: "Firmar"})
	second.Roles[0].Revisions = append(second.Roles[0].Revisions, bySecond)
	saveAreas(t, r, first, second)

	want := domain.Area{
		ID: "office", Name: "Mesa de entradas",
		Permissions: []domain.Permission{
			{ID: "review", Name: "Revisar", DeletedAt: at, DeletedBy: "admin"},
			{ID: "approve", Name: "Aprobar"},
			{ID: "sign", Name: "Firmar"},
		},
		Roles: []domain.Role{{ID: "reviewer", Name: "Revisor", Revisions: append(initial, byFirst, bySecond)}},
	}
	if got := getArea(t, r, "office"); !reflect.DeepEqual(got, want) {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func concurrentNewAreas(t *testing.T, r output.AreaRepository) {
	const areas = 10
	var wg sync.WaitGroup
	errs := make([]error, areas)
	for i := range areas {
		wg.Go(func() { errs[i] = r.Save(t.Context(), Area(fmt.Sprintf("area_%d", i))) })
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("Save() number %d = %v", i, err)
		}
	}
	if list, err := r.List(t.Context()); err != nil || len(list) != areas {
		t.Errorf("List() = %d areas, %v; want %d", len(list), err, areas)
	}
}
