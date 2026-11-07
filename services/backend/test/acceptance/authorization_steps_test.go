package acceptance

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
)

// authorizationWorld is the state of one scenario of US07. It calls the
// service directly: the service is what the server enforces, whatever the
// interface offers.
type authorizationWorld struct {
	types      *application.ProcedureTypeService
	service    *application.ProcedureService
	procedures *memory.Procedures
	denied     *memory.DeniedAttempts
	now        time.Time

	typeIDs  map[string]int64
	steps    map[string][]string
	profiles map[string][]domain.Profile
	nextID   int

	// current is the procedure the last Given created, and created every one
	// of them.
	current string
	created []string

	// The result of the last When.
	user   string
	err    error
	errs   []error
	view   input.ProcedureView
	listed []domain.Procedure
}

func registerAuthorizationSteps(sc *godog.ScenarioContext) {
	w := &authorizationWorld{
		procedures: memory.NewProcedures(),
		denied:     memory.NewDeniedAttempts(),
		now:        time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC),
		typeIDs:    map[string]int64{},
		steps:      map[string][]string{},
		profiles:   map[string][]domain.Profile{},
	}
	areas := memory.NewAreaCatalog(
		domain.Area{ID: "applicant", Name: "Titular", Kind: domain.AreaApplicant},
		domain.Area{ID: "office", Name: "Secretaría", Kind: domain.AreaOffice},
	)
	clock := func() time.Time { return w.now }
	w.types = application.NewProcedureTypeService(memory.NewProcedureVersions(), areas, clock)
	w.service = application.NewProcedureService(w.types, w.procedures, memory.NewAccess(), w.denied, w.newID, clock)

	sc.Given(`^a published "([^"]*)" procedure type with these steps:$`, w.givenProcedureType)
	sc.Given(`^SIGTA has these users:$`, w.givenUsers)
	sc.Given(`^SIGTA has these staff permissions:$`, w.givenPermissions)
	sc.Given(`^a "([^"]*)" procedure whose open step is "([^"]*)"$`, w.givenProcedureOn)
	sc.Given(`^a "([^"]*)" procedure$`, w.givenProcedure)
	sc.Given(`^"([^"]*)" has a "([^"]*)" procedure$`, w.givenOwnProcedure)

	sc.When(`^a user is registered as "([^"]*)"$`, w.register)
	sc.When(`^a staff permission is granted on "([^"]*)" with the "([^"]*)" profile$`, w.grantNew)
	sc.When(`^a staff permission is granted to "([^"]*)" on "([^"]*)" with the "([^"]*)" profile$`, w.grant)
	sc.When(`^"([^"]*)" starts a "([^"]*)" procedure$`, w.start)
	sc.When(`^"([^"]*)" acts on that step$`, w.actOnOpenStep)
	sc.When(`^"([^"]*)" acts on each open step$`, w.actOnEach)
	sc.When(`^"([^"]*)" sends the approval directly to the API$`, w.sendApproval)
	sc.When(`^"([^"]*)" opens that procedure(?: by its identifier)?$`, w.open)
	sc.When(`^"([^"]*)" lists (?:their|the) procedures$`, w.list)
	sc.When(`^a new version of "([^"]*)" is published$`, w.publishNewVersion)

	sc.Then(`^the user is registered$`, w.thenNoError)
	sc.Then(`^the user is not registered$`, w.thenFails(domain.ErrUnknownUserKind))
	sc.Then(`^the permission is not granted$`, w.thenFails(domain.ErrInvalidPermission))
	sc.Then(`^the procedure is not started$`, w.thenNotStarted)
	sc.Then(`^the action is (allowed|denied)$`, w.thenAction)
	sc.Then(`^both actions are allowed$`, w.thenAllAllowed)
	sc.Then(`^the procedure is shown$`, w.thenNoError)
	sc.Then(`^no action is offered on the "([^"]*)" step$`, w.thenNotOffered)
	sc.Then(`^access is denied$`, w.thenFails(domain.ErrProcedureNotFound))
	sc.Then(`^the response does not reveal whether the procedure exists$`, w.thenNotRevealed)
	sc.Then(`^they see (\d+) procedures?$`, w.thenSees)
	sc.Then(`^it is a "([^"]*)" procedure$`, w.thenOfType)
	sc.Then(`^none of them belongs to "([^"]*)"$`, w.thenNoneOf)
	sc.Then(`^the procedure stays on the "([^"]*)" step$`, w.thenStaysOn)
	sc.Then(`^a denied attempt is recorded with who tried, what they tried(, on which procedure)? and when$`, w.thenDeniedRecorded)
	sc.Then(`^no denied attempt is recorded$`, w.thenNoDenied)
}

func (w *authorizationWorld) newID() string {
	w.nextID++
	return "P-" + strconv.Itoa(w.nextID)
}

// definition puts an applicant step first, since a procedure starts there,
// and then the staff steps of the table in order.
func (w *authorizationWorld) definition(name, description string) domain.Definition {
	staff := w.steps[name]
	steps := []domain.Step{{ID: "request", Name: "Solicitud", Area: "applicant", Action: domain.ActionEnterData, Fields: []string{"reason"}, Next: staff[0]}}
	profiles := w.profiles[name]
	for i, id := range staff {
		s := domain.Step{ID: id, Name: id, Area: "office", Profile: profiles[i], Action: domain.ActionConfirmExternalFact}
		if i+1 < len(staff) {
			s.Next = staff[i+1]
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

// typeID gives a procedure type that was never published an identifier no
// published one has.
func (w *authorizationWorld) typeID(name string) int64 {
	if id, ok := w.typeIDs[name]; ok {
		return id
	}
	return 999
}

func (w *authorizationWorld) givenProcedureType(name string, table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		w.steps[name] = append(w.steps[name], row.Cells[0].Value)
		w.profiles[name] = append(w.profiles[name], domain.Profile(row.Cells[1].Value))
	}
	return w.publish(name, "Descripción de "+name)
}

func (w *authorizationWorld) givenUsers(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		if err := w.service.Register(context.Background(), row.Cells[0].Value, domain.UserKind(row.Cells[1].Value)); err != nil {
			return err
		}
	}
	return nil
}

func (w *authorizationWorld) givenPermissions(table *godog.Table) error {
	ctx := context.Background()
	for _, row := range table.Rows[1:] {
		user := row.Cells[0].Value
		if err := w.service.Register(ctx, user, domain.KindStaff); err != nil {
			return err
		}
		p := domain.Permission{UserID: user, ProcedureType: w.typeID(row.Cells[1].Value), Profile: domain.Profile(row.Cells[2].Value)}
		if err := w.service.Grant(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

// givenProcedureOn stores the procedure already on that step, since advancing
// to it is up to the step engine.
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
	return w.givenProcedureOn(name, w.steps[name][0])
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

func (w *authorizationWorld) register(kind string) {
	w.err = w.service.Register(context.Background(), "new@fi.uba.ar", domain.UserKind(kind))
}

func (w *authorizationWorld) grantNew(name, profile string) error {
	if err := w.service.Register(context.Background(), "new-staff@fi.uba.ar", domain.KindStaff); err != nil {
		return err
	}
	w.grant("new-staff@fi.uba.ar", name, profile)
	return nil
}

func (w *authorizationWorld) grant(user, name, profile string) {
	w.err = w.service.Grant(context.Background(), domain.Permission{UserID: user, ProcedureType: w.typeID(name), Profile: domain.Profile(profile)})
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

func (w *authorizationWorld) actOnEach(user string) error {
	for _, id := range w.created {
		p, err := w.procedures.Get(context.Background(), id)
		if err != nil {
			return err
		}
		w.errs = append(w.errs, w.act(user, p.ID, p.OpenStep))
	}
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
			return fmt.Errorf("got %v, want %v", w.err, want)
		}
		return nil
	}
}

func (w *authorizationWorld) thenNotStarted() error {
	if !errors.Is(w.err, domain.ErrAccessDenied) {
		return fmt.Errorf("got %v, want %v", w.err, domain.ErrAccessDenied)
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
		return fmt.Errorf("got %v, want a denial", w.err)
	}
	return nil
}

func (w *authorizationWorld) thenAllAllowed() error {
	if len(w.errs) != 2 {
		return fmt.Errorf("acted %d times, want 2", len(w.errs))
	}
	return errors.Join(w.errs...)
}

func (w *authorizationWorld) thenNotOffered(step string) error {
	for _, a := range w.view.Actions {
		if a == step {
			return fmt.Errorf("offered %q", step)
		}
	}
	return nil
}

// thenNotRevealed opens a procedure that does not exist and compares both
// answers, leaving out the identifier asked for.
func (w *authorizationWorld) thenNotRevealed() error {
	_, missing := w.service.Open(context.Background(), w.user, "P-missing")
	if !errors.Is(missing, domain.ErrProcedureNotFound) {
		return fmt.Errorf("a missing procedure got %v", missing)
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

func (w *authorizationWorld) thenDeniedRecorded(onProcedure string) error {
	all := w.denied.All()
	if len(all) == 0 {
		return errors.New("no denied attempt was recorded")
	}
	got := all[len(all)-1]
	want := domain.DeniedAttempt{UserID: w.user, Action: got.Action, At: w.now}
	if onProcedure != "" {
		want.Procedure = w.current
	}
	if got.Action == "" || got != want {
		return fmt.Errorf("recorded %+v, want %+v with an action", got, want)
	}
	return nil
}

func (w *authorizationWorld) thenNoDenied() error {
	if all := w.denied.All(); len(all) > 0 {
		return fmt.Errorf("recorded %+v", all)
	}
	return nil
}
