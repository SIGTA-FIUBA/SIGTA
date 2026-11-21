package output

import (
	"context"

	"backend/src/auth/domain"
)

// MembershipRepository keeps every membership, ended or not: they are never
// deleted, so that their history stays. The lists keep the order in which
// they were granted, and Get fails with domain.ErrMembershipNotFound if there
// is no such membership. Save grants a new membership or ends one: of an
// existing membership only From, Until and EndedBy change. Replace saves
// ended, an existing membership, and granted, a new one, at once, so that
// changing what a membership gives never leaves the person with both or
// with none. Both check, as they save, what a change made meanwhile could
// break: a new membership fails with domain.ErrMembershipOverlaps if it
// overlaps another of the person, and ending one fails with
// domain.ErrMembershipEnded if it had already ended by its new Until. A
// repository that also keeps the areas, as the database does, refuses with
// domain.ErrInvalidMembership a grant in an area, role or permission removed
// meanwhile, as AreaRepository.Save refuses the removal of one held by a
// membership that has not ended; one in memory, which does not see the
// areas, leaves that to the use cases.
type MembershipRepository interface {
	ListByUser(ctx context.Context, userID string) ([]domain.Membership, error)
	ListByArea(ctx context.Context, areaID string) ([]domain.Membership, error)
	Get(ctx context.Context, id string) (domain.Membership, error)
	Save(ctx context.Context, m domain.Membership) error
	Replace(ctx context.Context, ended, granted domain.Membership) error
}
