package application

import (
	"context"
	"fmt"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

// administratorGuard lets only an administrator manage areas and memberships,
// and records who else tried.
type administratorGuard struct {
	administrators output.Administrators
	deniedAttempts output.DeniedAttemptLog
	clock          func() time.Time
}

func (g administratorGuard) requireAdministrator(ctx context.Context, userID, action string) error {
	administrator, err := g.administrators.IsAdministrator(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s %s: %w", userID, action, err)
	}
	if administrator {
		return nil
	}
	if err := g.deniedAttempts.Record(ctx, domain.DeniedAttempt{UserID: userID, Action: action, At: g.clock().UTC()}); err != nil {
		return fmt.Errorf("record denied attempt: %w", err)
	}
	return fmt.Errorf("%s %s: %w", userID, action, domain.ErrAccessDenied)
}
