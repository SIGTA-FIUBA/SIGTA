package output

import (
	"context"

	"backend/src/auth/domain"
)

// MembershipRepository keeps every membership, ended or not: they are never
// deleted, so that their history stays. The lists keep the order in which
// they were granted, and Get fails with domain.ErrMembershipNotFound if there
// is no such membership. Save grants a new membership or ends one: of an
// existing membership only From, Until and EndedBy change.
type MembershipRepository interface {
	ListByUser(ctx context.Context, userID string) ([]domain.Membership, error)
	ListByArea(ctx context.Context, areaID string) ([]domain.Membership, error)
	Get(ctx context.Context, id string) (domain.Membership, error)
	Save(ctx context.Context, m domain.Membership) error
}
