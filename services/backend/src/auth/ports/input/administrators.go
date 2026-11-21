package input

import (
	"context"

	"backend/src/auth/domain"
)

// Administrators lets an administrator grant and end the others. SIGTA is
// never left without one: the last cannot be ended.
type Administrators interface {
	// List returns the administrators active now, in the order they started.
	List(ctx context.Context, administratorID string) ([]domain.Administrator, error)
	Grant(ctx context.Context, administratorID, userID string) (domain.Administrator, error)
	End(ctx context.Context, administratorID, id string) error
}
