package application_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"backend/src/auth/adapters/memory"
	"backend/src/auth/application"
	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
)

var platformPermissions = []domain.Permission{
	{ID: input.EditProcedureTypes, Name: "Editar tipos de trámite", System: true},
	{ID: input.PublishProcedureTypes, Name: "Publicar tipos de trámite", System: true},
}

func TestEnsureSecretariatCreatesIt(t *testing.T) {
	areas := memory.NewAreaRepository()
	if err := application.CheckSecretariat(t.Context(), areas); !errors.Is(err, application.ErrSecretariatIncomplete) {
		t.Fatalf("CheckSecretariat() without it = %v, want ErrSecretariatIncomplete", err)
	}
	if err := application.EnsureSecretariat(t.Context(), areas); err != nil {
		t.Fatal(err)
	}
	got, err := areas.Get(t.Context(), input.SecretariatArea)
	want := domain.Area{ID: input.SecretariatArea, Name: "Secretaría de Gestión Académica", System: true, Permissions: platformPermissions}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("Secretariat = %+v, %v; want %+v", got, err, want)
	}
	if err := application.CheckSecretariat(t.Context(), areas); err != nil {
		t.Errorf("CheckSecretariat() after EnsureSecretariat() = %v", err)
	}
}

// A deleted Secretariat is refused too. EnsureSecretariat does not bring it
// back, since no area comes back once deleted: in a database, it is repaired
// by hand with the role that owns the tables (docs/15-despliegue.md).
func TestCheckSecretariatRefusesItDeleted(t *testing.T) {
	ctx := t.Context()
	areas := memory.NewAreaRepository()
	if err := areas.Save(ctx, domain.Area{
		ID: input.SecretariatArea, Name: "Secretaría", System: true, Permissions: platformPermissions,
		DeletedAt: time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC), DeletedBy: "admin",
	}); err != nil {
		t.Fatal(err)
	}
	err := application.CheckSecretariat(ctx, areas)
	if !errors.Is(err, application.ErrSecretariatIncomplete) || !strings.Contains(err.Error(), "area secretariat is deleted") {
		t.Errorf("CheckSecretariat() = %v, want ErrSecretariatIncomplete saying it is deleted", err)
	}
}

// The stores keep every permission once saved, so a platform permission is
// missing only if the Secretariat was saved without it.
func TestCheckSecretariatSaysWhatIsMissing(t *testing.T) {
	ctx := t.Context()
	areas := memory.NewAreaRepository()
	if err := areas.Save(ctx, domain.Area{ID: input.SecretariatArea, Name: "Secretaría", System: true, Permissions: platformPermissions[1:]}); err != nil {
		t.Fatal(err)
	}
	err := application.CheckSecretariat(ctx, areas)
	if !errors.Is(err, application.ErrSecretariatIncomplete) || !strings.Contains(err.Error(), "permission edit_procedure_types is missing") {
		t.Errorf("CheckSecretariat() = %v, want ErrSecretariatIncomplete saying edit_procedure_types is missing", err)
	}
}

// Each case changes the Secretariat by hand: CheckSecretariat refuses it, and
// EnsureSecretariat puts back what the code needs and keeps the rest.
func TestEnsureSecretariatRestoresWhatTheCodeNeeds(t *testing.T) {
	deleted := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name   string
		change func(a *domain.Area)
		says   string
	}{
		{"not a system area", func(a *domain.Area) { a.System = false }, "area secretariat is not a system one"},
		{"with a platform permission deleted", func(a *domain.Area) {
			a.Permissions[0].DeletedAt, a.Permissions[0].DeletedBy = deleted, "admin"
		}, "permission edit_procedure_types is deleted"},
		{"with a platform permission that is not a system one", func(a *domain.Area) {
			a.Permissions[1].Name, a.Permissions[1].System = "Publicar", false
		}, "permission publish_procedure_types is not a system one"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			areas := memory.NewAreaRepository()
			if err := application.EnsureSecretariat(ctx, areas); err != nil {
				t.Fatal(err)
			}
			area, err := areas.Get(ctx, input.SecretariatArea)
			if err != nil {
				t.Fatal(err)
			}
			area.Name = "Secretaría renombrada"
			area.Permissions = append(area.Permissions, domain.Permission{ID: "review", Name: "Revisar"})
			tc.change(&area)
			if err := areas.Save(ctx, area); err != nil {
				t.Fatal(err)
			}
			if err := application.CheckSecretariat(ctx, areas); !errors.Is(err, application.ErrSecretariatIncomplete) || !strings.Contains(err.Error(), tc.says) {
				t.Fatalf("CheckSecretariat() = %v, want ErrSecretariatIncomplete saying %q", err, tc.says)
			}

			if err := application.EnsureSecretariat(ctx, areas); err != nil {
				t.Fatal(err)
			}
			if err := application.CheckSecretariat(ctx, areas); err != nil {
				t.Errorf("CheckSecretariat() after EnsureSecretariat() = %v", err)
			}
			got, err := areas.Get(ctx, input.SecretariatArea)
			if err != nil {
				t.Fatal(err)
			}
			if got.Name != "Secretaría renombrada" || !got.HasActivePermission("review") {
				t.Errorf("Secretariat = %+v, want its name and the review permission kept", got)
			}
		})
	}
}
