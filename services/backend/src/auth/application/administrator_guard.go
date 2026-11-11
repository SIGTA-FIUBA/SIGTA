package application

import (
	"context"
	"fmt"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

type administratorGuard struct {
	administrators output.Administrators
}

func (g administratorGuard) requireAdministrator(ctx context.Context, userID, action string) error {
	administrator, err := g.administrators.IsAdministrator(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s %s: %w", userID, action, err)
	}
	if !administrator {
		return fmt.Errorf("%s %s: %w", userID, action, domain.ErrAccessDenied)
	}
	return nil
}
