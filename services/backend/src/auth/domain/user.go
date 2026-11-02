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

func (r Role) Known() bool {
	switch r {
	case RoleApplicant, RoleStaffOperative, RoleStaffApprover:
		return true
	}
	return false
}

// User is a person registered in SIGTA. Their role lives in SIGTA's own
// storage, never in the identity provider.
type User struct {
	ID    string
	Email string
	Name  string
	Role  Role
}

// CanUseLoginLink tells whether the user may log in through an emailed link.
// Staff log in with their institutional account only.
func (u User) CanUseLoginLink() bool {
	return u.Role == RoleApplicant
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// InDomain tells whether a normalized email belongs to the given domain, such
// as "fi.uba.ar".
func InDomain(email, emailDomain string) bool {
	return strings.HasSuffix(email, "@"+emailDomain)
}
