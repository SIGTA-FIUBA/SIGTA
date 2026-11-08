// Package auth adapts the auth module to what procedure needs from it: the
// areas and the memberships of each person (ADR-014, ADR-017).
package auth

import (
	"context"
	"fmt"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var _ output.StaffDirectory = (*StaffDirectory)(nil)

type StaffDirectory struct {
	memberships authinput.Memberships
}

func NewStaffDirectory(memberships authinput.Memberships) *StaffDirectory {
	return &StaffDirectory{memberships: memberships}
}

func (d *StaffDirectory) Actor(ctx context.Context, userID string) (domain.Actor, error) {
	permissions, err := d.memberships.Permissions(ctx, userID)
	if err != nil {
		return domain.Actor{}, fmt.Errorf("permissions of %s: %w", userID, err)
	}
	actor := domain.Actor{UserID: userID}
	for _, m := range permissions.Memberships {
		actor.Memberships = append(actor.Memberships, domain.StaffMembership{
			Area:    m.Area,
			Profile: domain.Profile(m.Profile),
			Scope:   m.Scope,
		})
	}
	return actor, nil
}
