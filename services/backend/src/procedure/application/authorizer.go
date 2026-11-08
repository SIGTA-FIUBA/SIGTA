package application

import (
	"context"
	"fmt"
	"slices"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

func denied(userID, action string) error {
	return fmt.Errorf("%s %s: %w", userID, action, domain.ErrAccessDenied)
}

// canSee takes the areas where the person has an active membership: without
// any, they are an applicant (ADR-020).
func canSee(userID string, memberOf []authinput.AreaReach, p domain.Procedure, c version.Content) bool {
	isApplicant := len(memberOf) == 0
	if isApplicant {
		return p.Applicant == userID
	}
	return slices.ContainsFunc(memberOf, func(area authinput.AreaReach) bool {
		coversType := len(area.Scope) == 0 || slices.Contains(area.Scope, p.ProcedureType)
		return coversType && c.PassesThrough(area.Area)
	})
}

func (s *ProcedureService) canAct(ctx context.Context, userID string, isApplicant bool, p domain.Procedure, c version.Content, stepID string) (bool, error) {
	if stepID != p.OpenStep {
		return false, nil
	}
	i := slices.IndexFunc(c.Definition.Steps, func(step definition.Step) bool { return step.ID == stepID })
	if i < 0 {
		return false, fmt.Errorf("procedure %s: open step %q is not in version %d", p.ID, stepID, p.Version)
	}
	step := c.Definition.Steps[i]
	if step.Role == definition.RoleApplicant {
		return isApplicant && p.Applicant == userID, nil
	}
	holds, err := s.staffAuthorizer.Holds(ctx, userID, step.Area, p.ProcedureType, authinput.Profile(step.Profile))
	if err != nil {
		return false, fmt.Errorf("profile of %s in %s: %w", userID, step.Area, err)
	}
	return holds, nil
}
