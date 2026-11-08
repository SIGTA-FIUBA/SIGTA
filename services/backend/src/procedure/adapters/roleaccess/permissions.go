// Package roleaccess derives what a person can see from their role alone. It
// stands in for per-type permissions until they exist: the staff sees every
// procedure type with their role as the profile, and an applicant sees only
// their own procedures. Replacing it with the real permissions changes only
// the composition root.
package roleaccess

import (
	"context"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/output"
)

var _ output.Permissions = Permissions{}

type Permissions struct{}

func (Permissions) AccessOf(_ context.Context, person inbox.Person) (inbox.Access, error) {
	switch profile := definition.Profile(person.Role); profile {
	case definition.ProfileStaffOperative, definition.ProfileStaffApprover:
		return inbox.Access{EveryType: profile}, nil
	default:
		// Anyone else is treated as an applicant, who sees only their own.
		return inbox.Access{Applicant: person.ID}, nil
	}
}
