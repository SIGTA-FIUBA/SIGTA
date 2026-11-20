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

// TestMembershipRepository runs the contract of output.MembershipRepository.
// newRepository returns an empty repository each time it is called.
func TestMembershipRepository(t *testing.T, newRepository func(t *testing.T) output.MembershipRepository) {
	t.Run("round trip", func(t *testing.T) { membershipRoundTrip(t, newRepository(t)) })
	t.Run("lists in the order granted", func(t *testing.T) { listMemberships(t, newRepository(t)) })
	t.Run("end a membership", func(t *testing.T) { endMembership(t, newRepository(t)) })
	t.Run("only the validity of a membership changes", func(t *testing.T) { changeOnlyValidity(t, newRepository(t)) })
	t.Run("concurrent grants", func(t *testing.T) { concurrentGrants(t, newRepository(t)) })
}

func membership(id, userID, area string) domain.Membership {
	return domain.Membership{
		ID:          id,
		UserID:      userID,
		Area:        area,
		Roles:       []string{"reviewer"},
		Permissions: []string{"approve", "review"},
		Scope:       []int64{3, 1},
		From:        at,
		GrantedBy:   "admin",
	}
}

func grantAll(t *testing.T, r output.MembershipRepository, memberships ...domain.Membership) {
	t.Helper()
	for _, m := range memberships {
		if err := r.Save(t.Context(), m); err != nil {
			t.Fatalf("Save(%s) = %v", m.ID, err)
		}
	}
}

func membershipRoundTrip(t *testing.T, r output.MembershipRepository) {
	if _, err := r.Get(t.Context(), "M-1"); !errors.Is(err, domain.ErrMembershipNotFound) {
		t.Errorf("Get() of an empty repository = %v, want ErrMembershipNotFound", err)
	}
	ended := membership("M-1", "U-1", "office")
	ended.Until, ended.EndedBy = at.Add(time.Hour), "otro"
	bare := domain.Membership{ID: "M-2", UserID: "U-1", Area: "office", Permissions: []string{"review"}, From: at, GrantedBy: "admin"}
	grantAll(t, r, ended, bare)
	for _, want := range []domain.Membership{ended, bare} {
		got, err := r.Get(t.Context(), want.ID)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("Get(%s) = %+v, %v; want %+v", want.ID, got, err, want)
		}
	}

	list, err := r.ListByUser(t.Context(), "U-1")
	if err != nil || len(list) == 0 {
		t.Fatalf("ListByUser() = %+v, %v", list, err)
	}
	list[0].Roles[0] = "Cambiado"
	list[0].Scope[0] = 99
	if again, err := r.Get(t.Context(), "M-1"); err != nil || !reflect.DeepEqual(again, ended) {
		t.Error("the repository shares the membership with the caller")
	}
}

func listMemberships(t *testing.T, r output.MembershipRepository) {
	later := membership("M-1", "U-1", "committee")
	later.From = at.Add(24 * time.Hour)
	grantAll(t, r, later, membership("M-3", "U-2", "office"), membership("M-2", "U-1", "office"))
	byUser, err := r.ListByUser(t.Context(), "U-1")
	if want := []domain.Membership{later, membership("M-2", "U-1", "office")}; err != nil || !reflect.DeepEqual(byUser, want) {
		t.Errorf("ListByUser(U-1) = %+v, %v; want %+v", byUser, err, want)
	}
	byArea, err := r.ListByArea(t.Context(), "office")
	if want := []domain.Membership{membership("M-3", "U-2", "office"), membership("M-2", "U-1", "office")}; err != nil || !reflect.DeepEqual(byArea, want) {
		t.Errorf("ListByArea(office) = %+v, %v; want %+v", byArea, err, want)
	}
	if none, err := r.ListByUser(t.Context(), "U-3"); err != nil || len(none) != 0 {
		t.Errorf("ListByUser(U-3) = %+v, %v; want none", none, err)
	}
	if none, err := r.ListByArea(t.Context(), "archive"); err != nil || len(none) != 0 {
		t.Errorf("ListByArea(archive) = %+v, %v; want none", none, err)
	}
}

// endMembership ends a membership that has not started yet, which also moves
// its start.
func endMembership(t *testing.T, r output.MembershipRepository) {
	m := membership("M-1", "U-1", "office")
	m.From = at.Add(24 * time.Hour)
	grantAll(t, r, m)
	m.From, m.Until, m.EndedBy = at, at, "otro"
	grantAll(t, r, m)
	if got, err := r.Get(t.Context(), "M-1"); err != nil || !reflect.DeepEqual(got, m) {
		t.Errorf("Get() of an ended membership = %+v, %v; want %+v", got, err, m)
	}
	if list, err := r.ListByUser(t.Context(), "U-1"); err != nil || len(list) != 1 {
		t.Errorf("ListByUser() = %+v, %v; want the ended membership only", list, err)
	}
}

func changeOnlyValidity(t *testing.T, r output.MembershipRepository) {
	m := membership("M-1", "U-1", "office")
	grantAll(t, r, m)
	changed := membership("M-1", "U-2", "committee")
	changed.Roles, changed.Permissions, changed.Scope, changed.GrantedBy = []string{"chair"}, []string{"sign"}, []int64{7}, "otro"
	changed.From, changed.Until, changed.EndedBy = at.Add(-time.Hour), at, "otro"
	grantAll(t, r, changed)
	m.From, m.Until, m.EndedBy = changed.From, changed.Until, changed.EndedBy
	if got, err := r.Get(t.Context(), "M-1"); err != nil || !reflect.DeepEqual(got, m) {
		t.Errorf("Get() after saving other fields = %+v, %v; want %+v", got, err, m)
	}
}

func concurrentGrants(t *testing.T, r output.MembershipRepository) {
	const grants = 10
	var wg sync.WaitGroup
	errs := make([]error, grants)
	for i := range grants {
		wg.Go(func() { errs[i] = r.Save(t.Context(), membership(fmt.Sprintf("M-%d", i), "U-1", "office")) })
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("Save() number %d = %v", i, err)
		}
	}
	if list, err := r.ListByUser(t.Context(), "U-1"); err != nil || len(list) != grants {
		t.Errorf("ListByUser() = %d memberships, %v; want %d", len(list), err, grants)
	}
}
