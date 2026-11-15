package input

import (
	"context"
	"errors"

	"backend/src/procedure/domain/inbox"
)

// ErrNotStaff means the person is not part of the staff, so the inbox is not
// theirs.
var ErrNotStaff = errors.New("the inbox is only for the staff")

// ErrProcedureNotFound means the procedure does not exist or the person cannot
// see it; the two are not told apart.
var ErrProcedureNotFound = errors.New("procedure not found")

type Inbox interface {
	// List answers a query of the inbox, over what the person can see. It
	// fails with ErrNotStaff if the person is not part of the staff.
	List(ctx context.Context, person inbox.Person, q inbox.Query) (inbox.List, error)
	// MarkSeen records that the person saw a procedure now. It fails with
	// ErrNotStaff, or with ErrProcedureNotFound if the person cannot see it.
	MarkSeen(ctx context.Context, person inbox.Person, procedureID string) error
}
