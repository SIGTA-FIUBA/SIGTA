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

// TestAdministratorRepository runs the contract of
// output.FirstAdministratorRepository. newRepository returns an empty
// repository each time it is called.
func TestAdministratorRepository(t *testing.T, newRepository func(t *testing.T) output.FirstAdministratorRepository) {
	t.Run("grant the first", func(t *testing.T) { grantFirst(t, newRepository(t)) })
	// A race shows only in some rounds, so it runs several.
	t.Run("concurrent first grants", func(t *testing.T) {
		for round := range 20 {
			t.Run(fmt.Sprint(round), func(t *testing.T) { concurrentFirstGrants(t, newRepository(t)) })
		}
	})
}

func grantFirst(t *testing.T, r output.FirstAdministratorRepository) {
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

func concurrentFirstGrants(t *testing.T, r output.FirstAdministratorRepository) {
	ctx := t.Context()
	ids := []string{"U-1", "U-2", "U-3", "U-4", "U-5", "U-6", "U-7", "U-8"}
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

// TestAdministratorPeriodRepository runs the contract of
// output.AdministratorPeriodRepository. newRepository returns an empty
// repository each time it is called.
func TestAdministratorPeriodRepository(t *testing.T, newRepository func(t *testing.T) output.AdministratorPeriodRepository) {
	t.Run("grant and list", func(t *testing.T) { grantAndList(t, newRepository(t)) })
	t.Run("end a period", func(t *testing.T) { endAdministrator(t, newRepository(t)) })
	t.Run("never without administrators", func(t *testing.T) { endTheLast(t, newRepository(t)) })
	t.Run("concurrent ends", func(t *testing.T) {
		for round := range 20 {
			t.Run(fmt.Sprint(round), func(t *testing.T) { concurrentEnds(t, newRepository(t)) })
		}
	})
}

func granted(id, userID string, from time.Time) domain.Administrator {
	return domain.Administrator{ID: id, UserID: userID, From: from, GrantedBy: "U-0"}
}

func grantAdministrators(t *testing.T, r output.AdministratorPeriodRepository, administrators ...domain.Administrator) {
	t.Helper()
	for _, a := range administrators {
		if err := r.Grant(t.Context(), a); err != nil {
			t.Fatalf("Grant(%s) = %v", a.ID, err)
		}
	}
}

func grantAndList(t *testing.T, r output.AdministratorPeriodRepository) {
	ctx := t.Context()
	from := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	first, second := granted("A-1", "U-1", from), granted("A-2", "U-2", from.Add(time.Minute))
	grantAdministrators(t, r, second, first)
	if err := r.Grant(ctx, granted("A-3", "U-1", from)); !errors.Is(err, domain.ErrAlreadyAdministrator) {
		t.Errorf("Grant() to an administrator = %v, want ErrAlreadyAdministrator", err)
	}
	got, err := r.ListActive(ctx, from.Add(time.Hour))
	if want := []domain.Administrator{first, second}; err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("ListActive() = %+v, %v; want %+v", got, err, want)
	}
	if got, err := r.ListActive(ctx, from); err != nil || len(got) != 1 || got[0].ID != "A-1" {
		t.Errorf("ListActive() before the second starts = %+v, %v; want only A-1", got, err)
	}
	if got, err := r.IsAdministrator(ctx, "U-2"); err != nil || !got {
		t.Errorf("IsAdministrator(U-2) = %v, %v; want true", got, err)
	}
}

func endAdministrator(t *testing.T, r output.AdministratorPeriodRepository) {
	ctx := t.Context()
	from := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	grantAdministrators(t, r, granted("A-1", "U-1", from), granted("A-2", "U-2", from))
	end := from.Add(time.Minute)
	if err := r.End(ctx, "A-9", end, "U-1"); !errors.Is(err, domain.ErrAdministratorNotFound) {
		t.Errorf("End() of no period = %v, want ErrAdministratorNotFound", err)
	}
	if err := r.End(ctx, "A-2", end, "U-1"); err != nil {
		t.Fatalf("End(A-2) = %v", err)
	}
	if err := r.End(ctx, "A-2", end.Add(time.Minute), "U-1"); !errors.Is(err, domain.ErrAdministratorNotFound) {
		t.Errorf("End() of an ended period = %v, want ErrAdministratorNotFound", err)
	}
	if got, err := r.IsAdministrator(ctx, "U-2"); err != nil || got {
		t.Errorf("IsAdministrator() of an ended administrator = %v, %v; want false", got, err)
	}
	if got, err := r.ListActive(ctx, end); err != nil || len(got) != 1 || got[0].ID != "A-1" {
		t.Errorf("ListActive() after the end = %+v, %v; want only A-1", got, err)
	}
	if err := r.Grant(ctx, granted("A-3", "U-2", end)); err != nil {
		t.Errorf("Grant() again after the end = %v", err)
	}
}

func endTheLast(t *testing.T, r output.AdministratorPeriodRepository) {
	ctx := t.Context()
	from := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	grantAdministrators(t, r, granted("A-1", "U-1", from), granted("A-2", "U-2", from.Add(2*time.Hour)))
	// A-2 starts later, so it does not count yet.
	if err := r.End(ctx, "A-1", from.Add(time.Minute), "U-1"); !errors.Is(err, domain.ErrLastAdministrator) {
		t.Errorf("End() of the only active administrator = %v, want ErrLastAdministrator", err)
	}
	if got, err := r.IsAdministrator(ctx, "U-1"); err != nil || !got {
		t.Errorf("IsAdministrator(U-1) after a refused end = %v, %v; want true", got, err)
	}
}

// concurrentEnds has two administrators end each other at once: one of them
// must remain.
func concurrentEnds(t *testing.T, r output.AdministratorPeriodRepository) {
	ctx := t.Context()
	from := time.Now().UTC().Add(-time.Hour)
	grantAdministrators(t, r, granted("A-1", "U-1", from), granted("A-2", "U-2", from))
	end := from.Add(time.Minute)
	var wg sync.WaitGroup
	wg.Go(func() { _ = r.End(ctx, "A-1", end, "U-2") })
	wg.Go(func() { _ = r.End(ctx, "A-2", end, "U-1") })
	wg.Wait()
	if got, err := r.ListActive(ctx, end); err != nil || len(got) != 1 {
		t.Errorf("ListActive() after concurrent ends = %+v, %v; want one", got, err)
	}
}
