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
	now               time.Time

	typeIDs map[string]int64
	steps   map[string][]domain.Step
	nextID  int

	current    string
	created    []string
	membership authdomain.Membership

	user   string
	err    error
	view   input.ProcedureView
	listed []domain.Procedure
}

func registerAuthorizationSteps(sc *godog.ScenarioContext) {
	w := &authorizationWorld{
		administrators:    authmem.NewAdministrators("admin@fi.uba.ar"),
		storedMemberships: authmem.NewMembershipRepository(),
		procedures:        memory.NewProcedureRepository(),
		now:               time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC),
		typeIDs:           map[string]int64{},
		steps:             map[string][]domain.Step{},
	}

	sc.Given(`^the staff areas "([^"]*)" and "([^"]*)"$`, w.givenStaffAreas)
	sc.Given(`^a published "([^"]*)" procedure type with these steps:$`, w.givenProcedureType)
	sc.Given(`^"([^"]*)" is an administrator$`, w.givenAdministrator)
	sc.Given(`^SIGTA has these memberships:$`, w.givenMemberships)
	sc.Given(`^"([^"]*)" grants "([^"]*)" a membership in "([^"]*)" with the "([^"]*)" profile until "([^"]*)"$`, w.givenMembershipUntil)
	sc.Given(`^a "([^"]*)" procedure whose open step is "([^"]*)"$`, w.givenProcedureOn)
	sc.Given(`^a "([^"]*)" procedure$`, w.givenProcedure)
	sc.Given(`^"([^"]*)" has a "([^"]*)" procedure$`, w.givenOwnProcedure)

	sc.When(`^"([^"]*)" grants "([^"]*)" a membership in "([^"]*)" with the "([^"]*)" profile(?: on "([^"]*)")?$`, w.grant)
	sc.When(`^"([^"]*)" ends the membership of "([^"]*)" in "([^"]*)"$`, w.end)
	sc.When(`^it is "([^"]*)"$`, w.itIs)
	sc.When(`^"([^"]*)" creates the area "([^"]*)" named "([^"]*)"$`, w.createArea)
	sc.When(`^"([^"]*)" deletes the area "([^"]*)"$`, w.deleteArea)
	sc.When(`^"([^"]*)" starts a "([^"]*)" procedure$`, w.start)
	sc.When(`^"([^"]*)" acts on that step$`, w.actOnOpenStep)
	sc.When(`^"([^"]*)" sends the approval directly to the API$`, w.sendApproval)
	sc.When(`^"([^"]*)" opens that procedure(?: by its identifier)?$`, w.open)
	sc.When(`^"([^"]*)" lists (?:their|the) procedures$`, w.list)
	sc.When(`^a new version of "([^"]*)" is published$`, w.publishNewVersion)

	sc.Then(`^the areas are (.+)$`, w.thenAreas)
	sc.Then(`^the area is denied$`, w.thenFails(authdomain.ErrAccessDenied))
	sc.Then(`^the area is not (?:created|deleted)$`, w.thenFails(authdomain.ErrInvalidArea))
	sc.Then(`^the area "([^"]*)" records that "([^"]*)" deleted it$`, w.thenAreaDeletedBy)
	sc.Then(`^"([^"]*)" (can|cannot) grant "([^"]*)" a membership in "([^"]*)"$`, w.thenCanGrant)
	sc.Then(`^a procedure type with a step in "([^"]*)" can be published$`, w.thenPublishesWithArea)
	sc.Then(`^the procedure is started$`, w.thenStarted)
	sc.Then(`^the procedure is not started$`, w.thenNotStarted)
	sc.Then(`^"([^"]*)" cannot start a "([^"]*)" procedure$`, w.thenCannotStart)
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
}

func (w *authorizationWorld) newID() string {
	w.nextID++
	return "P-" + strconv.Itoa(w.nextID)
}

// givenStaffAreas wires both modules as the server does: auth keeps the areas
// and memberships, and procedure asks auth through its public API (ADR-018).
func (w *authorizationWorld) givenStaffAreas(first, second string) error {
	areas := authmem.NewAreaRepository(
		authdomain.Area{ID: areaID(first), Name: first},
		authdomain.Area{ID: areaID(second), Name: second},
	)
	w.storedAreas = areas
	clock := func() time.Time { return w.now }
	w.memberships = authapp.NewMembershipService(areas, w.administrators, w.storedMemberships, w.newID, clock)
	w.areas = authapp.NewAreaService(areas, w.administrators, w.storedMemberships, clock)
	w.types = application.NewProcedureTypeService(memory.NewProcedureVersions(), w.areas, clock)
	w.service = application.NewProcedureService(w.types, w.procedures, w.memberships, w.newID)
	return nil
}

func (w *authorizationWorld) definition(name, description string) domain.Definition {
	staff := w.steps[name]
	steps := []domain.Step{{ID: "request", Name: "Solicitud", Role: domain.RoleApplicant, Action: domain.ActionEnterData, Fields: []string{"reason"}, Next: staff[0].ID}}
	for i, s := range staff {
		if i+1 < len(staff) {
			s.Next = staff[i+1].ID
		}
		steps = append(steps, s)
	}
	return domain.Definition{
		Name:        name,
		Description: description,
		Fields:      []domain.Field{{ID: "reason", Name: "Motivo", Type: domain.FieldLongText}},
		Steps:       steps,
	}
}

func (w *authorizationWorld) publish(name, description string) error {
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
		Authorship:    domain.Authorship{ProposedBy: "proponente", ApprovedBy: "aprobador"},
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
		w.steps[name] = append(w.steps[name], domain.Step{
			ID:      id,
			Name:    id,
			Role:    domain.RoleStaff,
			Area:    areaID(row.Cells[1].Value),
			Profile: domain.Profile(row.Cells[2].Value),
			Action:  domain.ActionConfirmExternalFact,
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
			UserID:  row.Cells[0].Value,
			Area:    areaID(row.Cells[1].Value),
			Profile: authdomain.Profile(row.Cells[2].Value),
			Scope:   w.scope(row.Cells[3].Value),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *authorizationWorld) givenMembershipUntil(administrator, user, area, profile, until string) error {
	end, err := time.Parse(time.DateOnly, until)
	if err != nil {
		return err
	}
	_, err = w.memberships.Grant(context.Background(), administrator, authinput.MembershipGrant{
		UserID:  user,
		Area:    areaID(area),
		Profile: authdomain.Profile(profile),
		Until:   end,
	})
	return err
}

func (w *authorizationWorld) givenProcedureOn(name, step string) error {
	p := domain.Procedure{
		ID:            w.newID(),
		ProcedureType: w.typeID(name),
		Version:       w.latestVersion(name),
		Applicant:     "applicant@fi.uba.ar",
		OpenStep:      step,
	}
	if err := w.procedures.Add(context.Background(), p); err != nil {
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

func (w *authorizationWorld) grant(administrator, user, area, profile, scope string) {
	w.user = administrator
	w.membership, w.err = w.memberships.Grant(context.Background(), administrator, authinput.MembershipGrant{
		UserID:  user,
		Area:    areaID(area),
		Profile: authdomain.Profile(profile),
		Scope:   w.scope(scope),
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

func (w *authorizationWorld) thenCannotStart(user, name string) error {
	w.start(user, name)
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
		got = append(got, a.Name)
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

func (w *authorizationWorld) thenCanGrant(administrator, can, user, area string) error {
	_, err := w.memberships.Grant(context.Background(), administrator, authinput.MembershipGrant{
		UserID:  user,
		Area:    areaID(area),
		Profile: authdomain.ProfileStaffOperative,
	})
	if can == "can" {
		return err
	}
	if !errors.Is(err, authdomain.ErrInvalidMembership) {
		return fmt.Errorf("got %v, want %v", err, authdomain.ErrInvalidMembership)
	}
	return nil
}

func (w *authorizationWorld) thenPublishesWithArea(area string) error {
	w.steps["new procedure type"] = []domain.Step{{
		ID:      "review",
		Name:    "Revisión",
		Role:    domain.RoleStaff,
		Area:    areaID(area),
		Profile: domain.ProfileStaffOperative,
		Action:  domain.ActionConfirmExternalFact,
	}}
	return w.publish("new procedure type", "Descripción del trámite nuevo")
}
