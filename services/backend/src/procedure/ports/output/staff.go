package output

import (
	"context"

	"backend/src/procedure/domain"
)

// StaffDirectory tells who someone is when they act: their active memberships,
// which auth keeps. Someone unknown is an applicant with no memberships.
type StaffDirectory interface {
	Actor(ctx context.Context, userID string) (domain.Actor, error)
}
