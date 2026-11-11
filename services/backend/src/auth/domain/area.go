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
)

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func ValidID(id string) bool {
	return idPattern.MatchString(id)
}

// Area exists only within staff: the applicant has none (ADR-020). A deleted
// area is kept, so that memberships and published versions still name it.
type Area struct {
	ID          string
	Name        string
	Permissions []Permission
	DeletedAt   time.Time
	DeletedBy   string
}

func (a Area) Deleted() bool {
	return !a.DeletedAt.IsZero()
}

// Permission belongs to an area, and staff steps of that area ask for it. Its
// name means nothing to SIGTA: no permission covers another (ADR-020). A
// deleted permission is kept, so that its ID is never reused.
type Permission struct {
	ID        string
	Name      string
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
