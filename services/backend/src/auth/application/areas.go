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
	actionCreateRole       = "create role"
	actionUpdateRole       = "update role"
	actionDeleteRole       = "delete role"
)

var systemPermissions = []domain.Permission{
	{ID: input.EditProcedureTypes, Name: "Editar tipos de trámite", System: true},
	{ID: input.PublishProcedureTypes, Name: "Publicar tipos de trámite", System: true},
}

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

// EnsureSecretariat creates the Secretariat with its system permissions, or
// adds the ones it lacks. It runs on every start, before anything reads the
// areas, and changes nothing when they are already there.
func EnsureSecretariat(ctx context.Context, areas output.AreaRepository) error {
	area, err := areas.Get(ctx, input.SecretariatArea)
	switch {
	case errors.Is(err, domain.ErrAreaNotFound):
		area = domain.Area{ID: input.SecretariatArea, Name: "Secretaría de Gestión Académica"}
	case err != nil:
		return fmt.Errorf("area %s: %w", input.SecretariatArea, err)
	}
	changed := !area.System
	area.System = true
	for _, permission := range systemPermissions {
		i := slices.IndexFunc(area.Permissions, func(p domain.Permission) bool { return p.ID == permission.ID })
		switch {
		case i < 0:
			area.Permissions = append(area.Permissions, permission)
			changed = true
		case !area.Permissions[i].System || area.Permissions[i].Deleted():
			area.Permissions[i] = permission
			changed = true
		}
	}
	if !changed {
		return nil
	}
	if err := areas.Save(ctx, area); err != nil {
		return fmt.Errorf("save area %s: %w", input.SecretariatArea, err)
	}
	return nil
}

func (s *AreaService) List(ctx context.Context) ([]domain.Area, error) {
	areas, err := s.areas.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list areas: %w", err)
	}
	for i := range areas {
		areas[i].Permissions = areas[i].ActivePermissions()
		areas[i].Roles = areas[i].ActiveRoles()
	}
	return areas, nil
}

// Create refuses the ID of a deleted area too: memberships and published
// versions may still name it, and they would end up naming another area. It
// also refuses the ID of the Secretariat, even before it exists.
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
// would keep permissions in an area that no longer exists. A system area
// cannot be deleted at all.
func (s *AreaService) Delete(ctx context.Context, administratorID, areaID string) error {
	if err := s.requireAdministrator(ctx, administratorID, actionDeleteArea); err != nil {
		return err
	}
	area, err := s.activeArea(ctx, areaID)
	if err != nil {
		return err
	}
	if area.System {
		return fmt.Errorf("%w: area %s comes with SIGTA", domain.ErrInvalidArea, areaID)
	}
	memberships, err := s.unendedMemberships(ctx, areaID)
	if err != nil {
		return err
	}
	if len(memberships) > 0 {
		return fmt.Errorf("%w: area %s has memberships that have not ended", domain.ErrInvalidArea, areaID)
	}
	area.DeletedAt = s.clock().UTC()
	area.DeletedBy = administratorID
	if err := s.areas.Save(ctx, area); err != nil {
		return fmt.Errorf("delete area %s: %w", areaID, err)
	}
	return nil
}

const invalidID = "id %q has to be lowercase letters, digits and underscores, and start with a letter"

func (s *AreaService) validate(ctx context.Context, a domain.Area) error {
	if !domain.ValidID(a.ID) {
		return fmt.Errorf(invalidID, a.ID)
	}
	if a.ID == input.SecretariatArea {
		return fmt.Errorf("id %q is reserved for the Secretariat", a.ID)
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

// DeletePermission keeps the permission, marked as deleted. It refuses a
// permission that a membership that has not ended holds, or that a role that
// is not deleted includes: someone would keep a permission that no longer
// exists. A system permission cannot be deleted at all.
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
	if area.Permissions[i].System {
		return fmt.Errorf("%w: permission %s of area %s comes with SIGTA", domain.ErrInvalidPermission, permissionID, areaID)
	}
	if slices.ContainsFunc(area.ActiveRoles(), func(r domain.Role) bool { return slices.Contains(r.Permissions(), permissionID) }) {
		return fmt.Errorf("%w: permission %s of area %s is part of a role", domain.ErrInvalidPermission, permissionID, areaID)
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

// CreateRole refuses the ID of a deleted role of the area too, so that the
// history of a membership never names two different roles.
func (s *AreaService) CreateRole(ctx context.Context, administratorID, areaID string, r input.NewRole) (domain.Role, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionCreateRole); err != nil {
		return domain.Role{}, err
	}
	area, err := s.activeArea(ctx, areaID)
	if err != nil {
		return domain.Role{}, err
	}
	role := domain.Role{
		ID:        r.ID,
		Name:      strings.TrimSpace(r.Name),
		Revisions: []domain.RoleRevision{{Permissions: slices.Clone(r.Permissions), At: s.clock().UTC(), By: administratorID}},
	}
	if err := validateRole(area, role); err != nil {
		return domain.Role{}, fmt.Errorf("%w: %w", domain.ErrInvalidRole, err)
	}
	area.Roles = append(area.Roles, role)
	if err := s.areas.Save(ctx, area); err != nil {
		return domain.Role{}, fmt.Errorf("create role %s in area %s: %w", role.ID, areaID, err)
	}
	return role, nil
}

// UpdateRolePermissions adds a revision to the role. Every membership with the
// role holds the new permissions from then on.
func (s *AreaService) UpdateRolePermissions(ctx context.Context, administratorID, areaID, roleID string, permissions []string) (domain.Role, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionUpdateRole); err != nil {
		return domain.Role{}, err
	}
	area, i, err := s.activeRole(ctx, areaID, roleID)
	if err != nil {
		return domain.Role{}, err
	}
	if err := validateRolePermissions(area, permissions); err != nil {
		return domain.Role{}, fmt.Errorf("%w: %w", domain.ErrInvalidRole, err)
	}
	area.Roles[i].Revisions = append(area.Roles[i].Revisions, domain.RoleRevision{
		Permissions: slices.Clone(permissions),
		At:          s.clock().UTC(),
		By:          administratorID,
	})
	if err := s.areas.Save(ctx, area); err != nil {
		return domain.Role{}, fmt.Errorf("update role %s of area %s: %w", roleID, areaID, err)
	}
	return area.Roles[i], nil
}

// DeleteRole keeps the role, marked as deleted. A role that a membership that
// has not ended holds cannot be deleted.
func (s *AreaService) DeleteRole(ctx context.Context, administratorID, areaID, roleID string) error {
	if err := s.requireAdministrator(ctx, administratorID, actionDeleteRole); err != nil {
		return err
	}
	area, i, err := s.activeRole(ctx, areaID, roleID)
	if err != nil {
		return err
	}
	memberships, err := s.unendedMemberships(ctx, areaID)
	if err != nil {
		return err
	}
	if slices.ContainsFunc(memberships, func(m domain.Membership) bool { return slices.Contains(m.Roles, roleID) }) {
		return fmt.Errorf("%w: role %s of area %s is held by memberships that have not ended", domain.ErrInvalidRole, roleID, areaID)
	}
	area.Roles[i].DeletedAt = s.clock().UTC()
	area.Roles[i].DeletedBy = administratorID
	if err := s.areas.Save(ctx, area); err != nil {
		return fmt.Errorf("delete role %s of area %s: %w", roleID, areaID, err)
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

func (s *AreaService) activeRole(ctx context.Context, areaID, roleID string) (domain.Area, int, error) {
	area, err := s.activeArea(ctx, areaID)
	if err != nil {
		return domain.Area{}, 0, err
	}
	i := slices.IndexFunc(area.Roles, func(r domain.Role) bool { return r.ID == roleID && !r.Deleted() })
	if i < 0 {
		return domain.Area{}, 0, fmt.Errorf("role %s of area %s: %w", roleID, areaID, domain.ErrRoleNotFound)
	}
	return area, i, nil
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
		return fmt.Errorf(invalidID, p.ID)
	}
	if input.SystemPermission(area.ID, p.ID) {
		return fmt.Errorf("id %q is reserved for a system permission", p.ID)
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

func validateRole(area domain.Area, r domain.Role) error {
	if !domain.ValidID(r.ID) {
		return fmt.Errorf(invalidID, r.ID)
	}
	if r.Name == "" {
		return errors.New("a role needs a name")
	}
	if slices.ContainsFunc(area.Roles, func(other domain.Role) bool { return other.ID == r.ID }) {
		return fmt.Errorf("id %q is already used in area %s", r.ID, area.ID)
	}
	if slices.ContainsFunc(area.ActiveRoles(), func(other domain.Role) bool { return strings.EqualFold(other.Name, r.Name) }) {
		return fmt.Errorf("name %q is already used in area %s", r.Name, area.ID)
	}
	return validateRolePermissions(area, r.Permissions())
}

func validateRolePermissions(area domain.Area, permissions []string) error {
	if len(permissions) == 0 {
		return errors.New("a role needs at least one permission")
	}
	return checkListed("permission", permissions, area.ID, area.HasActivePermission)
}
