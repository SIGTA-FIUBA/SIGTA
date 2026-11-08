package application

import (
	"context"
	"fmt"
	"slices"
	"time"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

// authorizer asks auth what a person can do as staff (ADR-018) and records
// every denied attempt.
type authorizer struct {
	staff          authinput.StaffAuthorizer
	deniedAttempts output.DeniedAttemptLog
	clock          func() time.Time
}

func (a authorizer) deny(ctx context.Context, userID, action, procedureID string) error {
	attempt := domain.DeniedAttempt{UserID: userID, Action: action, Procedure: procedureID, At: a.clock().UTC()}
	if err := a.deniedAttempts.Record(ctx, attempt); err != nil {
		return fmt.Errorf("record denied attempt: %w", err)
	}
	return fmt.Errorf("%s %s: %w", userID, action, domain.ErrAccessDenied)
}

// canSee gives the applicant their own procedures, and staff those that pass
// through their areas within their scope.
func (a authorizer) canSee(userID string, reach []authinput.AreaReach, p domain.Procedure, c domain.Content) bool {
	if len(reach) == 0 {
		return p.Applicant == userID
	}
	return slices.ContainsFunc(reach, func(r authinput.AreaReach) bool {
		return (len(r.Scope) == 0 || slices.Contains(r.Scope, p.ProcedureType)) && c.PassesThrough(r.Area)
	})
}

// canAct gives the open step of the applicant area to the applicant of the
// procedure, and that of any other area to whoever auth says holds its
// profile there.
func (a authorizer) canAct(ctx context.Context, userID string, staff bool, p domain.Procedure, c domain.Content, stepID string) (bool, error) {
	if stepID != p.OpenStep {
		return false, nil
	}
	i := slices.IndexFunc(c.Definition.Steps, func(s domain.Step) bool { return s.ID == stepID })
	if i < 0 {
		return false, fmt.Errorf("procedure %s: open step %q is not in version %d", p.ID, stepID, p.Version)
	}
	step := c.Definition.Steps[i]
	area, ok := c.AreaOf(step)
	if !ok {
		return false, fmt.Errorf("procedure %s: area %q is not in version %d", p.ID, step.Area, p.Version)
	}
	if area.Kind == domain.AreaApplicant {
		return !staff && p.Applicant == userID, nil
	}
	holds, err := a.staff.Holds(ctx, userID, step.Area, p.ProcedureType, authinput.Profile(step.Profile))
	if err != nil {
		return false, fmt.Errorf("profile of %s in %s: %w", userID, step.Area, err)
	}
	return holds, nil
}
