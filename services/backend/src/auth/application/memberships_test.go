package application_test

import (
	"errors"
	"slices"
	"strconv"
	"testing"
	"time"

	"backend/src/auth/adapters/memory"
	"backend/src/auth/application"
	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
)

var office = domain.Area{ID: "office", Name: "Mesa", Permissions: []domain.Permission{{ID: "review", Name: "Revisar"}, {ID: "approve", Name: "Aprobar"}}}

type membershipFixture struct {
	memberships *memory.MembershipRepository
	service     *application.MembershipService
	now         time.Time
}

func newMembershipFixture(memberships ...domain.Membership) *membershipFixture {
	f := &membershipFixture{memberships: memory.NewMembershipRepository(memberships...), now: time.Date(2030, 3, 1, 12, 0, 0, 0, time.UTC)}
	ids := 0
	newID := func() string {
		ids++
		return "new" + strconv.Itoa(ids)
	}
	f.service = application.NewMembershipService(memory.NewAreaRepository(office), memory.NewAdministratorRepository("admin"), f.memberships, memory.NewUserRepository(applicant, operative), newID, func() time.Time { return f.now })
	return f
}

func TestChangingAMembershipEndsItAndOpensAnotherWithTheSameScope(t *testing.T) {
	until := time.Date(2030, 12, 1, 0, 0, 0, 0, time.UTC)
	f := newMembershipFixture(domain.Membership{ID: "m1", UserID: operative.ID, Area: "office", Permissions: []string{"review"}, Scope: []int64{2}, From: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC), Until: until, GrantedBy: "other"})

	m, err := f.service.Change(t.Context(), "admin", "m1", nil, []string{"review", "approve"})
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != "new1" || !slices.Equal(m.Scope, []int64{2}) || !m.From.Equal(f.now) || !m.Until.Equal(until) || m.GrantedBy != "admin" || !m.GrantedAt.Equal(f.now) {
		t.Errorf("new membership = %+v", m)
	}
	old, err := f.memberships.Get(t.Context(), "m1")
	if err != nil || !old.Until.Equal(f.now) || old.EndedBy != "admin" {
		t.Errorf("old membership = %+v, %v; want it ended now by admin", old, err)
	}

	_, err = f.service.Change(t.Context(), "admin", "m1", nil, []string{"approve"})
	if !errors.Is(err, domain.ErrMembershipEnded) {
		t.Errorf("Change() of an ended membership = %v, want ErrMembershipEnded", err)
	}
	_, err = f.service.Change(t.Context(), "admin", "new1", nil, []string{"approve", "review"})
	if !errors.Is(err, domain.ErrInvalidMembership) {
		t.Errorf("Change() to the same permissions = %v, want ErrInvalidMembership", err)
	}
}

func TestChangingAFutureMembershipKeepsItsStart(t *testing.T) {
	from := time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC)
	f := newMembershipFixture(domain.Membership{ID: "m1", UserID: operative.ID, Area: "office", Permissions: []string{"review"}, From: from, GrantedBy: "admin"})

	m, err := f.service.Change(t.Context(), "admin", "m1", nil, []string{"approve"})
	if err != nil || !m.From.Equal(from) {
		t.Fatalf("Change() = %+v, %v; want it to start at %v", m, err, from)
	}
	memberships, err := f.service.ListByArea(t.Context(), "admin", "office")
	if err != nil || len(memberships) != 1 || memberships[0].ID != "new1" {
		t.Errorf("ListByArea() = %+v, %v; want only the new membership", memberships, err)
	}
}

// A membership is for someone who entered SIGTA: an id typed by mistake is
// refused instead of granting access to nobody.
func TestAMembershipNeedsSomeoneWhoSignedIn(t *testing.T) {
	f := newMembershipFixture()

	_, err := f.service.Grant(t.Context(), "admin", input.MembershipGrant{UserID: "nadie", Area: "office", Permissions: []string{"review"}})
	if !errors.Is(err, domain.ErrInvalidMembership) || !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Grant() to nobody = %v, want ErrInvalidMembership and ErrUserNotFound", err)
	}
	if _, err := f.service.Grant(t.Context(), "admin", input.MembershipGrant{UserID: operative.ID, Area: "office", Permissions: []string{"review"}}); err != nil {
		t.Errorf("Grant() to someone who signed in = %v", err)
	}
}

func TestOnlyAnAdministratorChangesOrListsMemberships(t *testing.T) {
	f := newMembershipFixture(domain.Membership{ID: "m1", UserID: operative.ID, Area: "office", Permissions: []string{"review"}, GrantedBy: "admin"})

	if _, err := f.service.Change(t.Context(), operative.ID, "m1", nil, []string{"approve"}); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("Change() by staff = %v, want ErrAccessDenied", err)
	}
	if _, err := f.service.ListByArea(t.Context(), operative.ID, "office"); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("ListByArea() by staff = %v, want ErrAccessDenied", err)
	}
	if _, err := f.service.ListByArea(t.Context(), "admin", "dean"); !errors.Is(err, domain.ErrAreaNotFound) {
		t.Errorf("ListByArea() of an unknown area = %v, want ErrAreaNotFound", err)
	}
}
