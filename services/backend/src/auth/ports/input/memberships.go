package input

import (
	"context"
	"time"

	"backend/src/auth/domain"
)

// The Secretariat is the one area that always exists: SIGTA brings it, with
// the two system permissions that let someone edit and publish procedure
// types. Nobody else can do either, and no step can ask for them (ADR-020).
const (
	SecretariatArea       = "secretariat"
	EditProcedureTypes    = "edit_procedure_types"
	PublishProcedureTypes = "publish_procedure_types"
)

// SystemPermission tells whether a permission enables a function of SIGTA
// instead of a step.
func SystemPermission(area, permission string) bool {
	return area == SecretariatArea && (permission == EditProcedureTypes || permission == PublishProcedureTypes)
}

// List returns the areas that are not deleted, each with the permissions and
// roles that are not deleted.
type Areas interface {
	List(ctx context.Context) ([]domain.Area, error)
	Create(ctx context.Context, administratorID string, a domain.Area) (domain.Area, error)
	Delete(ctx context.Context, administratorID, areaID string) error
	CreatePermission(ctx context.Context, administratorID, areaID string, p domain.Permission) (domain.Permission, error)
	DeletePermission(ctx context.Context, administratorID, areaID, permissionID string) error
	CreateRole(ctx context.Context, administratorID, areaID string, r NewRole) (domain.Role, error)
	UpdateRolePermissions(ctx context.Context, administratorID, areaID, roleID string, permissions []string) (domain.Role, error)
	DeleteRole(ctx context.Context, administratorID, areaID, roleID string) error
}

type NewRole struct {
	ID          string
	Name        string
	Permissions []string
}

// MembershipGrant is what an administrator asks for. It needs at least one
// role or permission of the area. A zero From means now, and a zero Until
// means until it is ended.
type MembershipGrant struct {
	UserID      string
	Area        string
	Roles       []string
	Permissions []string
	Scope       []int64
	From        time.Time
	Until       time.Time
}

type Memberships interface {
	Grant(ctx context.Context, administratorID string, g MembershipGrant) (domain.Membership, error)
	End(ctx context.Context, administratorID, membershipID string) error
	Access(ctx context.Context, userID string) (domain.Access, error)
}

// An empty Scope covers every procedure type.
type AreaReach struct {
	Area  string
	Scope []int64
}

// Other modules ask StaffAuthorizer instead of reading memberships (ADR-021).
type StaffAuthorizer interface {
	Holds(ctx context.Context, userID, area string, procedureType int64, permissions []string) (bool, error)
	Reach(ctx context.Context, userID string) ([]AreaReach, error)
}
