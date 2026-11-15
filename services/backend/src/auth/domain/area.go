package domain

import (
	"errors"
	"regexp"
	"slices"
	"time"
)

var (
	ErrInvalidArea        = errors.New("invalid area")
	ErrAreaNotFound       = errors.New("area not found")
	ErrInvalidPermission  = errors.New("invalid permission")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrInvalidRole        = errors.New("invalid role")
	ErrRoleNotFound       = errors.New("role not found")
)

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func ValidID(id string) bool {
	return idPattern.MatchString(id)
}

// Area exists only within staff (ADR-020). A deleted area is kept, so that
// memberships and published versions still name it. A system area is one
// SIGTA brings, like the Secretariat: it cannot be deleted.
type Area struct {
	ID          string
	Name        string
	System      bool
	Permissions []Permission
	Roles       []Role
	DeletedAt   time.Time
	DeletedBy   string
}

func (a Area) Deleted() bool {
	return !a.DeletedAt.IsZero()
}

// Permission belongs to an area, and staff steps of that area ask for it. Its
// name means nothing to SIGTA and no permission covers another (ADR-020). The
// exception are system permissions, which enable a function of SIGTA instead
// of a step, and cannot be deleted. A deleted permission is kept, so that its
// ID is never reused.
type Permission struct {
	ID        string
	Name      string
	System    bool
	DeletedAt time.Time
	DeletedBy string
}

func (p Permission) Deleted() bool {
	return !p.DeletedAt.IsZero()
}

func (a Area) ActivePermissions() []Permission {
	return slices.DeleteFunc(slices.Clone(a.Permissions), Permission.Deleted)
}

func (a Area) HasActivePermission(id string) bool {
	return slices.ContainsFunc(a.Permissions, func(p Permission) bool { return p.ID == id && !p.Deleted() })
}

// Role groups permissions of its area, so that they are not granted one by
// one. A membership refers to the role, so changing the role changes what
// every member with it may do. Each change is a new revision: the last one
// holds the current permissions, and the earlier ones tell what they were.
type Role struct {
	ID        string
	Name      string
	Revisions []RoleRevision
	DeletedAt time.Time
	DeletedBy string
}

type RoleRevision struct {
	Permissions []string
	At          time.Time
	By          string
}

func (r Role) Deleted() bool {
	return !r.DeletedAt.IsZero()
}

func (r Role) Permissions() []string {
	if len(r.Revisions) == 0 {
		return nil
	}
	return slices.Clone(r.Revisions[len(r.Revisions)-1].Permissions)
}

func (a Area) ActiveRoles() []Role {
	return slices.DeleteFunc(slices.Clone(a.Roles), Role.Deleted)
}

func (a Area) HasActiveRole(id string) bool {
	return slices.ContainsFunc(a.Roles, func(r Role) bool { return r.ID == id && !r.Deleted() })
}

// GrantedBy returns the permissions that a membership with these roles and
// loose permissions holds in the area. Deleted roles and permissions give
// nothing.
func (a Area) GrantedBy(roles, permissions []string) []string {
	var granted []string
	add := func(id string) {
		if a.HasActivePermission(id) && !slices.Contains(granted, id) {
			granted = append(granted, id)
		}
	}
	for _, id := range permissions {
		add(id)
	}
	for _, r := range a.Roles {
		if !r.Deleted() && slices.Contains(roles, r.ID) {
			for _, id := range r.Permissions() {
				add(id)
			}
		}
	}
	return granted
}
