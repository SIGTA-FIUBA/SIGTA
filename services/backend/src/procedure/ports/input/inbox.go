package input

import (
	"context"
	"errors"

	"backend/src/procedure/domain/inbox"
)

// ErrNotStaff means the person is not part of the staff, so the inbox is not
// theirs.
var ErrNotStaff = errors.New("the inbox is only for the staff")

type Inbox interface {
	// List answers a query of the inbox, over what the person can see. It
	// fails with ErrNotStaff if the person is not part of the staff.
	List(ctx context.Context, person inbox.Person, q inbox.Query) (inbox.List, error)
}
