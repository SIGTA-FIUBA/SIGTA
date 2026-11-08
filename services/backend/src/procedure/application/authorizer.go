package application

import (
	"context"
	"fmt"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

type authorizer struct {
	staff          output.StaffDirectory
	deniedAttempts output.DeniedAttemptLog
	clock          func() time.Time
}

func (a authorizer) actor(ctx context.Context, userID string) (domain.Actor, error) {
	actor, err := a.staff.Actor(ctx, userID)
	if err != nil {
		return domain.Actor{}, fmt.Errorf("actor %s: %w", userID, err)
	}
	return actor, nil
}

func (a authorizer) deny(ctx context.Context, actor domain.Actor, action, procedureID string) error {
	attempt := domain.DeniedAttempt{UserID: actor.UserID, Action: action, Procedure: procedureID, At: a.clock().UTC()}
	if err := a.deniedAttempts.Record(ctx, attempt); err != nil {
		return fmt.Errorf("record denied attempt: %w", err)
	}
	return fmt.Errorf("%s %s: %w", actor.UserID, action, domain.ErrAccessDenied)
}
