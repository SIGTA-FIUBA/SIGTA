package input

import (
	"context"
	"time"

	"backend/src/auth/domain"
)

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
	// Change ends the membership and grants another with these roles and
	// permissions, and returns the new one.
	Change(ctx context.Context, administratorID, membershipID string, roles, permissions []string) (domain.Membership, error)
	// ListByArea returns the memberships of the area that have not ended.
	ListByArea(ctx context.Context, administratorID, areaID string) ([]domain.Membership, error)
	Access(ctx context.Context, userID string) (domain.Access, error)
}
