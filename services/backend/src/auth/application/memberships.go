package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
	"backend/src/auth/ports/output"
)

var (
	_ input.Memberships     = (*MembershipService)(nil)
	_ input.StaffAuthorizer = (*MembershipService)(nil)
)

const (
	actionGrant  = "grant membership"
	actionEnd    = "end membership"
	actionChange = "change membership"
	actionList   = "list memberships"
)

type MembershipService struct {
	administratorGuard
	areas       output.AreaRepository
	memberships output.MembershipRepository
	users       output.UserRepository
	newID       func() string
	now         func() time.Time
}

func NewMembershipService(areas output.AreaRepository, administrators output.AdministratorRepository, memberships output.MembershipRepository, users output.UserRepository, newID func() string, now func() time.Time) *MembershipService {
	return &MembershipService{
		administratorGuard: administratorGuard{administrators: administrators},
		areas:              areas,
		memberships:        memberships,
		users:              users,
		newID:              newID,
		now:                now,
	}
}

func (s *MembershipService) Access(ctx context.Context, userID string) (domain.Access, error) {
	administrator, err := s.administrators.IsAdministrator(ctx, userID)
	if err != nil {
		return domain.Access{}, fmt.Errorf("access of %s: %w", userID, err)
	}
	memberships, err := s.memberships.ListByUser(ctx, userID)
	if err != nil {
		return domain.Access{}, fmt.Errorf("access of %s: %w", userID, err)
	}
	areas, err := s.areas.List(ctx)
	if err != nil {
		return domain.Access{}, fmt.Errorf("access of %s: %w", userID, err)
	}
	return domain.NewAccess(administrator, memberships, areas, s.now()), nil
}

func (s *MembershipService) Holds(ctx context.Context, userID, area string, procedureType int64, permissions []string) (bool, error) {
	access, err := s.Access(ctx, userID)
	if err != nil {
		return false, err
	}
	return access.Holds(area, procedureType, permissions), nil
}

func (s *MembershipService) Scopes(ctx context.Context, userID string) ([]input.AreaScope, error) {
	access, err := s.Access(ctx, userID)
	if err != nil {
		return nil, err
	}
	scopes := make([]input.AreaScope, 0, len(access.Memberships))
	for _, m := range access.Memberships {
		scopes = append(scopes, input.AreaScope{Area: m.Area, Scope: slices.Clone(m.Scope)})
	}
	return scopes, nil
}

func (s *MembershipService) Grant(ctx context.Context, administratorID string, g input.MembershipGrant) (domain.Membership, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionGrant); err != nil {
		return domain.Membership{}, err
	}
	m := domain.Membership{
		ID:          s.newID(),
		UserID:      g.UserID,
		Area:        g.Area,
		Roles:       slices.Clone(g.Roles),
		Permissions: slices.Clone(g.Permissions),
		Scope:       slices.Clone(g.Scope),
		From:        g.From,
		Until:       g.Until,
		GrantedBy:   administratorID,
		GrantedAt:   s.now().UTC(),
	}
	if m.From.IsZero() {
		m.From = m.GrantedAt
	}
	if err := s.validate(ctx, m, ""); err != nil {
		return domain.Membership{}, fmt.Errorf("%w: %w", domain.ErrInvalidMembership, err)
	}
	if err := s.memberships.Save(ctx, m); err != nil {
		return domain.Membership{}, fmt.Errorf("grant membership to %s: %w", m.UserID, err)
	}
	return m, nil
}

// End closes a membership now. It is never deleted: what the person did while
// it held stays attributed to them.
func (s *MembershipService) End(ctx context.Context, administratorID, membershipID string) error {
	if err := s.requireAdministrator(ctx, administratorID, actionEnd); err != nil {
		return err
	}
	m, err := s.unended(ctx, membershipID)
	if err != nil {
		return err
	}
	if err := s.memberships.Save(ctx, ended(m, s.now().UTC(), administratorID)); err != nil {
		return fmt.Errorf("end membership %s: %w", membershipID, err)
	}
	return nil
}

// Change ends the membership and grants another with these roles and
// permissions, for the same person, area, scope and end, so that the history
// tells what the person could do at each moment. The new one starts now, or
// when the old one started if it had not yet.
func (s *MembershipService) Change(ctx context.Context, administratorID, membershipID string, roles, permissions []string) (domain.Membership, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionChange); err != nil {
		return domain.Membership{}, err
	}
	old, err := s.unended(ctx, membershipID)
	if err != nil {
		return domain.Membership{}, err
	}
	if sameIDs(old.Roles, roles) && sameIDs(old.Permissions, permissions) {
		return domain.Membership{}, fmt.Errorf("%w: membership %s already has those roles and permissions", domain.ErrInvalidMembership, membershipID)
	}
	now := s.now().UTC()
	m := domain.Membership{
		ID:          s.newID(),
		UserID:      old.UserID,
		Area:        old.Area,
		Roles:       slices.Clone(roles),
		Permissions: slices.Clone(permissions),
		Scope:       slices.Clone(old.Scope),
		From:        laterOf(now, old.From),
		Until:       old.Until,
		GrantedBy:   administratorID,
		GrantedAt:   now,
	}
	if err := s.validate(ctx, m, old.ID); err != nil {
		return domain.Membership{}, fmt.Errorf("%w: %w", domain.ErrInvalidMembership, err)
	}
	if err := s.memberships.Replace(ctx, ended(old, now, administratorID), m); err != nil {
		return domain.Membership{}, fmt.Errorf("change membership %s: %w", membershipID, err)
	}
	return m, nil
}

// ListByArea returns the memberships of the area that have not ended, active
// or future, in the order they were granted.
func (s *MembershipService) ListByArea(ctx context.Context, administratorID, areaID string) ([]domain.Membership, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionList); err != nil {
		return nil, err
	}
	area, err := s.areas.Get(ctx, areaID)
	if err == nil && area.Deleted() {
		err = domain.ErrAreaNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("area %s: %w", areaID, err)
	}
	memberships, err := s.memberships.ListByArea(ctx, areaID)
	if err != nil {
		return nil, fmt.Errorf("memberships of area %s: %w", areaID, err)
	}
	now := s.now().UTC()
	return slices.DeleteFunc(memberships, func(m domain.Membership) bool { return !m.EndsAfter(now) }), nil
}

func (s *MembershipService) unended(ctx context.Context, membershipID string) (domain.Membership, error) {
	m, err := s.memberships.Get(ctx, membershipID)
	if err != nil {
		return domain.Membership{}, fmt.Errorf("membership %s: %w", membershipID, err)
	}
	if !m.EndsAfter(s.now().UTC()) {
		return domain.Membership{}, fmt.Errorf("%w: %w: %s", domain.ErrInvalidMembership, domain.ErrMembershipEnded, membershipID)
	}
	return m, nil
}

// ended returns m ended at now. One that had not started yet also moves its
// start, so that it never holds.
func ended(m domain.Membership, now time.Time, by string) domain.Membership {
	m.Until, m.EndedBy = now, by
	if m.From.After(now) {
		m.From = now
	}
	return m
}

func laterOf(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}

func sameIDs(a, b []string) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

// validate leaves the scope unchecked: a procedure type that does not exist
// matches no procedure, so checking it would only make auth depend on
// procedure. replaced is the membership that m follows, if any, which does
// not count as an overlap. The repository checks the overlap again as it
// saves, for a grant made meanwhile.
func (s *MembershipService) validate(ctx context.Context, m domain.Membership, replaced string) error {
	if m.UserID == "" {
		return errors.New("a membership needs a person")
	}
	if _, err := s.users.Get(ctx, m.UserID); err != nil {
		return fmt.Errorf("person %s: %w", m.UserID, err)
	}
	if len(m.Roles) == 0 && len(m.Permissions) == 0 {
		return errors.New("a membership needs at least one role or permission")
	}
	if !m.Until.IsZero() && (!m.Until.After(m.From) || !m.Until.After(s.now())) {
		return fmt.Errorf("%w: from %s until %s", domain.ErrMembershipPeriod, m.From, m.Until)
	}
	areas, err := s.areas.List(ctx)
	if err != nil {
		return fmt.Errorf("list areas: %w", err)
	}
	i := slices.IndexFunc(areas, func(a domain.Area) bool { return a.ID == m.Area })
	if i < 0 {
		return fmt.Errorf("%w: %q", domain.ErrAreaNotFound, m.Area)
	}
	if areas[i].Deleted() {
		return fmt.Errorf("%w: %q was removed", domain.ErrAreaNotFound, m.Area)
	}
	if err := checkListed(domain.ErrPermissionNotFound, m.Permissions, m.Area, areas[i].HasActivePermission); err != nil {
		return err
	}
	if err := checkListed(domain.ErrRoleNotFound, m.Roles, m.Area, areas[i].HasActiveRole); err != nil {
		return err
	}
	existing, err := s.memberships.ListByUser(ctx, m.UserID)
	if err != nil {
		return fmt.Errorf("memberships of %s: %w", m.UserID, err)
	}
	existing = slices.DeleteFunc(existing, func(other domain.Membership) bool { return other.ID == replaced })
	if m.OverlapsAny(existing) {
		return fmt.Errorf("%w: %s in area %q", domain.ErrMembershipOverlaps, m.UserID, m.Area)
	}
	return nil
}

// checkListed checks that each ID exists in the area, failing with notFound
// if not, and appears only once.
func checkListed(notFound error, ids []string, area string, exists func(string) bool) error {
	seen := map[string]bool{}
	for _, id := range ids {
		if !exists(id) {
			return fmt.Errorf("%w: %q in area %q", notFound, id, area)
		}
		if seen[id] {
			return fmt.Errorf("%q is listed twice", id)
		}
		seen[id] = true
	}
	return nil
}
