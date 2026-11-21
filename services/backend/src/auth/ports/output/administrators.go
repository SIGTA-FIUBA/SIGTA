package output

import (
	"context"
	"time"

	"backend/src/auth/domain"
)

// AdministratorRepository keeps the periods of the administrators, ended or
// not: they are never deleted, so that the history keeps who granted and who
// ended each. IsAdministrator tells whether the user has a period active now.
type AdministratorRepository interface {
	IsAdministrator(ctx context.Context, userID string) (bool, error)
}

// AdministratorPeriodRepository is what administrators need to grant and end
// the others in SIGTA.
type AdministratorPeriodRepository interface {
	AdministratorRepository
	// ListActive returns the periods active at at, in the order they started.
	ListActive(ctx context.Context, at time.Time) ([]domain.Administrator, error)
	// Grant adds a period granted by an administrator. It fails with
	// domain.ErrAlreadyAdministrator if the person has a period without an
	// end.
	Grant(ctx context.Context, a domain.Administrator) error
	// End ends at at the period with id, recording by as who ended it. It
	// fails with domain.ErrAdministratorNotFound if there is no such period
	// active at at, and with domain.ErrLastAdministrator if no other is
	// active then: SIGTA is never left without administrators.
	End(ctx context.Context, id string, at time.Time, by string) error
}

// FirstAdministratorRepository is what the grant-admin command needs.
type FirstAdministratorRepository interface {
	AdministratorRepository
	// GrantFirst adds a as the initial grant of the system, granted by
	// nobody, if there is no administrator, and reports false, adding
	// nothing, if there is one: from then on administrators grant the others
	// in SIGTA. The server never calls it: only the grant-admin command,
	// which connects as the role that owns the tables. Once every
	// administrator has ended, it can grant one again.
	GrantFirst(ctx context.Context, a domain.Administrator) (bool, error)
}
