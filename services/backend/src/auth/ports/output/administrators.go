package output

import (
	"context"

	"backend/src/auth/domain"
)

type AdministratorRepository interface {
	IsAdministrator(ctx context.Context, userID string) (bool, error)
	// GrantFirst adds a as the initial grant of the system, granted by
	// nobody, if there is no administrator, and reports false, adding
	// nothing, if there is one: from then on administrators grant the others
	// in SIGTA. The server never calls it: only the grant-admin command,
	// which connects as the role that owns the tables.
	GrantFirst(ctx context.Context, a domain.Administrator) (bool, error)
}
