package acceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cucumber/godog"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
)

// administrationWorld adds to the authorization scenarios what
// staff_administration.feature needs: the people who signed in, the
// administration of others and what an administrator lists.
type administrationWorld struct {
	*authorizationWorld
	users          *authmem.UserRepository
	administration *authapp.AdministratorService
	people         *authapp.PeopleService

	listed []authdomain.Area
	found  authdomain.User
}

// administrationRefusals are the reasons a scenario expects, in its words.
var administrationRefusals = map[string]error{
	"it ended": authdomain.ErrMembershipEnded,
	"the area has memberships that have not ended":              authdomain.ErrAreaHasMemberships,
	"the area comes with SIGTA":                                 authdomain.ErrSystemArea,
	"the permission comes with SIGTA":                           authdomain.ErrSystemPermission,
	"the permission is held by memberships that have not ended": authdomain.ErrPermissionHeld,
	"the permission is part of a role":                          authdomain.ErrPermissionInRole,
	"the role is held by memberships that have not ended":       authdomain.ErrRoleHeld,
	"nobody signed in with that account":                        authdomain.ErrUserNotFound,
	"the person is an administrator already":                    authdomain.ErrAlreadyAdministrator,
	"SIGTA would be left without administrators":                authdomain.ErrLastAdministrator,
}

func registerStaffAdministrationSteps(sc *godog.ScenarioContext, authorization *authorizationWorld) {
	w := &administrationWorld{authorizationWorld: authorization, users: authmem.NewUserRepository()}
	clock := func() time.Time { return w.now }
	ids := 0
	newID := func() string {
		ids++
		return fmt.Sprintf("A-%d", ids)
	}
	w.administration = authapp.NewAdministratorService(w.administrators, w.users, newID, clock)
	w.people = authapp.NewPeopleService(w.users, w.administrators)

	sc.Given(`^"([^"]*)" signed in to SIGTA$`, w.givenSignedIn)
	sc.When(`^"([^"]*)" lists the areas to administer$`, w.listAreas)
	sc.When(`^"([^"]*)" lists the memberships of "([^"]*)"$`, w.listMemberships)
	sc.When(`^"([^"]*)" looks for "([^"]*)"$`, w.find)
	sc.When(`^"([^"]*)" lists the administrators$`, w.listAdministrators)
	sc.When(`^"([^"]*)" ends that membership again$`, w.endAgain)
	sc.When(`^"([^"]*)" changes the membership of "([^"]*)" in "([^"]*)" to the permissions "([^"]*)"$`, w.change)
	sc.Step(`^"([^"]*)" grants the administration to "([^"]*)"$`, w.grantAdministration)
	sc.When(`^"([^"]*)" ends the administration of "([^"]*)"$`, w.endAdministration)
	sc.Then(`^the area "([^"]*)" is listed with the permissions "([^"]*)" and the roles "([^"]*)"$`, w.thenListed)
	sc.Then(`^they find "([^"]*)"$`, w.thenFound)
	sc.Then(`^nobody is found$`, w.thenFails(authdomain.ErrUserNotFound))
	sc.Then(`^it is denied$`, w.thenFails(authdomain.ErrAccessDenied))
	sc.Then(`^(?:it|the membership) is refused because (.+)$`, w.thenRefused)
	sc.Then(`^the memberships of "([^"]*)" are:$`, w.thenMemberships)
	sc.Then(`^no membership of "([^"]*)" is listed$`, w.thenNoMemberships)
	sc.Then(`^the area "([^"]*)" and its permission "([^"]*)" record that "([^"]*)" created them on "([^"]*)"$`, w.thenCreatedBy)
	sc.Then(`^the administrators are (.+)$`, w.thenAdministrators)
	sc.Then(`^the administration of "([^"]*)" records that "([^"]*)" granted it$`, w.thenAdministrationGrantedBy)
	sc.Then(`^"([^"]*)" is an administrator who is not staff$`, w.thenAdministratorNotStaff)
	sc.Then(`^"([^"]*)" can no longer administer$`, w.thenNoLongerAdministers)
}

// givenSignedIn registers the person with their email as ID, as the other
// scenarios name staff.
func (w *administrationWorld) givenSignedIn(email string) {
	w.users.Add(authdomain.User{ID: email, Email: email, Name: email})
}

func (w *administrationWorld) listAreas(user string) {
	w.listed, w.err = w.areas.ListToAdminister(context.Background(), user)
}

func (w *administrationWorld) listMemberships(user, area string) {
	_, w.err = w.memberships.ListByArea(context.Background(), user, idOf(area))
}

func (w *administrationWorld) find(user, email string) {
	w.found, w.err = w.people.Find(context.Background(), user, email)
}

func (w *administrationWorld) listAdministrators(user string) {
	_, w.err = w.administration.List(context.Background(), user)
}

func (w *administrationWorld) endAgain(administrator string) {
	w.err = w.memberships.End(context.Background(), administrator, w.membership.ID)
}

// change changes the only membership of the person in the area that has not
// ended, and keeps the old one to check how it ended.
func (w *administrationWorld) change(administrator, user, area, permissions string) error {
	all, err := w.storedMemberships.ListByUser(context.Background(), user)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(all, func(m authdomain.Membership) bool { return m.Area == idOf(area) })
	if i < 0 {
		return fmt.Errorf("%s has no membership in %s", user, area)
	}
	w.membership = all[i]
	_, w.err = w.memberships.Change(context.Background(), administrator, all[i].ID, nil, list(permissions))
	return nil
}

func (w *administrationWorld) grantAdministration(administrator, user string) {
	_, w.err = w.administration.Grant(context.Background(), administrator, user)
}

func (w *administrationWorld) endAdministration(administrator, user string) error {
	active, err := w.administrators.ListActive(context.Background(), w.now)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(active, func(a authdomain.Administrator) bool { return a.UserID == user })
	if i < 0 {
		return fmt.Errorf("%s is not an administrator", user)
	}
	w.err = w.administration.End(context.Background(), administrator, active[i].ID)
	return nil
}

func (w *administrationWorld) thenListed(area, permissions, roles string) error {
	if w.err != nil {
		return w.err
	}
	i := slices.IndexFunc(w.listed, func(a authdomain.Area) bool { return a.ID == idOf(area) })
	if i < 0 {
		return fmt.Errorf("%s is not listed", area)
	}
	var gotPermissions, gotRoles []string
	for _, p := range w.listed[i].ActivePermissions() {
		gotPermissions = append(gotPermissions, p.ID)
	}
	for _, r := range w.listed[i].ActiveRoles() {
		gotRoles = append(gotRoles, r.ID)
	}
	if !slices.Equal(gotPermissions, list(permissions)) || !slices.Equal(gotRoles, list(roles)) {
		return fmt.Errorf("%s has the permissions %q and the roles %q", area, gotPermissions, gotRoles)
	}
	return nil
}

func (w *administrationWorld) thenFound(email string) error {
	if w.err != nil {
		return w.err
	}
	if w.found.Email != email {
		return fmt.Errorf("found %+v, want %s", w.found, email)
	}
	return nil
}

func (w *administrationWorld) thenRefused(reason string) error {
	want, ok := administrationRefusals[reason]
	if !ok {
		return fmt.Errorf("no error for the reason %q", reason)
	}
	return w.thenFails(want)()
}

func (w *administrationWorld) thenMemberships(area string, table *godog.Table) error {
	memberships, err := w.memberships.ListByArea(context.Background(), scenarioAdministrator, idOf(area))
	if err != nil {
		return err
	}
	got := make([]string, 0, len(memberships))
	for _, m := range memberships {
		until := ""
		if !m.Until.IsZero() {
			until = m.Until.Format(time.DateOnly)
		}
		got = append(got, strings.Join([]string{m.UserID, strings.Join(m.Roles, ", "), strings.Join(m.Permissions, ", "), m.From.Format(time.DateOnly), until, m.GrantedBy}, " | "))
	}
	want := make([]string, 0, len(table.Rows)-1)
	for _, row := range table.Rows[1:] {
		cells := make([]string, 0, len(row.Cells))
		for _, c := range row.Cells {
			cells = append(cells, c.Value)
		}
		want = append(want, strings.Join(cells, " | "))
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("the memberships are %q, want %q", got, want)
	}
	return nil
}

func (w *administrationWorld) thenNoMemberships(area string) error {
	memberships, err := w.memberships.ListByArea(context.Background(), scenarioAdministrator, idOf(area))
	if err != nil {
		return err
	}
	if len(memberships) > 0 {
		return fmt.Errorf("listed %+v, want none", memberships)
	}
	return nil
}

func (w *administrationWorld) thenCreatedBy(area, permission, administrator, date string) error {
	stored, err := w.storedAreas.Get(context.Background(), idOf(area))
	if err != nil {
		return err
	}
	i := slices.IndexFunc(stored.Permissions, func(p authdomain.Permission) bool { return p.ID == permission })
	if i < 0 {
		return fmt.Errorf("%s has no permission %s", area, permission)
	}
	p := stored.Permissions[i]
	if stored.CreatedBy != administrator || stored.CreatedAt.Format(time.DateOnly) != date || p.CreatedBy != administrator || p.CreatedAt.Format(time.DateOnly) != date {
		return fmt.Errorf("stored %+v, want it and %s created by %s on %s", stored, permission, administrator, date)
	}
	return nil
}

func (w *administrationWorld) thenAdministrators(names string) error {
	var want []string
	for _, m := range quoted.FindAllStringSubmatch(names, -1) {
		want = append(want, m[1])
	}
	active, err := w.administration.List(context.Background(), want[0])
	if err != nil {
		return err
	}
	got := make([]string, 0, len(active))
	for _, a := range active {
		got = append(got, a.UserID)
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("the administrators are %q, want %q", got, want)
	}
	return nil
}

func (w *administrationWorld) thenAdministrationGrantedBy(user, administrator string) error {
	active, err := w.administrators.ListActive(context.Background(), w.now)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(active, func(a authdomain.Administrator) bool { return a.UserID == user })
	if i < 0 || active[i].GrantedBy != administrator || !active[i].From.Equal(w.now) {
		return fmt.Errorf("the administrators are %+v, want %s granted by %s at %s", active, user, administrator, w.now)
	}
	return nil
}

func (w *administrationWorld) thenAdministratorNotStaff(user string) error {
	if w.err != nil {
		return w.err
	}
	access, err := w.memberships.Access(context.Background(), user)
	if err != nil {
		return err
	}
	if !access.Administrator || access.IsStaff() {
		return fmt.Errorf("access = %+v, want an administrator who is not staff", access)
	}
	return nil
}

func (w *administrationWorld) thenNoLongerAdministers(user string) error {
	if w.err != nil {
		return w.err
	}
	if _, err := w.areas.ListToAdminister(context.Background(), user); !errors.Is(err, authdomain.ErrAccessDenied) {
		return fmt.Errorf("got %w, want %w", err, authdomain.ErrAccessDenied)
	}
	return nil
}
