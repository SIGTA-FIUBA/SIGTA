package input

import (
	"context"
	"errors"

	"backend/src/procedure/domain/inbox"
)

// ErrNotStaff means the person has no membership in force, so the inbox is
// not theirs (ADR-020).
var ErrNotStaff = errors.New("the inbox is only for the staff")

type Inbox interface {
	// List answers a query of the inbox over the procedures in progress that
	// the person sees as staff. It fails with ErrNotStaff if the person is not
	// part of the staff.
	List(ctx context.Context, userID string, q inbox.Query) (inbox.List, error)
}
