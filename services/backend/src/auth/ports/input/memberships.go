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

// AreaReach is an area where a person is staff today, and the procedure types
// their membership covers there. An empty Scope covers every type.
type AreaReach struct {
	Area  string
	Scope []int64
}

// StaffAuthorizer answers what a person can do as staff (ADR-017). Other
// modules ask it instead of reading memberships (ADR-018).
type StaffAuthorizer interface {
	// IsStaff tells whether the person has an active membership. Staff cannot
	// start procedures.
	IsStaff(ctx context.Context, userID string) (bool, error)
	// Holds tells whether the person has at least the profile in the area,
	// for the procedure type, today.
	Holds(ctx context.Context, userID, area string, procedureType int64, profile Profile) (bool, error)
	// Reach lists the areas where the person is staff today.
	Reach(ctx context.Context, userID string) ([]AreaReach, error)
}
