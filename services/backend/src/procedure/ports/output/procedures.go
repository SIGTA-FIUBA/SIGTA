package output

import (
	"context"

	"backend/src/procedure/domain"
)

type ProcedureRepository interface {
	Add(ctx context.Context, p domain.Procedure) error
	Get(ctx context.Context, id string) (domain.Procedure, error)
	List(ctx context.Context) ([]domain.Procedure, error)
}

// Access keeps who administers SIGTA and every membership, ended or not:
// memberships are never deleted, so that their history stays.
type Access interface {
	Administrator(ctx context.Context, userID string) (bool, error)
	Memberships(ctx context.Context, userID string) ([]domain.Membership, error)
	Membership(ctx context.Context, id string) (domain.Membership, error)
	SaveMembership(ctx context.Context, m domain.Membership) error
}

type DeniedAttempts interface {
	Record(ctx context.Context, a domain.DeniedAttempt) error
}
