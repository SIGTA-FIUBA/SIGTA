package output

import (
	"context"

	"backend/src/auth/domain"
)

// AreaRepository keeps every area, deleted or not. List returns only those
// not deleted, and Get fails with domain.ErrAreaNotFound if there is no such
// area.
type AreaRepository interface {
	List(ctx context.Context) ([]domain.Area, error)
	Get(ctx context.Context, id string) (domain.Area, error)
	Save(ctx context.Context, a domain.Area) error
}

type Administrators interface {
	IsAdministrator(ctx context.Context, userID string) (bool, error)
}

// MembershipRepository keeps every membership, ended or not: they are never
// deleted, so that their history stays. Get fails with
// domain.ErrMembershipNotFound if there is no such membership.
type MembershipRepository interface {
	ListByUser(ctx context.Context, userID string) ([]domain.Membership, error)
	ListByArea(ctx context.Context, areaID string) ([]domain.Membership, error)
	Get(ctx context.Context, id string) (domain.Membership, error)
	Save(ctx context.Context, m domain.Membership) error
}
