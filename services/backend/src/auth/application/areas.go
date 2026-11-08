package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
	"backend/src/auth/ports/output"
)

var _ input.Areas = (*AreaService)(nil)

const (
	actionCreateArea = "create area"
	actionDeleteArea = "delete area"
)

type AreaService struct {
	administratorGuard
	areas       output.AreaRepository
	memberships output.MembershipRepository
	clock       func() time.Time
}

func NewAreaService(areas output.AreaRepository, administrators output.Administrators, memberships output.MembershipRepository, clock func() time.Time) *AreaService {
	return &AreaService{
		administratorGuard: administratorGuard{administrators: administrators},
		areas:              areas,
		memberships:        memberships,
		clock:              clock,
	}
}

func (s *AreaService) List(ctx context.Context) ([]domain.Area, error) {
	areas, err := s.areas.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list areas: %w", err)
	}
	return areas, nil
}

// Create refuses the ID of a deleted area too: memberships and published
// versions may still name it, and they would end up naming another area.
func (s *AreaService) Create(ctx context.Context, administratorID string, a domain.Area) (domain.Area, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionCreateArea); err != nil {
		return domain.Area{}, err
	}
	area := domain.Area{ID: a.ID, Name: strings.TrimSpace(a.Name)}
	if err := s.validate(ctx, area); err != nil {
		return domain.Area{}, fmt.Errorf("%w: %w", domain.ErrInvalidArea, err)
	}
	if err := s.areas.Save(ctx, area); err != nil {
		return domain.Area{}, fmt.Errorf("create area %s: %w", area.ID, err)
	}
	return area, nil
}

// Delete keeps the area, marked as deleted, so that history still names it.
// An area with memberships that have not ended cannot be deleted: its members
// would keep a role in an area that no longer exists.
func (s *AreaService) Delete(ctx context.Context, administratorID, areaID string) error {
	if err := s.requireAdministrator(ctx, administratorID, actionDeleteArea); err != nil {
		return err
	}
	area, err := s.areas.Get(ctx, areaID)
	if err != nil {
		return fmt.Errorf("area %s: %w", areaID, err)
	}
	if area.Deleted() {
		return fmt.Errorf("area %s: %w", areaID, domain.ErrAreaNotFound)
	}
	memberships, err := s.memberships.ListByArea(ctx, areaID)
	if err != nil {
		return fmt.Errorf("memberships of area %s: %w", areaID, err)
	}
	now := s.clock().UTC()
	if slices.ContainsFunc(memberships, func(m domain.Membership) bool { return m.Until.IsZero() || m.Until.After(now) }) {
		return fmt.Errorf("%w: area %s has memberships that have not ended", domain.ErrInvalidArea, areaID)
	}
	area.DeletedAt = now
	area.DeletedBy = administratorID
	if err := s.areas.Save(ctx, area); err != nil {
		return fmt.Errorf("delete area %s: %w", areaID, err)
	}
	return nil
}

func (s *AreaService) validate(ctx context.Context, a domain.Area) error {
	if !domain.ValidAreaID(a.ID) {
		return fmt.Errorf("id %q has to be lowercase letters, digits and underscores, and start with a letter", a.ID)
	}
	if a.Name == "" {
		return errors.New("an area needs a name")
	}
	_, err := s.areas.Get(ctx, a.ID)
	switch {
	case err == nil:
		return fmt.Errorf("id %q is already used", a.ID)
	case !errors.Is(err, domain.ErrAreaNotFound):
		return fmt.Errorf("area %s: %w", a.ID, err)
	}
	active, err := s.areas.List(ctx)
	if err != nil {
		return fmt.Errorf("list areas: %w", err)
	}
	if slices.ContainsFunc(active, func(other domain.Area) bool { return strings.EqualFold(other.Name, a.Name) }) {
		return fmt.Errorf("name %q is already used", a.Name)
	}
	return nil
}
