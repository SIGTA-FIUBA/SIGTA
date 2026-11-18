package input

import (
	"context"

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
