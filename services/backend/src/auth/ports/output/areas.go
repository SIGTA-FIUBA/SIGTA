package output

import (
	"context"

	"backend/src/auth/domain"
)

// AreaRepository keeps every area, deleted or not. List returns only those
// not deleted, in the order they were added, and Get fails with
// domain.ErrAreaNotFound if there is no such area.
// Save takes the whole area as Get returned it, with its changes: nothing is
// removed from it, permissions and roles are only added or changed, and the
// revisions of a role are only added at the end. An area read before another
// Save does not undo it: what that Save added or deleted stays, and the
// revisions of a role that are not stored yet go after the stored ones.
type AreaRepository interface {
	List(ctx context.Context) ([]domain.Area, error)
	Get(ctx context.Context, id string) (domain.Area, error)
	Save(ctx context.Context, a domain.Area) error
}
