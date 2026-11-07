package output

import (
	"context"
	"errors"

	"backend/src/procedure/domain"
)

var ErrUnknownUser = errors.New("user not registered for access")

type ProcedureRepository interface {
	Add(ctx context.Context, p domain.Procedure) error
	Get(ctx context.Context, id string) (domain.Procedure, error)
	List(ctx context.Context) ([]domain.Procedure, error)
}

type Access interface {
	Register(ctx context.Context, userID string, kind domain.UserKind) error
	Actor(ctx context.Context, userID string) (domain.Actor, error)
	Grant(ctx context.Context, p domain.Permission) error
}

type DeniedAttempts interface {
	Record(ctx context.Context, a domain.DeniedAttempt) error
}
