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

// canSee lets anyone see the procedures they started, and staff also those
// that go through their areas, within the scope of their memberships
// (ADR-020).
func canSee(userID string, memberOf []authinput.AreaScope, p domain.Procedure, c version.Content) bool {
	if p.Applicant == userID {
		return true
	}
	return slices.ContainsFunc(memberOf, func(area authinput.AreaScope) bool {
		coversType := len(area.Scope) == 0 || slices.Contains(area.Scope, p.ProcedureType)
		return coversType && c.PassesThrough(area.Area)
	})
}

// canAct leaves the steps of the applicant to the applicant, and the steps of
// staff to whoever holds their permissions, except the applicant: nobody
// moves their own procedure forward as staff.
func (s *ProcedureService) canAct(ctx context.Context, userID string, p domain.Procedure, c version.Content, stepID string) (bool, error) {
	if stepID != p.OpenStep {
		return false, nil
	}
	i := slices.IndexFunc(c.Definition.Steps, func(step definition.Step) bool { return step.ID == stepID })
	if i < 0 {
		return false, fmt.Errorf("procedure %s: open step %q is not in version %d", p.ID, stepID, p.Version)
	}
	step := c.Definition.Steps[i]
	isApplicant := p.Applicant == userID
	if step.Role == definition.RoleApplicant || isApplicant {
		return step.Role == definition.RoleApplicant && isApplicant, nil
	}
	holds, err := s.staffAuthorizer.Holds(ctx, userID, step.Area, p.ProcedureType, step.Permissions)
	if err != nil {
		return false, fmt.Errorf("permissions of %s in %s: %w", userID, step.Area, err)
	}
	return holds, nil
}
