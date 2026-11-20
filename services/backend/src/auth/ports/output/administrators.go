package output

import (
	"context"

	"backend/src/auth/domain"
)

type AdministratorRepository interface {
	IsAdministrator(ctx context.Context, userID string) (bool, error)
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
