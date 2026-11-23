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
	for _, c := range []struct {
		name string
		test func(*testing.T, output.MembershipRepository)
	}{
		{"round trip", membershipRoundTrip},
		{"lists in the order granted", listMemberships},
		{"end a membership", endMembership},
		{"only the validity of a membership changes", changeOnlyValidity},
		{"concurrent grants", concurrentGrants},
		{"replace a membership", replaceMembership},
		{"a membership that overlaps another is not granted", refuseOverlap},
		{"an ended membership is not ended again", refuseEndingTwice},
		{"of overlapping grants at once only one is saved", concurrentOverlappingGrants},
		{"of changes to one membership at once only one is saved", concurrentReplaces},
	} {
		t.Run(c.name, func(t *testing.T) { c.test(t, newRepository(t)) })
	}
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
		GrantedAt:   at.Add(-time.Hour),
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
		wg.Go(func() {
			m := membership(fmt.Sprintf("M-%d", i), "U-1", "office")
			m.Scope = []int64{int64(i)}
			errs[i] = r.Save(t.Context(), m)
		})
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

// Replace ends one membership and grants the one that follows it, both or
// neither.
func replaceMembership(t *testing.T, r output.MembershipRepository) {
	old := membership("M-1", "U-1", "office")
	grantAll(t, r, old)
	ended := old
	ended.Until, ended.EndedBy = at.Add(time.Hour), "admin"
	granted := membership("M-2", "U-1", "office")
	granted.Roles, granted.Permissions = nil, []string{"review"}
	granted.From, granted.GrantedAt = ended.Until, ended.Until
	if err := r.Replace(t.Context(), ended, granted); err != nil {
		t.Fatalf("Replace() = %v", err)
	}
	got, err := r.ListByUser(t.Context(), "U-1")
	if want := []domain.Membership{ended, granted}; err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("ListByUser() after Replace() = %+v, %v; want %+v", got, err, want)
	}
}

func refuseOverlap(t *testing.T, r output.MembershipRepository) {
	grantAll(t, r, membership("M-1", "U-1", "office"))
	again := membership("M-2", "U-1", "office")
	again.From = at.Add(24 * time.Hour)
	if err := r.Save(t.Context(), again); !errors.Is(err, domain.ErrMembershipOverlaps) {
		t.Errorf("Save() of an overlapping membership = %v, want ErrMembershipOverlaps", err)
	}
	ended := membership("M-1", "U-1", "office")
	ended.Until, ended.EndedBy = at.Add(time.Hour), "admin"
	granted := membership("M-3", "U-1", "office")
	granted.From = at
	if err := r.Replace(t.Context(), ended, granted); !errors.Is(err, domain.ErrMembershipOverlaps) {
		t.Errorf("Replace() with one that overlaps the ended = %v, want ErrMembershipOverlaps", err)
	}
	if list, err := r.ListByUser(t.Context(), "U-1"); err != nil || len(list) != 1 || !list[0].Until.IsZero() {
		t.Errorf("ListByUser() = %+v, %v; want M-1 alone and not ended", list, err)
	}
}

func refuseEndingTwice(t *testing.T, r output.MembershipRepository) {
	m := membership("M-1", "U-1", "office")
	grantAll(t, r, m)
	m.Until, m.EndedBy = at.Add(time.Hour), "admin"
	grantAll(t, r, m)
	later := m
	later.Until, later.EndedBy = at.Add(2*time.Hour), "otro"
	if err := r.Save(t.Context(), later); !errors.Is(err, domain.ErrMembershipEnded) {
		t.Errorf("Save() ending an ended membership = %v, want ErrMembershipEnded", err)
	}
	granted := membership("M-2", "U-1", "office")
	granted.From = later.Until
	if err := r.Replace(t.Context(), later, granted); !errors.Is(err, domain.ErrMembershipEnded) {
		t.Errorf("Replace() of an ended membership = %v, want ErrMembershipEnded", err)
	}
	if got, err := r.Get(t.Context(), "M-1"); err != nil || !reflect.DeepEqual(got, m) {
		t.Errorf("Get() = %+v, %v; want %+v", got, err, m)
	}
	if _, err := r.Get(t.Context(), "M-2"); !errors.Is(err, domain.ErrMembershipNotFound) {
		t.Errorf("Get(M-2) = %v, want ErrMembershipNotFound", err)
	}
}

func concurrentOverlappingGrants(t *testing.T, r output.MembershipRepository) {
	const grants = 10
	var wg sync.WaitGroup
	errs := make([]error, grants)
	for i := range grants {
		wg.Go(func() { errs[i] = r.Save(t.Context(), membership(fmt.Sprintf("M-%d", i), "U-1", "office")) })
	}
	wg.Wait()
	if saved := countSaved(t, errs, domain.ErrMembershipOverlaps); saved != 1 {
		t.Errorf("%d grants saved, want 1", saved)
	}
}

// concurrentReplaces changes one membership several times at once, and ends
// it at the same time: only one of them can end it.
func concurrentReplaces(t *testing.T, r output.MembershipRepository) {
	old := membership("M-1", "U-1", "office")
	grantAll(t, r, old)
	const changes = 10
	var wg sync.WaitGroup
	errs := make([]error, changes+1)
	for i := range changes {
		wg.Go(func() {
			ended := old
			ended.Until, ended.EndedBy = at.Add(time.Hour), "admin"
			granted := membership(fmt.Sprintf("M-%d", i+2), "U-1", "office")
			granted.From = ended.Until
			errs[i] = r.Replace(t.Context(), ended, granted)
		})
	}
	wg.Go(func() {
		ended := old
		ended.Until, ended.EndedBy = at.Add(time.Hour), "admin"
		errs[changes] = r.Save(t.Context(), ended)
	})
	wg.Wait()
	if saved := countSaved(t, errs, domain.ErrMembershipEnded); saved != 1 {
		t.Errorf("%d changes saved, want 1", saved)
	}
	list, err := r.ListByUser(t.Context(), "U-1")
	if err != nil {
		t.Fatalf("ListByUser() = %v", err)
	}
	want := 2
	if errs[changes] == nil {
		want = 1
	}
	if len(list) != want || list[0].Until.IsZero() {
		t.Errorf("ListByUser() = %+v; want M-1 ended and %d memberships", list, want)
	}
}

// countSaved returns how many calls succeeded, failing the test for an error
// other than refused.
func countSaved(t *testing.T, errs []error, refused error) int {
	t.Helper()
	saved := 0
	for i, err := range errs {
		switch {
		case err == nil:
			saved++
		case !errors.Is(err, refused):
			t.Errorf("call number %d = %v, want nil or %v", i, err, refused)
		}
	}
	return saved
}
