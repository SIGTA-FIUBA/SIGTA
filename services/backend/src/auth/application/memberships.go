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

var _ input.Memberships = (*MembershipService)(nil)

const (
	actionGrant = "grant membership"
	actionEnd   = "end membership"
)

type MembershipService struct {
	areas          output.AreaRepository
	administrators output.Administrators
	memberships    output.MembershipRepository
	deniedAttempts output.DeniedAttemptLog
	newID          func() string
	clock          func() time.Time
}

func NewMembershipService(areas output.AreaRepository, administrators output.Administrators, memberships output.MembershipRepository, deniedAttempts output.DeniedAttemptLog, newID func() string, clock func() time.Time) *MembershipService {
	return &MembershipService{areas: areas, administrators: administrators, memberships: memberships, deniedAttempts: deniedAttempts, newID: newID, clock: clock}
}

func (s *MembershipService) Permissions(ctx context.Context, userID string) (domain.Permissions, error) {
	administrator, err := s.administrators.IsAdministrator(ctx, userID)
	if err != nil {
		return domain.Permissions{}, fmt.Errorf("permissions of %s: %w", userID, err)
	}
	memberships, err := s.memberships.ListByUser(ctx, userID)
	if err != nil {
		return domain.Permissions{}, fmt.Errorf("permissions of %s: %w", userID, err)
	}
	return domain.NewPermissions(userID, administrator, memberships, s.clock()), nil
}

func (s *MembershipService) Grant(ctx context.Context, administratorID string, g input.MembershipGrant) (domain.Membership, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionGrant); err != nil {
		return domain.Membership{}, err
	}
	m := domain.Membership{
		ID:        s.newID(),
		UserID:    g.UserID,
		Area:      g.Area,
		Profile:   g.Profile,
		Scope:     slices.Clone(g.Scope),
		From:      g.From,
		Until:     g.Until,
		GrantedBy: administratorID,
	}
	if m.From.IsZero() {
		m.From = s.clock().UTC()
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
	now := s.clock().UTC()
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

func (s *MembershipService) requireAdministrator(ctx context.Context, userID, action string) error {
	administrator, err := s.administrators.IsAdministrator(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s %s: %w", userID, action, err)
	}
	if administrator {
		return nil
	}
	if err := s.deniedAttempts.Record(ctx, domain.DeniedAttempt{UserID: userID, Action: action, At: s.clock().UTC()}); err != nil {
		return fmt.Errorf("record denied attempt: %w", err)
	}
	return fmt.Errorf("%s %s: %w", userID, action, domain.ErrAccessDenied)
}

// validate leaves the scope unchecked: a procedure type that does not exist
// matches no procedure, so checking it would only make auth depend on
// procedure.
func (s *MembershipService) validate(ctx context.Context, m domain.Membership) error {
	if m.UserID == "" {
		return errors.New("a membership needs a person")
	}
	if !m.Profile.Known() {
		return fmt.Errorf("unknown profile %q", m.Profile)
	}
	if !m.Until.IsZero() && !m.Until.After(m.From) {
		return errors.New("a membership has to end after it starts")
	}
	areas, err := s.areas.List(ctx)
	if err != nil {
		return fmt.Errorf("list areas: %w", err)
	}
	if !slices.ContainsFunc(areas, func(a domain.Area) bool { return a.ID == m.Area }) {
		return fmt.Errorf("unknown area %q", m.Area)
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
