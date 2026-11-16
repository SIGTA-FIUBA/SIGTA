package acceptance

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
)

type authorizationWorld struct {
	types             *application.ProcedureTypeService
	service           *application.ProcedureService
	memberships       *authapp.MembershipService
	areas             *authapp.AreaService
	storedAreas       *authmem.AreaRepository
	administrators    *authmem.Administrators
	storedMemberships *authmem.MembershipRepository
	procedures        *memory.ProcedureRepository
	documents         *memory.DocumentRepository
	now               time.Time

	typeIDs map[string]int64
	steps   map[string][]definition.Step
	nextID  int

	current    string
	created    []string
	membership authdomain.Membership

	user   string
	err    error
	view   input.ProcedureView
	listed []domain.Procedure
}

func registerAuthorizationSteps(sc *godog.ScenarioContext) *authorizationWorld {
	w := &authorizationWorld{
		administrators:    authmem.NewAdministrators("admin@fi.uba.ar"),
		storedMemberships: authmem.NewMembershipRepository(),
		procedures:        memory.NewProcedureRepository(),
		documents:         memory.NewDocumentRepository(),
		now:               time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC),
		typeIDs:           map[string]int64{},
		steps:             map[string][]definition.Step{},
	}

	sc.Given(`^the staff areas "([^"]*)" and "([^"]*)"$`, w.givenStaffAreas)
	sc.Given(`^a published "([^"]*)" procedure type with these steps:$`, w.givenProcedureType)
	sc.Given(`^"([^"]*)" is an administrator$`, w.givenAdministrator)
	sc.Given(`^the area "([^"]*)" has the permissions? (.+)$`, w.givenAreaPermissions)
	sc.Given(`^SIGTA has these memberships:$`, w.givenMemberships)
	sc.Given(`^"([^"]*)" grants "([^"]*)" a membership in "([^"]*)" with the permissions? "([^"]*)" until "([^"]*)"$`, w.givenMembershipUntil)
	sc.Given(`^a "([^"]*)" procedure whose open step is "([^"]*)"$`, w.givenProcedureOn)
	sc.Given(`^a "([^"]*)" procedure$`, w.givenProcedure)
	sc.Given(`^"([^"]*)" has a "([^"]*)" procedure$`, w.givenOwnProcedure)
	sc.Given(`^"([^"]*)" has a "([^"]*)" procedure whose open step is "([^"]*)"$`, w.givenOwnProcedureOn)
	sc.Given(`^the area "([^"]*)" has the role "([^"]*)" with the permissions? "([^"]*)"$`, w.givenRole)

	sc.When(`^"([^"]*)" grants "([^"]*)" a membership in "([^"]*)" with the permissions? "([^"]*)"(?: on "([^"]*)")?$`, w.grant)
	sc.Step(`^"([^"]*)" grants "([^"]*)" a membership in "([^"]*)" with the role "([^"]*)"(?: and the permissions? "([^"]*)")?$`, w.grantRole)
	sc.When(`^"([^"]*)" changes the permissions of the role "([^"]*)" of "([^"]*)" to "([^"]*)"$`, w.changeRole)
	sc.Step(`^"([^"]*)" creates the role "([^"]*)" named "([^"]*)" with "([^"]*)" in "([^"]*)"$`, w.createRole)
	sc.When(`^"([^"]*)" deletes the role "([^"]*)" of "([^"]*)"$`, w.deleteRole)
	sc.When(`^"([^"]*)" proposes and "([^"]*)" approves a version of "([^"]*)"$`, w.proposeAndApprove)
	sc.When(`^"([^"]*)" ends the membership of "([^"]*)" in "([^"]*)"$`, w.end)
	sc.When(`^it is "([^"]*)"$`, w.itIs)
	sc.When(`^"([^"]*)" creates the area "([^"]*)" named "([^"]*)"$`, w.createArea)
	sc.When(`^"([^"]*)" deletes the area "([^"]*)"$`, w.deleteArea)
	sc.When(`^"([^"]*)" creates the permission "([^"]*)" named "([^"]*)" in "([^"]*)"$`, w.createPermission)
	sc.When(`^"([^"]*)" deletes the permission "([^"]*)" of "([^"]*)"$`, w.deletePermission)
	sc.When(`^"([^"]*)" starts a "([^"]*)" procedure$`, w.start)
	sc.When(`^"([^"]*)" acts on that step$`, w.actOnOpenStep)
	sc.When(`^"([^"]*)" sends the approval directly to the API$`, w.sendApproval)
	sc.When(`^"([^"]*)" opens that procedure(?: by its identifier)?$`, w.open)
	sc.When(`^"([^"]*)" lists (?:their|the) procedures$`, w.list)
	sc.When(`^a new version of "([^"]*)" is published$`, w.publishNewVersion)

	sc.Then(`^the areas besides the Secretariat are (.+)$`, w.thenAreas)
	sc.Then(`^the area "([^"]*)" exists without anyone creating it$`, w.thenSystemArea)
	sc.Then(`^the role is denied$`, w.thenFails(authdomain.ErrAccessDenied))
	sc.Then(`^the role is not (?:created|deleted)$`, w.thenFails(authdomain.ErrInvalidRole))
	sc.Then(`^the role "([^"]*)" of "([^"]*)" records that "([^"]*)" changed it$`, w.thenRoleChangedBy)
	sc.Then(`^the version is (published|denied)$`, w.thenVersion)
	sc.Then(`^the area is denied$`, w.thenFails(authdomain.ErrAccessDenied))
	sc.Then(`^the area is not (?:created|deleted)$`, w.thenFails(authdomain.ErrInvalidArea))
	sc.Then(`^the area "([^"]*)" records that "([^"]*)" deleted it$`, w.thenAreaDeletedBy)
	sc.Then(`^the permission is denied$`, w.thenFails(authdomain.ErrAccessDenied))
	sc.Then(`^the permission is not (?:created|deleted)$`, w.thenFails(authdomain.ErrInvalidPermission))
	sc.Then(`^"([^"]*)" (can|cannot) grant "([^"]*)" a membership in "([^"]*)" with the (permissions?|role) "([^"]*)"$`, w.thenCanGrant)
	sc.Then(`^a procedure type with a step in "([^"]*)" that asks for "([^"]*)" (can|cannot) be published$`, w.thenPublishesWithArea)
	sc.Then(`^the procedure is started$`, w.thenStarted)
	sc.Then(`^the procedure is not started$`, w.thenNotStarted)
	sc.Then(`^"([^"]*)" (can|cannot) start a "([^"]*)" procedure$`, w.thenCanStart)
	sc.Then(`^the membership is granted by "([^"]*)" from now on$`, w.thenGranted)
	sc.Then(`^the membership is denied$`, w.thenFails(authdomain.ErrAccessDenied))
	sc.Then(`^the membership is not granted$`, w.thenFails(authdomain.ErrInvalidMembership))
	sc.Then(`^the membership records that "([^"]*)" ended it$`, w.thenEndedBy)
	sc.Then(`^the action is (allowed|denied)$`, w.thenAction)
	sc.Then(`^the procedure is shown$`, w.thenNoError)
	sc.Then(`^no action is offered on the "([^"]*)" step$`, w.thenNotOffered)
	sc.Then(`^access is denied$`, w.thenFails(domain.ErrProcedureNotFound))
	sc.Then(`^the response does not reveal whether the procedure exists$`, w.thenNotRevealed)
	sc.Then(`^they see (\d+) procedures?$`, w.thenSees)
	sc.Then(`^it is a "([^"]*)" procedure$`, w.thenOfType)
	sc.Then(`^none of them belongs to "([^"]*)"$`, w.thenNoneOf)
	sc.Then(`^the procedure stays on the "([^"]*)" step$`, w.thenStaysOn)
	return w
}

func (w *authorizationWorld) newID() string {
	w.nextID++
	return "P-" + strconv.Itoa(w.nextID)
}

// givenStaffAreas wires both modules as the server does: auth keeps the areas
// and memberships, and procedure asks auth through its public API (ADR-021).
// Besides the two areas there is the Secretariat, where the default authors
// can edit and publish procedure types.
func (w *authorizationWorld) givenStaffAreas(first, second string) error {
	areas := authmem.NewAreaRepository(
		authdomain.Area{ID: areaID(first), Name: first},
		authdomain.Area{ID: areaID(second), Name: second},
	)
	if err := authapp.EnsureSecretariat(context.Background(), areas); err != nil {
		return err
	}
	w.storedAreas = areas
	for _, author := range []authdomain.Membership{
		{ID: "M-proposer", UserID: defaultAuthors.ProposedBy, Area: authinput.SecretariatArea, Permissions: []string{authinput.EditProcedureTypes}},
		{ID: "M-approver", UserID: defaultAuthors.ApprovedBy, Area: authinput.SecretariatArea, Permissions: []string{authinput.PublishProcedureTypes}},
	} {
		if err := w.storedMemberships.Save(context.Background(), author); err != nil {
			return err
		}
	}
	clock := func() time.Time { return w.now }
	w.memberships = authapp.NewMembershipService(areas, w.administrators, w.storedMemberships, w.newID, clock)
	w.areas = authapp.NewAreaService(areas, w.administrators, w.storedMemberships, clock)
	w.types = application.NewProcedureTypeService(memory.NewProcedureVersions(), w.areas, w.memberships, clock)
	w.service = application.NewProcedureService(w.types, w.procedures, w.documents, w.memberships, w.newID, clock)
	return nil
}

func (w *authorizationWorld) definition(name, description string) definition.ProcedureTypeDefinition {
	staff := w.steps[name]
	steps := []definition.Step{{ID: "request", Name: "Solicitud", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Fields: []string{"reason"}}}
	if len(staff) > 0 {
		steps[0].Next = staff[0].ID
	}
	for i, s := range staff {
		if i+1 < len(staff) {
			s.Next = staff[i+1].ID
		}
		steps = append(steps, s)
	}
	return definition.ProcedureTypeDefinition{
		Name:        name,
		Description: description,
		Fields:      []definition.Field{{ID: "reason", Name: "Motivo", Type: definition.FieldLongText}},
		Steps:       steps,
	}
}

var defaultAuthors = version.Authorship{ProposedBy: "proponente", ApprovedBy: "aprobador"}

func (w *authorizationWorld) publish(name, description string) error {
	return w.publishBy(name, description, defaultAuthors)
}

// publishBy gives the procedure type an ID the first time it is published,
// even if publishing fails.
func (w *authorizationWorld) publishBy(name, description string, authors version.Authorship) error {
	typeID, known := w.typeIDs[name]
	if !known {
		typeID = int64(len(w.typeIDs) + 1)
		w.typeIDs[name] = typeID
	}
	base := 0
	if known {
		base = w.latestVersion(name)
	}
	_, err := w.types.Publish(context.Background(), input.Publication{
		ProcedureType: typeID,
		Base:          base,
		Definition:    w.definition(name, description),
		Authorship:    authors,
	})
	return err
}

func (w *authorizationWorld) latestVersion(name string) int {
	entries, err := w.types.Catalog(context.Background())
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if e.ProcedureType == w.typeIDs[name] {
			return e.Version
		}
	}
	return 0
}

// latestDigest is the digest of the latest version, which a procedure keeps
// along with its number.
func (w *authorizationWorld) latestDigest(name string) string {
	v, err := w.types.Version(context.Background(), w.typeID(name), w.latestVersion(name))
	if err != nil {
		return ""
	}
	return v.Digest
}

func (w *authorizationWorld) typeID(name string) int64 {
	if id, ok := w.typeIDs[name]; ok {
		return id
	}
	return 999
}

func (w *authorizationWorld) scope(name string) []int64 {
	if name == "" {
		return nil
	}
	return []int64{w.typeID(name)}
}

func (w *authorizationWorld) givenProcedureType(name string, table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		id := row.Cells[0].Value
		w.steps[name] = append(w.steps[name], definition.Step{
			ID:          id,
			Name:        id,
			Role:        definition.RoleStaff,
			Area:        areaID(row.Cells[1].Value),
			Permissions: list(row.Cells[2].Value),
			Action:      definition.ActionConfirmExternalFact,
		})
	}
	return w.publish(name, "Descripción de "+name)
}

func (w *authorizationWorld) givenAdministrator(user string) error {
	administrator, err := w.administrators.IsAdministrator(context.Background(), user)
	if err != nil {
		return err
	}
	if !administrator {
		return fmt.Errorf("%s is not an administrator", user)
	}
	return nil
}

func (w *authorizationWorld) givenMemberships(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		_, err := w.memberships.Grant(context.Background(), "admin@fi.uba.ar", authinput.MembershipGrant{
			UserID:      row.Cells[0].Value,
			Area:        areaID(row.Cells[1].Value),
			Permissions: list(row.Cells[2].Value),
			Scope:       w.scope(row.Cells[3].Value),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *authorizationWorld) givenMembershipUntil(administrator, user, area, permissions, until string) error {
	end, err := time.Parse(time.DateOnly, until)
	if err != nil {
		return err
	}
	_, err = w.memberships.Grant(context.Background(), administrator, authinput.MembershipGrant{
		UserID:      user,
		Area:        areaID(area),
		Permissions: list(permissions),
		Until:       end,
	})
	return err
}

func (w *authorizationWorld) givenProcedureOn(name, step string) error {
	return w.givenOwnProcedureOn("applicant@fi.uba.ar", name, step)
}

func (w *authorizationWorld) givenOwnProcedureOn(applicant, name, step string) error {
	p := domain.Procedure{
		ID:            w.newID(),
		ProcedureType: w.typeID(name),
		Version:       w.latestVersion(name),
		VersionDigest: w.latestDigest(name),
		Applicant:     applicant,
		OpenStep:      step,
	}
	if err := w.procedures.Add(context.Background(), p, nil); err != nil {
		return err
	}
	w.current = p.ID
	w.created = append(w.created, p.ID)
	return nil
}

func (w *authorizationWorld) givenProcedure(name string) error {
	return w.givenProcedureOn(name, w.steps[name][0].ID)
}

func (w *authorizationWorld) givenOwnProcedure(user, name string) error {
	p, err := w.service.Start(context.Background(), user, w.typeID(name))
	if err != nil {
		return err
	}
	w.current = p.ID
	w.created = append(w.created, p.ID)
	return nil
}

func (w *authorizationWorld) grant(administrator, user, area, permissions, scope string) {
	w.user = administrator
	w.membership, w.err = w.memberships.Grant(context.Background(), administrator, authinput.MembershipGrant{
		UserID:      user,
		Area:        areaID(area),
		Permissions: list(permissions),
		Scope:       w.scope(scope),
	})
}

func (w *authorizationWorld) end(administrator, user, area string) error {
	ctx := context.Background()
	all, err := w.storedMemberships.ListByUser(ctx, user)
	if err != nil {
		return err
	}
	for _, m := range all {
		if m.Area == areaID(area) {
			w.membership = m
			return w.memberships.End(ctx, administrator, m.ID)
		}
	}
	return fmt.Errorf("%s has no membership in %s", user, area)
}

func (w *authorizationWorld) itIs(date string) error {
	day, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return err
	}
	w.now = day.Add(12 * time.Hour)
	return nil
}

func (w *authorizationWorld) start(user, name string) {
	w.user = user
	_, w.err = w.service.Start(context.Background(), user, w.typeID(name))
}

func (w *authorizationWorld) act(user, procedureID, step string) error {
	w.user = user
	return w.service.Act(context.Background(), user, procedureID, step)
}

func (w *authorizationWorld) actOnOpenStep(user string) error {
	p, err := w.procedures.Get(context.Background(), w.current)
	if err != nil {
		return err
	}
	w.err = w.act(user, p.ID, p.OpenStep)
	return nil
}

func (w *authorizationWorld) sendApproval(user string) {
	w.err = w.act(user, w.current, "approve")
}

func (w *authorizationWorld) open(user string) {
	w.user = user
	w.view, w.err = w.service.Open(context.Background(), user, w.current)
}

func (w *authorizationWorld) list(user string) {
	w.listed, w.err = w.service.List(context.Background(), user)
}

func (w *authorizationWorld) publishNewVersion(name string) error {
	return w.publish(name, "Otra descripción de "+name)
}

func (w *authorizationWorld) thenNoError() error {
	return w.err
}

func (w *authorizationWorld) thenFails(want error) func() error {
	return func() error {
		if !errors.Is(w.err, want) {
			return fmt.Errorf("got %w, want %w", w.err, want)
		}
		return nil
	}
}

func (w *authorizationWorld) thenStarted() error {
	if w.err != nil {
		return w.err
	}
	all, err := w.procedures.List(context.Background())
	if err != nil {
		return err
	}
	for _, p := range all {
		if p.Applicant == w.user {
			return nil
		}
	}
	return fmt.Errorf("%s started no procedure", w.user)
}

func (w *authorizationWorld) thenCanStart(user, can, name string) error {
	w.start(user, name)
	if can == "can" {
		return w.thenStarted()
	}
	return w.thenNotStarted()
}

func (w *authorizationWorld) thenGranted(administrator string) error {
	if w.err != nil {
		return w.err
	}
	stored, err := w.storedMemberships.Get(context.Background(), w.membership.ID)
	if err != nil {
		return err
	}
	if stored.GrantedBy != administrator || !stored.From.Equal(w.now) || !stored.Until.IsZero() {
		return fmt.Errorf("stored %+v, want granted by %s from %s with no end", stored, administrator, w.now)
	}
	return nil
}

func (w *authorizationWorld) thenEndedBy(administrator string) error {
	stored, err := w.storedMemberships.Get(context.Background(), w.membership.ID)
	if err != nil {
		return err
	}
	if stored.EndedBy != administrator || !stored.Until.Equal(w.now) {
		return fmt.Errorf("stored %+v, want ended by %s at %s", stored, administrator, w.now)
	}
	return nil
}

func (w *authorizationWorld) thenNotStarted() error {
	if !errors.Is(w.err, domain.ErrAccessDenied) {
		return fmt.Errorf("got %w, want %w", w.err, domain.ErrAccessDenied)
	}
	all, err := w.procedures.List(context.Background())
	if err != nil {
		return err
	}
	for _, p := range all {
		if p.Applicant == w.user {
			return fmt.Errorf("procedure %s was started by %s", p.ID, w.user)
		}
	}
	return nil
}

func (w *authorizationWorld) thenAction(result string) error {
	if result == "allowed" {
		return w.err
	}
	if !errors.Is(w.err, domain.ErrAccessDenied) && !errors.Is(w.err, domain.ErrProcedureNotFound) {
		return fmt.Errorf("got %w, want a denial", w.err)
	}
	return nil
}

func (w *authorizationWorld) thenNotOffered(step string) error {
	for _, a := range w.view.Actions {
		if a == step {
			return fmt.Errorf("offered %q", step)
		}
	}
	return nil
}

func (w *authorizationWorld) thenNotRevealed() error {
	_, missing := w.service.Open(context.Background(), w.user, "P-missing")
	if !errors.Is(missing, domain.ErrProcedureNotFound) {
		return fmt.Errorf("a missing procedure got %w", missing)
	}
	existing := strings.ReplaceAll(w.err.Error(), w.current, "ID")
	absent := strings.ReplaceAll(missing.Error(), "P-missing", "ID")
	if existing != absent {
		return fmt.Errorf("answers differ: %q and %q", existing, absent)
	}
	return nil
}

func (w *authorizationWorld) thenSees(n int) error {
	if w.err != nil {
		return w.err
	}
	if len(w.listed) != n {
		return fmt.Errorf("sees %d procedures, want %d", len(w.listed), n)
	}
	return nil
}

func (w *authorizationWorld) thenOfType(name string) error {
	for _, p := range w.listed {
		if p.ProcedureType != w.typeID(name) {
			return fmt.Errorf("sees %s of procedure type %d", p.ID, p.ProcedureType)
		}
	}
	return nil
}

func (w *authorizationWorld) thenNoneOf(user string) error {
	for _, p := range w.listed {
		if p.Applicant == user {
			return fmt.Errorf("sees %s, which belongs to %s", p.ID, user)
		}
	}
	return nil
}

func (w *authorizationWorld) thenStaysOn(step string) error {
	p, err := w.procedures.Get(context.Background(), w.current)
	if err != nil {
		return err
	}
	if p.OpenStep != step {
		return fmt.Errorf("open step is %q, want %q", p.OpenStep, step)
	}
	return nil
}

func (w *authorizationWorld) createArea(administrator, id, name string) {
	w.user = administrator
	_, w.err = w.areas.Create(context.Background(), administrator, authdomain.Area{ID: areaID(id), Name: name})
}

func (w *authorizationWorld) deleteArea(administrator, name string) {
	w.user = administrator
	w.err = w.areas.Delete(context.Background(), administrator, areaID(name))
}

var quoted = regexp.MustCompile(`"([^"]*)"`)

// list reads "a, b" as the permissions a and b, and "" as none.
func list(cell string) []string {
	var items []string
	for _, item := range strings.Split(cell, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// givenAreaPermissions creates each quoted permission, named as its ID.
func (w *authorizationWorld) givenAreaPermissions(area, permissions string) error {
	for _, m := range quoted.FindAllStringSubmatch(permissions, -1) {
		if _, err := w.areas.CreatePermission(context.Background(), "admin@fi.uba.ar", areaID(area), authdomain.Permission{ID: m[1], Name: m[1]}); err != nil {
			return err
		}
	}
	return nil
}

func (w *authorizationWorld) createPermission(administrator, id, name, area string) {
	w.user = administrator
	_, w.err = w.areas.CreatePermission(context.Background(), administrator, areaID(area), authdomain.Permission{ID: id, Name: name})
}

func (w *authorizationWorld) deletePermission(administrator, id, area string) {
	w.user = administrator
	w.err = w.areas.DeletePermission(context.Background(), administrator, areaID(area), id)
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

func (w *authorizationWorld) thenAreaDeletedBy(name, administrator string) error {
	area, err := w.storedAreas.Get(context.Background(), areaID(name))
	if err != nil {
		return err
	}
	if area.DeletedBy != administrator || !area.DeletedAt.Equal(w.now) {
		return fmt.Errorf("stored %+v, want deleted by %s at %s", area, administrator, w.now)
	}
	return nil
}

func (w *authorizationWorld) thenCanGrant(administrator, can, user, area, kind, ids string) error {
	g := authinput.MembershipGrant{UserID: user, Area: areaID(area)}
	if kind == "role" {
		g.Roles = list(ids)
	} else {
		g.Permissions = list(ids)
	}
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

func (w *authorizationWorld) thenPublishesWithArea(area, permissions, can string) error {
	w.steps["new procedure type"] = []definition.Step{{
		ID:          "review",
		Name:        "Revisión",
		Role:        definition.RoleStaff,
		Area:        areaID(area),
		Permissions: list(permissions),
		Action:      definition.ActionConfirmExternalFact,
	}}
	err := w.publish("new procedure type", "Descripción del trámite nuevo")
	if can == "can" {
		return err
	}
	if !errors.Is(err, definition.ErrInvalidDefinition) {
		return fmt.Errorf("got %w, want %w", err, definition.ErrInvalidDefinition)
	}
	return nil
}

func (w *authorizationWorld) givenRole(area, role, permissions string) error {
	_, err := w.areas.CreateRole(context.Background(), "admin@fi.uba.ar", areaID(area), authinput.NewRole{ID: role, Name: role, Permissions: list(permissions)})
	return err
}

func (w *authorizationWorld) grantRole(administrator, user, area, role, permissions string) {
	w.user = administrator
	w.membership, w.err = w.memberships.Grant(context.Background(), administrator, authinput.MembershipGrant{
		UserID:      user,
		Area:        areaID(area),
		Roles:       []string{role},
		Permissions: list(permissions),
	})
}

func (w *authorizationWorld) changeRole(administrator, role, area, permissions string) error {
	_, err := w.areas.UpdateRolePermissions(context.Background(), administrator, areaID(area), role, list(permissions))
	return err
}

func (w *authorizationWorld) createRole(administrator, id, name, permissions, area string) {
	w.user = administrator
	_, w.err = w.areas.CreateRole(context.Background(), administrator, areaID(area), authinput.NewRole{ID: id, Name: name, Permissions: list(permissions)})
}

func (w *authorizationWorld) deleteRole(administrator, role, area string) {
	w.user = administrator
	w.err = w.areas.DeleteRole(context.Background(), administrator, areaID(area), role)
}

// proposeAndApprove publishes a version that changes the description, so that
// it is a new version whenever the authors are allowed.
func (w *authorizationWorld) proposeAndApprove(proposer, approver, name string) {
	w.err = w.publishBy(name, "Versión de "+name+" propuesta por "+proposer, version.Authorship{ProposedBy: proposer, ApprovedBy: approver})
}

func (w *authorizationWorld) thenVersion(result string) error {
	if result == "published" {
		return w.err
	}
	return w.thenFails(domain.ErrAccessDenied)()
}

func (w *authorizationWorld) thenSystemArea(area string) error {
	stored, err := w.storedAreas.Get(context.Background(), areaID(area))
	if err != nil {
		return err
	}
	if !stored.System || stored.Deleted() {
		return fmt.Errorf("stored %+v, want a system area", stored)
	}
	return nil
}

func (w *authorizationWorld) thenRoleChangedBy(role, area, administrator string) error {
	stored, err := w.storedAreas.Get(context.Background(), areaID(area))
	if err != nil {
		return err
	}
	i := slices.IndexFunc(stored.Roles, func(r authdomain.Role) bool { return r.ID == role })
	if i < 0 {
		return fmt.Errorf("area %s has no role %s", area, role)
	}
	revisions := stored.Roles[i].Revisions
	last := revisions[len(revisions)-1]
	if len(revisions) < 2 || last.By != administrator || !last.At.Equal(w.now) {
		return fmt.Errorf("revisions %+v, want the last one by %s at %s", revisions, administrator, w.now)
	}
	return nil
}
