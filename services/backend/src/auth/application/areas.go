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
	actionListAreas        = "list areas"
	actionCreateArea       = "create area"
	actionDeleteArea       = "delete area"
	actionCreatePermission = "create permission"
	actionDeletePermission = "delete permission"
	actionCreateRole       = "create role"
	actionUpdateRole       = "update role"
	actionDeleteRole       = "delete role"
)

var platformPermissions = []domain.Permission{
	{ID: input.EditProcedureTypes, Name: "Editar tipos de trámite", System: true},
	{ID: input.PublishProcedureTypes, Name: "Publicar tipos de trámite", System: true},
}

type AreaService struct {
	administratorGuard
	areas       output.AreaRepository
	memberships output.MembershipRepository
	now         func() time.Time
}

func NewAreaService(areas output.AreaRepository, administrators output.AdministratorRepository, memberships output.MembershipRepository, now func() time.Time) *AreaService {
	return &AreaService{
		administratorGuard: administratorGuard{administrators: administrators},
		areas:              areas,
		memberships:        memberships,
		now:                now,
	}
}

// ErrSecretariatIncomplete means the Secretariat or one of its platform
// permissions, which migration 0007 adds, is missing, deleted or no longer a
// system one. The error that wraps it says which.
var ErrSecretariatIncomplete = errors.New("the Secretariat is not as the migrations leave it")

// EnsureSecretariat creates the Secretariat with its platform permissions, or
// adds the ones it lacks, in stores that are not migrated: those in memory.
// It changes nothing when they are already there.
func EnsureSecretariat(ctx context.Context, areas output.AreaRepository) error {
	area, err := areas.Get(ctx, input.SecretariatArea)
	switch {
	case errors.Is(err, domain.ErrAreaNotFound):
		area = domain.Area{ID: input.SecretariatArea, Name: "Secretaría de Gestión Académica"}
	case err != nil:
		return fmt.Errorf("area %s: %w", input.SecretariatArea, err)
	}
	if !completeSecretariat(&area) {
		return nil
	}
	if err := areas.Save(ctx, area); err != nil {
		return fmt.Errorf("save area %s: %w", input.SecretariatArea, err)
	}
	return nil
}

// CheckSecretariat refuses a database whose Secretariat lacks what the code
// needs of it, as when it was changed by hand, and says what. The server does
// not fix it: what the code needs comes with the migrations. It does not look
// at the names, nor at the other permissions of the area.
func CheckSecretariat(ctx context.Context, areas output.AreaRepository) error {
	area, err := areas.Get(ctx, input.SecretariatArea)
	if errors.Is(err, domain.ErrAreaNotFound) {
		return fmt.Errorf("%w: area %s is missing", ErrSecretariatIncomplete, input.SecretariatArea)
	}
	if err != nil {
		return fmt.Errorf("area %s: %w", input.SecretariatArea, err)
	}
	if problems := secretariatProblems(area); len(problems) > 0 {
		return fmt.Errorf("%w: %s", ErrSecretariatIncomplete, strings.Join(problems, ", "))
	}
	return nil
}

// secretariatProblems lists what the area lacks to be the Secretariat that the
// code needs.
func secretariatProblems(area domain.Area) []string {
	var problems []string
	if area.Deleted() {
		problems = append(problems, "area "+area.ID+" is deleted")
	}
	if !area.System {
		problems = append(problems, "area "+area.ID+" is not a system one")
	}
	for _, permission := range platformPermissions {
		i := slices.IndexFunc(area.Permissions, func(p domain.Permission) bool { return p.ID == permission.ID })
		switch {
		case i < 0:
			problems = append(problems, "permission "+permission.ID+" is missing")
		case area.Permissions[i].Deleted():
			problems = append(problems, "permission "+permission.ID+" is deleted")
		case !area.Permissions[i].System:
			problems = append(problems, "permission "+permission.ID+" is not a system one")
		}
	}
	return problems
}

// completeSecretariat marks the area as the Secretariat and adds or restores
// its platform permissions. It reports whether it changed anything.
func completeSecretariat(area *domain.Area) bool {
	changed := !area.System
	area.System = true
	for _, permission := range platformPermissions {
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
	return changed
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

// ListToAdminister is List for an administrator, who sees which areas and
// permissions come with SIGTA.
func (s *AreaService) ListToAdminister(ctx context.Context, administratorID string) ([]domain.Area, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionListAreas); err != nil {
		return nil, err
	}
	return s.List(ctx)
}

// Create refuses the ID of a deleted area too: memberships and published
// versions may still name it, and they would end up naming another area. It
// also refuses the ID of the Secretariat, even before it exists.
func (s *AreaService) Create(ctx context.Context, administratorID string, a domain.Area) (domain.Area, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionCreateArea); err != nil {
		return domain.Area{}, err
	}
	area := domain.Area{ID: a.ID, Name: strings.TrimSpace(a.Name), CreatedAt: s.now().UTC(), CreatedBy: administratorID}
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
		return fmt.Errorf("%w: %w: %s", domain.ErrInvalidArea, domain.ErrSystemArea, areaID)
	}
	memberships, err := s.unendedMemberships(ctx, areaID)
	if err != nil {
		return err
	}
	if len(memberships) > 0 {
		return fmt.Errorf("%w: %w: %s", domain.ErrInvalidArea, domain.ErrAreaHasMemberships, areaID)
	}
	area.DeletedAt = s.now().UTC()
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
	permission := domain.Permission{ID: p.ID, Name: strings.TrimSpace(p.Name), CreatedAt: s.now().UTC(), CreatedBy: administratorID}
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
// exists. A platform permission cannot be deleted at all.
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
		return fmt.Errorf("%w: %w: %s of area %s", domain.ErrInvalidPermission, domain.ErrSystemPermission, permissionID, areaID)
	}
	if slices.ContainsFunc(area.ActiveRoles(), func(r domain.Role) bool { return slices.Contains(r.Permissions(), permissionID) }) {
		return fmt.Errorf("%w: %w: %s of area %s", domain.ErrInvalidPermission, domain.ErrPermissionInRole, permissionID, areaID)
	}
	memberships, err := s.unendedMemberships(ctx, areaID)
	if err != nil {
		return err
	}
	if slices.ContainsFunc(memberships, func(m domain.Membership) bool { return slices.Contains(m.Permissions, permissionID) }) {
		return fmt.Errorf("%w: %w: %s of area %s", domain.ErrInvalidPermission, domain.ErrPermissionHeld, permissionID, areaID)
	}
	area.Permissions[i].DeletedAt = s.now().UTC()
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
		Revisions: []domain.RoleRevision{{Permissions: slices.Clone(r.Permissions), At: s.now().UTC(), By: administratorID}},
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
		At:          s.now().UTC(),
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
		return fmt.Errorf("%w: %w: %s of area %s", domain.ErrInvalidRole, domain.ErrRoleHeld, roleID, areaID)
	}
	area.Roles[i].DeletedAt = s.now().UTC()
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
	now := s.now().UTC()
	return slices.DeleteFunc(memberships, func(m domain.Membership) bool { return !m.Until.IsZero() && !m.Until.After(now) }), nil
}

func validatePermission(area domain.Area, p domain.Permission) error {
	if !domain.ValidID(p.ID) {
		return fmt.Errorf(invalidID, p.ID)
	}
	// In every area, so that no other permission goes by the name of one.
	if slices.ContainsFunc(platformPermissions, func(platform domain.Permission) bool { return platform.ID == p.ID }) {
		return fmt.Errorf("id %q is reserved for a platform permission", p.ID)
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
