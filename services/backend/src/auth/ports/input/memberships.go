package input

import (
	"context"
	"time"

	"backend/src/auth/domain"
)

type Areas interface {
	List(ctx context.Context) ([]domain.Area, error)
}

// MembershipGrant is what an administrator asks for. A zero From means now,
// and a zero Until means until it is ended.
type MembershipGrant struct {
	UserID  string
	Area    string
	Profile domain.Profile
	Scope   []int64
	From    time.Time
	Until   time.Time
}

type Memberships interface {
	Grant(ctx context.Context, administratorID string, g MembershipGrant) (domain.Membership, error)
	End(ctx context.Context, administratorID, membershipID string) error
	Permissions(ctx context.Context, userID string) (domain.Permissions, error)
}
