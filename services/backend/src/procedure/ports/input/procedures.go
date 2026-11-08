package input

import (
	"context"
	"time"

	"backend/src/procedure/domain"
)

type ProcedureView struct {
	Procedure domain.Procedure
	Actions   []string
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
}

type Procedures interface {
	Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error)
	Open(ctx context.Context, userID, procedureID string) (ProcedureView, error)
	List(ctx context.Context, userID string) ([]domain.Procedure, error)
	Act(ctx context.Context, userID, procedureID, stepID string) error
}
