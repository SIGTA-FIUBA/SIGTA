package acceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/cucumber/godog"

	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
)

// registerStaffAreaSteps adds the steps of staff_areas.feature, where an
// administrator creates and deletes areas, permissions and roles, to the
// world of the authorization scenarios.
func registerStaffAreaSteps(sc *godog.ScenarioContext, w *authorizationWorld) {
	sc.Step(`^"([^"]*)" creates the role "([^"]*)" named "([^"]*)" with "([^"]*)" in "([^"]*)"$`, w.createRole)
	sc.When(`^"([^"]*)" deletes the role "([^"]*)" of "([^"]*)"$`, w.deleteRole)
	sc.When(`^"([^"]*)" creates the area "([^"]*)" named "([^"]*)"$`, w.createArea)
	sc.When(`^"([^"]*)" deletes the area "([^"]*)"$`, w.deleteArea)
	sc.When(`^"([^"]*)" creates the permission "([^"]*)" named "([^"]*)" in "([^"]*)"$`, w.createPermission)
	sc.When(`^"([^"]*)" deletes the permission "([^"]*)" of "([^"]*)"$`, w.deletePermission)
	sc.Then(`^the areas besides the Secretariat are (.+)$`, w.thenAreas)
	sc.Then(`^the area "([^"]*)" exists without anyone creating it$`, w.thenBuiltInArea)
	sc.Then(`^the role is denied$`, w.thenFails(authdomain.ErrAccessDenied))
	sc.Then(`^the role is not (?:created|deleted)$`, w.thenFails(authdomain.ErrInvalidRole))
	sc.Then(`^the area is denied$`, w.thenFails(authdomain.ErrAccessDenied))
	sc.Then(`^the area is not (?:created|deleted)$`, w.thenFails(authdomain.ErrInvalidArea))
	sc.Then(`^the area "([^"]*)" records that "([^"]*)" deleted it$`, w.thenAreaDeletedBy)
	sc.Then(`^the permission is denied$`, w.thenFails(authdomain.ErrAccessDenied))
	sc.Then(`^the permission is not (?:created|deleted)$`, w.thenFails(authdomain.ErrInvalidPermission))
	sc.Then(`^"([^"]*)" (can|cannot) grant "([^"]*)" a membership in "([^"]*)" with the (permissions?|role) "([^"]*)"$`, w.thenCanGrant)
}

func (w *authorizationWorld) createRole(administrator, id, name, permissions, area string) {
	w.user = administrator
	_, w.err = w.areas.CreateRole(context.Background(), administrator, idOf(area), authinput.NewRole{ID: id, Name: name, Permissions: list(permissions)})
}

func (w *authorizationWorld) deleteRole(administrator, role, area string) {
	w.user = administrator
	w.err = w.areas.DeleteRole(context.Background(), administrator, idOf(area), role)
}

func (w *authorizationWorld) createArea(administrator, id, name string) {
	w.user = administrator
	_, w.err = w.areas.Create(context.Background(), administrator, authdomain.Area{ID: idOf(id), Name: name})
}

func (w *authorizationWorld) deleteArea(administrator, name string) {
	w.user = administrator
	w.err = w.areas.Delete(context.Background(), administrator, idOf(name))
}

func (w *authorizationWorld) createPermission(administrator, id, name, area string) {
	w.user = administrator
	_, w.err = w.areas.CreatePermission(context.Background(), administrator, idOf(area), authdomain.Permission{ID: id, Name: name})
}

func (w *authorizationWorld) deletePermission(administrator, id, area string) {
	w.user = administrator
	w.err = w.areas.DeletePermission(context.Background(), administrator, idOf(area), id)
}

func (w *authorizationWorld) thenAreas(list string) error {
	var want []string
	for _, m := range quoted.FindAllStringSubmatch(list, -1) {
		want = append(want, m[1])
	}
	areas, err := w.areas.List(context.Background())
	if err != nil {
		return err
	}
	got := make([]string, 0, len(areas))
	for _, a := range areas {
		if !a.System {
			got = append(got, a.Name)
		}
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("the areas are %q, want %q", got, want)
	}
	return nil
}

func (w *authorizationWorld) thenBuiltInArea(area string) error {
	stored, err := w.storedAreas.Get(context.Background(), idOf(area))
	if err != nil {
		return err
	}
	if !stored.System || stored.Deleted() {
		return fmt.Errorf("stored %+v, want a system area", stored)
	}
	return nil
}

func (w *authorizationWorld) thenAreaDeletedBy(name, administrator string) error {
	area, err := w.storedAreas.Get(context.Background(), idOf(name))
	if err != nil {
		return err
	}
	if area.DeletedBy != administrator || !area.DeletedAt.Equal(w.now) {
		return fmt.Errorf("stored %+v, want deleted by %s at %s", area, administrator, w.now)
	}
	return nil
}

func (w *authorizationWorld) thenCanGrant(administrator, can, user, area, kind, ids string) error {
	g := authinput.MembershipGrant{UserID: user, Area: idOf(area)}
	if kind == "role" {
		g.Roles = list(ids)
	} else {
		g.Permissions = list(ids)
	}
	w.signIn(user)
	_, err := w.memberships.Grant(context.Background(), administrator, g)
	if can == "can" {
		return err
	}
	if err == nil {
		return fmt.Errorf("granted, want %w", authdomain.ErrInvalidMembership)
	}
	if !errors.Is(err, authdomain.ErrInvalidMembership) {
		return fmt.Errorf("got %w, want %w", err, authdomain.ErrInvalidMembership)
	}
	return nil
}
