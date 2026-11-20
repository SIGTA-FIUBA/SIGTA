package outputtest

import (
	"sync"
	"testing"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

// TestAdministratorRepository runs the contract of
// output.AdministratorRepository. newRepository returns an empty repository
// each time it is called.
func TestAdministratorRepository(t *testing.T, newRepository func(t *testing.T) output.AdministratorRepository) {
	t.Run("grant the first", func(t *testing.T) { grantFirst(t, newRepository(t)) })
	t.Run("concurrent first grants", func(t *testing.T) { concurrentFirstGrants(t, newRepository(t)) })
}

func grantFirst(t *testing.T, r output.AdministratorRepository) {
	ctx := t.Context()
	at := time.Date(2026, 11, 20, 13, 0, 0, 0, time.UTC)
	if got, err := r.IsAdministrator(ctx, "U-1"); err != nil || got {
		t.Errorf("IsAdministrator() of an empty repository = %v, %v; want false", got, err)
	}
	if granted, err := r.GrantFirst(ctx, domain.Administrator{ID: "A-1", UserID: "U-1", From: at}); err != nil || !granted {
		t.Fatalf("GrantFirst() = %v, %v; want true", granted, err)
	}
	for _, a := range []domain.Administrator{{ID: "A-2", UserID: "U-2", From: at}, {ID: "A-3", UserID: "U-1", From: at}} {
		if granted, err := r.GrantFirst(ctx, a); err != nil || granted {
			t.Errorf("GrantFirst(%s) with an administrator = %v, %v; want false", a.UserID, granted, err)
		}
	}
	if got, err := r.IsAdministrator(ctx, "U-1"); err != nil || !got {
		t.Errorf("IsAdministrator(U-1) = %v, %v; want true", got, err)
	}
	if got, err := r.IsAdministrator(ctx, "U-2"); err != nil || got {
		t.Errorf("IsAdministrator(U-2) = %v, %v; want false", got, err)
	}
}

func concurrentFirstGrants(t *testing.T, r output.AdministratorRepository) {
	ctx := t.Context()
	ids := []string{"U-1", "U-2", "U-3", "U-4", "U-5"}
	var wg sync.WaitGroup
	granted := make([]bool, len(ids))
	for i, id := range ids {
		wg.Go(func() {
			var err error
			granted[i], err = r.GrantFirst(ctx, domain.Administrator{ID: "A-" + id, UserID: id, From: time.Now()})
			if err != nil {
				t.Errorf("GrantFirst(%s) = %v", id, err)
			}
		})
	}
	wg.Wait()
	administrators := 0
	for i, id := range ids {
		got, err := r.IsAdministrator(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if got != granted[i] {
			t.Errorf("IsAdministrator(%s) = %v, but GrantFirst() reported %v", id, got, granted[i])
		}
		if got {
			administrators++
		}
	}
	if administrators != 1 {
		t.Errorf("%d administrators after concurrent first grants, want 1", administrators)
	}
}
