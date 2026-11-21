package acceptance

import (
	"context"
	"time"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
)

// The scenarios that are not about who publishes name these people as the
// authors of each version: they propose and approve in the Secretariat.
var (
	secretariatEditors    = []string{"ana", "carla"}
	secretariatPublishers = []string{"beto", "dario"}
)

// withSecretariat adds the Secretariat to the areas and wires auth as the
// server does, with the editors and publishers above as its members.
func withSecretariat(areas *authmem.AreaRepository, clock func() time.Time) (*authapp.AreaService, *authapp.MembershipService, error) {
	if err := authapp.EnsureSecretariat(context.Background(), areas); err != nil {
		return nil, nil, err
	}
	var members []authdomain.Membership
	for _, user := range secretariatEditors {
		members = append(members, authdomain.Membership{ID: "edit-" + user, UserID: user, Area: authinput.SecretariatArea, Permissions: []string{authinput.EditProcedureTypes}})
	}
	for _, user := range secretariatPublishers {
		members = append(members, authdomain.Membership{ID: "publish-" + user, UserID: user, Area: authinput.SecretariatArea, Permissions: []string{authinput.PublishProcedureTypes}})
	}
	memberships := authmem.NewMembershipRepository(members...)
	administrators := authmem.NewAdministratorRepository()
	return authapp.NewAreaService(areas, administrators, memberships, clock),
		authapp.NewMembershipService(areas, administrators, memberships, authmem.NewUserRepository(), func() string { return "id" }, clock),
		nil
}
