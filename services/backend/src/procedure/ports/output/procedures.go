package output

import (
	"context"
	"errors"

	"backend/src/procedure/domain"
)

var ErrUnknownUser = errors.New("user not registered for access")

type Procedures interface {
	Add(ctx context.Context, p domain.Procedure) error
	// Get fails with domain.ErrProcedureNotFound if there is no such
	// procedure.
	Get(ctx context.Context, id string) (domain.Procedure, error)
	// List returns every procedure, ordered by identifier. Filtering what
	// someone can see is up to the caller.
	List(ctx context.Context) ([]domain.Procedure, error)
}

// Access stores what kind of user each person is and the permissions of the
// staff.
type Access interface {
	Register(ctx context.Context, userID string, kind domain.UserKind) error
	// Actor fails with ErrUnknownUser if the person is not registered.
	Actor(ctx context.Context, userID string) (domain.Actor, error)
	// Grant replaces the profile the person had on that procedure type, if any.
	Grant(ctx context.Context, p domain.Permission) error
}

// DeniedAttempts keeps every refused request. It only grows.
type DeniedAttempts interface {
	Record(ctx context.Context, a domain.DeniedAttempt) error
}
