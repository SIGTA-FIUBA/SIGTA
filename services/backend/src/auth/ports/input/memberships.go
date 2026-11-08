package input

import (
	"context"
	"time"

	"backend/src/auth/domain"
)

// Profile is the rank of a membership, and what a step of a procedure type
// asks for.
type Profile = domain.Profile

const (
	ProfileStaffOperative = domain.ProfileStaffOperative
	ProfileStaffApprover  = domain.ProfileStaffApprover
)

type Areas interface {
	List(ctx context.Context) ([]domain.Area, error)
	Create(ctx context.Context, administratorID string, a domain.Area) (domain.Area, error)
	Delete(ctx context.Context, administratorID, areaID string) error
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

// An empty Scope covers every procedure type.
type AreaReach struct {
	Area  string
	Scope []int64
}

// Other modules ask StaffAuthorizer instead of reading memberships (ADR-018).
type StaffAuthorizer interface {
	IsStaff(ctx context.Context, userID string) (bool, error)
	Holds(ctx context.Context, userID, area string, procedureType int64, profile Profile) (bool, error)
	Reach(ctx context.Context, userID string) ([]AreaReach, error)
}
