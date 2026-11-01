package domain

import (
	"errors"
	"strings"
)

var ErrUnknownUser = errors.New("no user registered with that identity")

type Role string

const (
	RoleApplicant      Role = "applicant"
	RoleStaffOperative Role = "staff_operative"
	RoleStaffApprover  Role = "staff_approver"
)

// User is a person registered in SIGTA. Their profile lives in SIGTA's own
// storage, never in the identity provider.
type User struct {
	ID    string
	Email string
	Name  string
	Role  Role
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// IsStaff tells whether the user belongs to the Secretaría, whatever their
// profile.
func (u User) IsStaff() bool {
	return u.Role == RoleStaffOperative || u.Role == RoleStaffApprover
}
