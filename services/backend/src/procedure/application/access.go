package application

import (
	"context"
	"fmt"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

type gate struct {
	access output.Access
	denied output.DeniedAttempts
	now    func() time.Time
}

func (g gate) actor(ctx context.Context, userID string) (domain.Actor, error) {
	administrator, err := g.access.Administrator(ctx, userID)
	if err != nil {
		return domain.Actor{}, fmt.Errorf("actor %s: %w", userID, err)
	}
	memberships, err := g.access.Memberships(ctx, userID)
	if err != nil {
		return domain.Actor{}, fmt.Errorf("actor %s: %w", userID, err)
	}
	return domain.NewActor(userID, administrator, memberships, g.now()), nil
}

func (g gate) deny(ctx context.Context, actor domain.Actor, action, procedureID string) error {
	attempt := domain.DeniedAttempt{UserID: actor.UserID, Action: action, Procedure: procedureID, At: g.now().UTC()}
	if err := g.denied.Record(ctx, attempt); err != nil {
		return fmt.Errorf("record denied attempt: %w", err)
	}
	return fmt.Errorf("%s %s: %w", actor.UserID, action, domain.ErrAccessDenied)
}
