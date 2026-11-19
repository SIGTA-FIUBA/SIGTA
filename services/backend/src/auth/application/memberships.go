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
	actionGrant = "grant membership"
	actionEnd   = "end membership"
)

type MembershipService struct {
	administratorGuard
	areas       output.AreaRepository
	memberships output.MembershipRepository
	newID       func() string
	now         func() time.Time
}

func NewMembershipService(areas output.AreaRepository, administrators output.AdministratorRepository, memberships output.MembershipRepository, newID func() string, now func() time.Time) *MembershipService {
	return &MembershipService{
		administratorGuard: administratorGuard{administrators: administrators},
		areas:              areas,
		memberships:        memberships,
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
	}
	if m.From.IsZero() {
		m.From = s.now().UTC()
	}
	if err := s.validate(ctx, m); err != nil {
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
	m, err := s.memberships.Get(ctx, membershipID)
	if err != nil {
		return fmt.Errorf("membership %s: %w", membershipID, err)
	}
	now := s.now().UTC()
	if !m.Until.IsZero() && !now.Before(m.Until) {
		return fmt.Errorf("%w: membership %s already ended", domain.ErrInvalidMembership, membershipID)
	}
	m.Until = now
	if m.From.After(now) {
		m.From = now
	}
	m.EndedBy = administratorID
	if err := s.memberships.Save(ctx, m); err != nil {
		return fmt.Errorf("end membership %s: %w", membershipID, err)
	}
	return nil
}

// validate leaves the scope unchecked: a procedure type that does not exist
// matches no procedure, so checking it would only make auth depend on
// procedure.
func (s *MembershipService) validate(ctx context.Context, m domain.Membership) error {
	if m.UserID == "" {
		return errors.New("a membership needs a person")
	}
	if len(m.Roles) == 0 && len(m.Permissions) == 0 {
		return errors.New("a membership needs at least one role or permission")
	}
	if !m.Until.IsZero() && !m.Until.After(m.From) {
		return errors.New("a membership has to end after it starts")
	}
	areas, err := s.areas.List(ctx)
	if err != nil {
		return fmt.Errorf("list areas: %w", err)
	}
	i := slices.IndexFunc(areas, func(a domain.Area) bool { return a.ID == m.Area })
	if i < 0 {
		return fmt.Errorf("unknown area %q", m.Area)
	}
	if err := checkListed("permission", m.Permissions, m.Area, areas[i].HasActivePermission); err != nil {
		return err
	}
	if err := checkListed("role", m.Roles, m.Area, areas[i].HasActiveRole); err != nil {
		return err
	}
	existing, err := s.memberships.ListByUser(ctx, m.UserID)
	if err != nil {
		return fmt.Errorf("memberships of %s: %w", m.UserID, err)
	}
	if slices.ContainsFunc(existing, m.Overlaps) {
		return fmt.Errorf("%s already has a membership in area %q with the same scope", m.UserID, m.Area)
	}
	return nil
}

// checkListed checks that each ID exists in the area and appears only once.
func checkListed(kind string, ids []string, area string, exists func(string) bool) error {
	seen := map[string]bool{}
	for _, id := range ids {
		if !exists(id) {
			return fmt.Errorf("area %q has no %s %q", area, kind, id)
		}
		if seen[id] {
			return fmt.Errorf("%s %q is listed twice", kind, id)
		}
		seen[id] = true
	}
	return nil
}
