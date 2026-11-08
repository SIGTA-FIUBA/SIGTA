package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.Memberships = (*MembershipService)(nil)

const (
	actionGrant = "grant membership"
	actionEnd   = "end membership"
)

type MembershipService struct {
	gate
	types input.ProcedureTypes
	areas output.AreaCatalog
	newID func() string
}

func NewMembershipService(types input.ProcedureTypes, areas output.AreaCatalog, access output.Access, denied output.DeniedAttempts, newID func() string, now func() time.Time) *MembershipService {
	return &MembershipService{gate: gate{access: access, denied: denied, now: now}, types: types, areas: areas, newID: newID}
}

func (s *MembershipService) Grant(ctx context.Context, administratorID string, g input.MembershipGrant) (domain.Membership, error) {
	if err := s.administrator(ctx, administratorID, actionGrant); err != nil {
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
		m.From = s.now().UTC()
	}
	if err := s.validate(ctx, m); err != nil {
		return domain.Membership{}, fmt.Errorf("%w: %w", domain.ErrInvalidMembership, err)
	}
	if err := s.access.SaveMembership(ctx, m); err != nil {
		return domain.Membership{}, fmt.Errorf("grant membership to %s: %w", m.UserID, err)
	}
	return m, nil
}

func (s *MembershipService) End(ctx context.Context, administratorID, membershipID string) error {
	if err := s.administrator(ctx, administratorID, actionEnd); err != nil {
		return err
	}
	m, err := s.access.Membership(ctx, membershipID)
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
	if err := s.access.SaveMembership(ctx, m); err != nil {
		return fmt.Errorf("end membership %s: %w", membershipID, err)
	}
	return nil
}

func (s *MembershipService) administrator(ctx context.Context, userID, action string) error {
	actor, err := s.actor(ctx, userID)
	if err != nil {
		return err
	}
	if !actor.Administrator {
		return s.deny(ctx, actor, action, "")
	}
	return nil
}

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
	if err := s.staffArea(ctx, m.Area); err != nil {
		return err
	}
	if err := s.published(ctx, m.Scope); err != nil {
		return err
	}
	existing, err := s.access.Memberships(ctx, m.UserID)
	if err != nil {
		return fmt.Errorf("memberships of %s: %w", m.UserID, err)
	}
	if slices.ContainsFunc(existing, m.Overlaps) {
		return fmt.Errorf("%s already has a membership in area %q with the same scope", m.UserID, m.Area)
	}
	return nil
}

func (s *MembershipService) staffArea(ctx context.Context, id string) error {
	areas, err := s.areas.List(ctx)
	if err != nil {
		return fmt.Errorf("list areas: %w", err)
	}
	i := slices.IndexFunc(areas, func(a domain.Area) bool { return a.ID == id })
	if i < 0 {
		return fmt.Errorf("unknown area %q", id)
	}
	if areas[i].Kind != domain.AreaOffice {
		return fmt.Errorf("area %q has no staff", id)
	}
	return nil
}

func (s *MembershipService) published(ctx context.Context, scope []int64) error {
	if len(scope) == 0 {
		return nil
	}
	entries, err := s.types.Catalog(ctx)
	if err != nil {
		return fmt.Errorf("catalog: %w", err)
	}
	for _, procedureType := range scope {
		if !slices.ContainsFunc(entries, func(e input.CatalogEntry) bool { return e.ProcedureType == procedureType }) {
			return fmt.Errorf("procedure type %d is not published", procedureType)
		}
	}
	return nil
}
