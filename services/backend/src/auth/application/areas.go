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
	actionCreateArea       = "create area"
	actionDeleteArea       = "delete area"
	actionCreatePermission = "create permission"
	actionDeletePermission = "delete permission"
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
	for i := range areas {
		areas[i].Permissions = areas[i].ActivePermissions()
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
	memberships, err := s.unendedMemberships(ctx, areaID)
	if err != nil {
		return err
	}
	if len(memberships) > 0 {
		return fmt.Errorf("%w: area %s has memberships that have not ended", domain.ErrInvalidArea, areaID)
	}
	now := s.clock().UTC()
	area.DeletedAt = now
	area.DeletedBy = administratorID
	if err := s.areas.Save(ctx, area); err != nil {
		return fmt.Errorf("delete area %s: %w", areaID, err)
	}
	return nil
}

func (s *AreaService) validate(ctx context.Context, a domain.Area) error {
	if !domain.ValidID(a.ID) {
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

// CreatePermission refuses the ID of a deleted permission of the area too:
// published versions may still ask for it.
func (s *AreaService) CreatePermission(ctx context.Context, administratorID, areaID string, p domain.Permission) (domain.Permission, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionCreatePermission); err != nil {
		return domain.Permission{}, err
	}
	area, err := s.activeArea(ctx, areaID)
	if err != nil {
		return domain.Permission{}, err
	}
	permission := domain.Permission{ID: p.ID, Name: strings.TrimSpace(p.Name)}
	if err := validatePermission(area, permission); err != nil {
		return domain.Permission{}, fmt.Errorf("%w: %w", domain.ErrInvalidPermission, err)
	}
	area.Permissions = append(area.Permissions, permission)
	if err := s.areas.Save(ctx, area); err != nil {
		return domain.Permission{}, fmt.Errorf("create permission %s in area %s: %w", permission.ID, areaID, err)
	}
	return permission, nil
}

// DeletePermission keeps the permission, marked as deleted. A permission that
// a membership that has not ended holds cannot be deleted: its member would
// keep a permission that no longer exists.
func (s *AreaService) DeletePermission(ctx context.Context, administratorID, areaID, permissionID string) error {
	if err := s.requireAdministrator(ctx, administratorID, actionDeletePermission); err != nil {
		return err
	}
	area, err := s.activeArea(ctx, areaID)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(area.Permissions, func(p domain.Permission) bool { return p.ID == permissionID && !p.Deleted() })
	if i < 0 {
		return fmt.Errorf("permission %s of area %s: %w", permissionID, areaID, domain.ErrPermissionNotFound)
	}
	memberships, err := s.unendedMemberships(ctx, areaID)
	if err != nil {
		return err
	}
	if slices.ContainsFunc(memberships, func(m domain.Membership) bool { return slices.Contains(m.Permissions, permissionID) }) {
		return fmt.Errorf("%w: permission %s of area %s is held by memberships that have not ended", domain.ErrInvalidPermission, permissionID, areaID)
	}
	area.Permissions[i].DeletedAt = s.clock().UTC()
	area.Permissions[i].DeletedBy = administratorID
	if err := s.areas.Save(ctx, area); err != nil {
		return fmt.Errorf("delete permission %s of area %s: %w", permissionID, areaID, err)
	}
	return nil
}

func (s *AreaService) activeArea(ctx context.Context, areaID string) (domain.Area, error) {
	area, err := s.areas.Get(ctx, areaID)
	if err != nil {
		return domain.Area{}, fmt.Errorf("area %s: %w", areaID, err)
	}
	if area.Deleted() {
		return domain.Area{}, fmt.Errorf("area %s: %w", areaID, domain.ErrAreaNotFound)
	}
	return area, nil
}

func (s *AreaService) unendedMemberships(ctx context.Context, areaID string) ([]domain.Membership, error) {
	memberships, err := s.memberships.ListByArea(ctx, areaID)
	if err != nil {
		return nil, fmt.Errorf("memberships of area %s: %w", areaID, err)
	}
	now := s.clock().UTC()
	return slices.DeleteFunc(memberships, func(m domain.Membership) bool { return !m.Until.IsZero() && !m.Until.After(now) }), nil
}

func validatePermission(area domain.Area, p domain.Permission) error {
	if !domain.ValidID(p.ID) {
		return fmt.Errorf("id %q has to be lowercase letters, digits and underscores, and start with a letter", p.ID)
	}
	if p.Name == "" {
		return errors.New("a permission needs a name")
	}
	if slices.ContainsFunc(area.Permissions, func(other domain.Permission) bool { return other.ID == p.ID }) {
		return fmt.Errorf("id %q is already used in area %s", p.ID, area.ID)
	}
	if slices.ContainsFunc(area.ActivePermissions(), func(other domain.Permission) bool { return strings.EqualFold(other.Name, p.Name) }) {
		return fmt.Errorf("name %q is already used in area %s", p.Name, area.ID)
	}
	return nil
}
