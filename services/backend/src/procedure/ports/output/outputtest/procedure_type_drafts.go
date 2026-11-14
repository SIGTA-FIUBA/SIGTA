package outputtest

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"backend/src/procedure/domain/definition/definitiontest"
	"backend/src/procedure/domain/draft"
	"backend/src/procedure/ports/output"
)

// TestProcedureTypeDraftRepository runs the contract of output.ProcedureTypeDraftRepository.
// newRepository returns an empty store each time it is called.
func TestProcedureTypeDraftRepository(t *testing.T, newRepository func(t *testing.T) output.ProcedureTypeDraftRepository) {
	t.Run("round trip", func(t *testing.T) { draftRoundTrip(t, newRepository(t)) })
	t.Run("one draft per procedure type", func(t *testing.T) { oneDraftPerType(t, newRepository(t)) })
	t.Run("changes check the revision", func(t *testing.T) { revisionChecks(t, newRepository(t)) })
	t.Run("concurrent updates", func(t *testing.T) { concurrentUpdates(t, newRepository(t)) })
}

// Draft builds a proposed draft that was rejected once, so that every field
// has a value. Times have microseconds, which PostgreSQL keeps.
func Draft(procedureType int64) draft.Draft {
	at := time.Date(2026, 11, 14, 10, 0, 0, 123456000, time.UTC)
	d := draft.New(procedureType, 2, definitiontest.Definition(), "ana", at)
	d.Status = draft.StatusProposed
	d.ProposedBy, d.ProposedAt = "ana", at.Add(time.Hour)
	d.Rejection = draft.Rejection{By: "beto", At: at.Add(time.Minute), Reason: "Falta el plazo"}
	return d
}

func draftRoundTrip(t *testing.T, s output.ProcedureTypeDraftRepository) {
	ctx := t.Context()
	if _, err := s.Get(ctx, 1); !errors.Is(err, output.ErrDraftNotFound) {
		t.Errorf("Get() of an empty store = %v, want ErrDraftNotFound", err)
	}
	full := Draft(2)
	editing := draft.New(1, 0, definitiontest.Definition(), "ana", full.EditedAt)
	for _, d := range []draft.Draft{full, editing} {
		if err := s.Add(ctx, d); err != nil {
			t.Fatal(err)
		}
		got, err := s.Get(ctx, d.ProcedureTypeID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, d) {
			t.Errorf("Get() = %+v, want %+v", got, d)
		}
	}
	list, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].ProcedureTypeID != 1 || list[1].ProcedureTypeID != 2 {
		t.Errorf("List() = %+v, want drafts 1 and 2 in order", list)
	}

	// What the caller does with its copy does not reach the store.
	list[0].Definition.Steps[0].Name = "Cambiado"
	got, err := s.Get(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Definition.Steps[0].Name == "Cambiado" {
		t.Error("the store shares the definition with the caller")
	}
}

func oneDraftPerType(t *testing.T, s output.ProcedureTypeDraftRepository) {
	ctx := t.Context()
	if err := s.Add(ctx, Draft(1)); err != nil {
		t.Fatal(err)
	}
	if err := s.Add(ctx, Draft(1)); !errors.Is(err, output.ErrDraftExists) {
		t.Errorf("Add() of a second draft = %v, want ErrDraftExists", err)
	}
}

func revisionChecks(t *testing.T, s output.ProcedureTypeDraftRepository) {
	ctx := t.Context()
	d := Draft(1)
	if err := s.Update(ctx, d); !errors.Is(err, output.ErrDraftNotFound) {
		t.Errorf("Update() without a draft = %v, want ErrDraftNotFound", err)
	}
	if err := s.Remove(ctx, 1, 1); !errors.Is(err, output.ErrDraftNotFound) {
		t.Errorf("Remove() without a draft = %v, want ErrDraftNotFound", err)
	}
	if err := s.Add(ctx, d); err != nil {
		t.Fatal(err)
	}
	next := d
	next.Revision = 2
	next.Status, next.ProposedBy, next.ProposedAt = draft.StatusEditing, "", time.Time{}
	next.Rejection = draft.Rejection{}
	if err := s.Update(ctx, next); err != nil {
		t.Fatalf("Update() = %v", err)
	}
	got, err := s.Get(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, next) {
		t.Errorf("Get() after Update() = %+v, want %+v", got, next)
	}
	if err := s.Update(ctx, next); !errors.Is(err, output.ErrDraftChanged) {
		t.Errorf("Update() from an old revision = %v, want ErrDraftChanged", err)
	}
	if err := s.Remove(ctx, 1, 1); !errors.Is(err, output.ErrDraftChanged) {
		t.Errorf("Remove() of an old revision = %v, want ErrDraftChanged", err)
	}
	if err := s.Remove(ctx, 1, 2); err != nil {
		t.Fatalf("Remove() = %v", err)
	}
	if _, err := s.Get(ctx, 1); !errors.Is(err, output.ErrDraftNotFound) {
		t.Errorf("Get() after Remove() = %v, want ErrDraftNotFound", err)
	}
}

// concurrentUpdates checks that of several changes made on the same copy,
// only one is kept.
func concurrentUpdates(t *testing.T, s output.ProcedureTypeDraftRepository) {
	ctx := t.Context()
	d := Draft(1)
	if err := s.Add(ctx, d); err != nil {
		t.Fatal(err)
	}
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		kept int
	)
	for range publishers {
		wg.Go(func() {
			next := d
			next.Revision = 2
			err := s.Update(ctx, next)
			if err != nil && !errors.Is(err, output.ErrDraftChanged) {
				t.Errorf("Update() = %v, want nil or ErrDraftChanged", err)
			}
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				kept++
			}
		})
	}
	wg.Wait()
	if kept != 1 {
		t.Errorf("%d updates were kept, want 1", kept)
	}
}
