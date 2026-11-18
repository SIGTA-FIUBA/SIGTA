package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"backend/src/api"
	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/ports/input"
)

// inboxWorld wires the services as the server does, and asks the inbox over
// HTTP, so that the names and the session go through the handler.
type inboxWorld struct {
	now         time.Time
	areas       *authapp.AreaService
	memberships *authapp.MembershipService
	types       *application.ProcedureTypeService
	procedures  movableProcedures
	handler     http.Handler

	known   map[string]bool
	typeIDs map[string]int64
	nextID  int

	status int
	list   api.ProcedureList
}

func registerInboxSteps(sc *godog.ScenarioContext) {
	w := &inboxWorld{}
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return ctx, w.setUp()
	})

	sc.Given(`^these procedure types:$`, w.givenProcedureTypes)
	sc.Given(`^these memberships:$`, w.givenMemberships)
	sc.Given(`^these procedures in progress:$`, w.givenProcedures)
	sc.Given(`^the procedure "([^"]*)" has no open step$`, func(id string) error {
		w.procedures.openStep[id] = ""
		return nil
	})
	sc.Given(`^"([^"]*)" has a procedure "([^"]*)" of "([^"]*)" waiting on "([^"]*)"$`, func(applicant, id, procedureType, step string) error {
		return w.addProcedure(id, procedureType, applicant, step, 0, 0)
	})

	sc.When(`^"([^"]*)" opens their inbox$`, func(user string) error { return w.ask(user, nil) })
	sc.When(`^"([^"]*)" lists all the procedures$`, func(user string) error { return w.ask(user, url.Values{"view": {"all"}}) })
	sc.When(`^"([^"]*)" lists all the procedures of the "([^"]*)" type$`, func(user, name string) error {
		return w.ask(user, url.Values{"view": {"all"}, "procedure_type": {w.typeID(name)}})
	})
	sc.When(`^"([^"]*)" lists all the procedures in the state "([^"]*)"$`, func(user, state string) error {
		return w.ask(user, url.Values{"view": {"all"}, "state": {state}})
	})
	sc.When(`^"([^"]*)" lists all the procedures in the hands of "([^"]*)"$`, func(user, area string) error {
		return w.ask(user, url.Values{"view": {"all"}, "area": {slug(area)}})
	})
	sc.When(`^"([^"]*)" lists all the procedures of the "([^"]*)" type in the state "([^"]*)"$`, func(user, name, state string) error {
		return w.ask(user, url.Values{"view": {"all"}, "procedure_type": {w.typeID(name)}, "state": {state}})
	})
	sc.When(`^"([^"]*)" lists all the procedures sorted by "([^"]*)"$`, func(user, sort string) error {
		return w.ask(user, url.Values{"view": {"all"}, "sort": {sort}})
	})
	sc.When(`^"([^"]*)" asks for the inbox$`, func(user string) error { return w.ask(user, nil) })
	sc.When(`^someone without a session asks for the inbox$`, func() error { return w.ask("", nil) })

	sc.Then(`^it lists the procedures "([^"]*)"$`, w.thenLists)
	sc.Then(`^"([^"]*)" shows the procedure type "([^"]*)"$`, func(id, name string) error {
		return w.check(id, func(item api.ProcedureListItem) bool { return item.ProcedureTypeName == name })
	})
	sc.Then(`^"([^"]*)" shows the applicant "([^"]*)"$`, func(id, applicant string) error {
		return w.check(id, func(item api.ProcedureListItem) bool { return item.Applicant.Name == nameOf(applicant) })
	})
	sc.Then(`^"([^"]*)" shows the pending step "([^"]*)" in "([^"]*)"$`, func(id, step, area string) error {
		return w.check(id, func(item api.ProcedureListItem) bool {
			return item.StepName == step && item.Area != nil && item.Area.Name == area
		})
	})
	sc.Then(`^"([^"]*)" shows the pending step "([^"]*)" with no area$`, func(id, step string) error {
		return w.check(id, func(item api.ProcedureListItem) bool { return item.StepName == step && item.Area == nil })
	})
	sc.Then(`^"([^"]*)" is (\d+) days old$`, func(id string, days int) error {
		return w.check(id, func(item api.ProcedureListItem) bool { return item.AgeDays == days })
	})
	sc.Then(`^"([^"]*)" is marked as actionable$`, func(id string) error {
		return w.check(id, func(item api.ProcedureListItem) bool { return item.Actionable })
	})
	sc.Then(`^"([^"]*)" is not marked as actionable$`, func(id string) error {
		return w.check(id, func(item api.ProcedureListItem) bool { return !item.Actionable })
	})
	sc.Then(`^"([^"]*)" is in the state "([^"]*)"$`, func(id, state string) error {
		return w.check(id, func(item api.ProcedureListItem) bool { return string(item.State) == state })
	})
	sc.Then(`^the type filter offers "([^"]*)"$`, func(names string) error { return w.thenOffers("type", w.list.ProcedureTypes, names) })
	sc.Then(`^the area filter offers "([^"]*)"$`, func(names string) error { return w.thenOffers("area", w.list.Areas, names) })
	sc.Then(`^they are told the inbox is only for the staff$`, func() error { return w.thenStatus(http.StatusForbidden) })
	sc.Then(`^they are asked to sign in$`, func() error { return w.thenStatus(http.StatusUnauthorized) })
}

func (w *inboxWorld) setUp() error {
	*w = inboxWorld{now: time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC), known: map[string]bool{}, typeIDs: map[string]int64{}}
	areas := authmem.NewAreaRepository()
	if err := authapp.EnsureSecretariat(context.Background(), areas); err != nil {
		return err
	}
	administrators := authmem.NewAdministrators(scenarioAdministrator)
	stored := authmem.NewMembershipRepository(
		authdomain.Membership{ID: "M-proposer", UserID: defaultAuthors.ProposedBy, Area: authinput.SecretariatArea, Permissions: []string{authinput.EditProcedureTypes}},
		authdomain.Membership{ID: "M-approver", UserID: defaultAuthors.ApprovedBy, Area: authinput.SecretariatArea, Permissions: []string{authinput.PublishProcedureTypes}},
	)
	clock := func() time.Time { return w.now }
	w.areas = authapp.NewAreaService(areas, administrators, stored, clock)
	w.memberships = authapp.NewMembershipService(areas, administrators, stored, w.newID, clock)
	w.types = application.NewProcedureTypeService(memory.NewProcedureTypeVersionRepository(), w.areas, w.memberships, clock)
	w.procedures = movableProcedures{ProcedureRepository: memory.NewProcedureRepository(), openStep: map[string]string{}}
	procedures := application.NewProcedureService(w.types, w.procedures, memory.NewDocumentRepository(), w.memberships, w.newID, clock)
	inbox := rest.NewInboxHandler(sessionIsUser{}, namesFromEmails{}, application.NewInboxService(procedures))
	w.handler = api.Mount(http.NewServeMux(), api.NewServer(nil, nil, nil, nil, nil, inbox), slog.New(slog.DiscardHandler))
	return nil
}

func (w *inboxWorld) newID() string {
	w.nextID++
	return "I-" + strconv.Itoa(w.nextID)
}

func (w *inboxWorld) typeID(name string) string {
	return strconv.FormatInt(w.typeIDs[name], 10)
}

// givenProcedureTypes publishes one procedure type per name, with its steps
// in the order of the table. Each step is done by "the applicant" or by an
// area with a quoted permission; the step ID is its name.
func (w *inboxWorld) givenProcedureTypes(table *godog.Table) error {
	var names []string
	definitions := map[string]*definition.ProcedureTypeDefinition{}
	for _, row := range table.Rows[1:] {
		name, stepName, doneBy := row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value
		def, ok := definitions[name]
		if !ok {
			def = &definition.ProcedureTypeDefinition{
				Name: name, Description: "Un trámite para las pruebas de la bandeja",
			}
			definitions[name] = def
			names = append(names, name)
		}
		step := definition.Step{ID: slug(stepName), Name: stepName, Action: definition.ActionConfirmExternalFact}
		if doneBy == "the applicant" {
			// Each step of the applicant asks for a field of its own.
			field := definition.Field{ID: step.ID + "_data", Name: stepName, Type: definition.FieldLongText}
			def.Fields = append(def.Fields, field)
			step.Role, step.Action, step.Fields = definition.RoleApplicant, definition.ActionEnterData, []string{field.ID}
		} else {
			area, permission, found := strings.Cut(doneBy, ", ")
			if !found {
				return fmt.Errorf("unknown author %q", doneBy)
			}
			permission = strings.Trim(permission, `"`)
			if err := ensure(w.areas, w.known, slug(area), area, permission); err != nil {
				return err
			}
			step.Role, step.Area, step.Permissions = definition.RoleStaff, slug(area), []string{permission}
		}
		if n := len(def.Steps); n > 0 {
			def.Steps[n-1].Next = step.ID
		}
		def.Steps = append(def.Steps, step)
	}
	for i, name := range names {
		id := int64(i + 1)
		w.typeIDs[name] = id
		if _, err := w.types.Publish(context.Background(), input.Publication{ProcedureType: id, Definition: *definitions[name], Authorship: defaultAuthors}); err != nil {
			return fmt.Errorf("publish %s: %w", name, err)
		}
	}
	return nil
}

// givenMemberships reads the procedure types a membership covers by name; an
// empty cell covers all of them.
func (w *inboxWorld) givenMemberships(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		user, area, permissions := row.Cells[0].Value, row.Cells[1].Value, list(row.Cells[2].Value)
		for _, permission := range permissions {
			if err := ensure(w.areas, w.known, slug(area), area, permission); err != nil {
				return err
			}
		}
		var scope []int64
		for _, name := range list(row.Cells[3].Value) {
			scope = append(scope, w.typeIDs[name])
		}
		if _, err := w.memberships.Grant(context.Background(), scenarioAdministrator, authinput.MembershipGrant{
			UserID: user, Area: slug(area), Permissions: permissions, Scope: scope,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (w *inboxWorld) givenProcedures(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		started, err := strconv.Atoi(row.Cells[4].Value)
		if err != nil {
			return err
		}
		moved, err := strconv.Atoi(row.Cells[5].Value)
		if err != nil {
			return err
		}
		if err := w.addProcedure(row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value, row.Cells[3].Value, started, moved); err != nil {
			return err
		}
	}
	return nil
}

func (w *inboxWorld) addProcedure(id, procedureType, applicant, step string, startedDaysAgo, movedDaysAgo int) error {
	typeID, ok := w.typeIDs[procedureType]
	if !ok {
		return fmt.Errorf("unknown procedure type %q", procedureType)
	}
	published, err := w.types.Version(context.Background(), typeID, 1)
	if err != nil {
		return err
	}
	daysAgo := func(days int) time.Time { return w.now.Add(-time.Duration(days) * 24 * time.Hour) }
	return w.procedures.Add(context.Background(), domain.Procedure{
		ID: id, ProcedureType: typeID, Version: 1, VersionDigest: published.Digest,
		Applicant: applicant, OpenStep: slug(step),
		StartedAt: daysAgo(startedDaysAgo), MovedAt: daysAgo(movedDaysAgo),
	}, nil)
}

// ask uses the user as the session token, as sessionIsUser reads it; an
// empty user sends no session.
func (w *inboxWorld) ask(user string, query url.Values) error {
	target := "/procedures"
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, http.NoBody)
	if user != "" {
		r.Header.Set("Cookie", "sigta_session="+user)
	}
	recorder := httptest.NewRecorder()
	w.handler.ServeHTTP(recorder, r)
	w.status, w.list = recorder.Code, api.ProcedureList{}
	if recorder.Code == http.StatusOK {
		return json.Unmarshal(recorder.Body.Bytes(), &w.list)
	}
	return nil
}

func (w *inboxWorld) thenStatus(want int) error {
	if w.status != want {
		return fmt.Errorf("status = %d, want %d", w.status, want)
	}
	return nil
}

func (w *inboxWorld) thenLists(ids string) error {
	if err := w.thenStatus(http.StatusOK); err != nil {
		return err
	}
	var got []string
	for _, item := range w.list.Items {
		got = append(got, item.Id)
	}
	if want := list(ids); !slices.Equal(got, want) {
		return fmt.Errorf("listed %v, want %v", got, want)
	}
	return nil
}

func (w *inboxWorld) check(id string, ok func(api.ProcedureListItem) bool) error {
	i := slices.IndexFunc(w.list.Items, func(item api.ProcedureListItem) bool { return item.Id == id })
	if i < 0 {
		return fmt.Errorf("%s is not listed", id)
	}
	if !ok(w.list.Items[i]) {
		return fmt.Errorf("%s is %+v", id, w.list.Items[i])
	}
	return nil
}

func (w *inboxWorld) thenOffers(filter string, options []api.NamedOption, names string) error {
	var got []string
	for _, option := range options {
		got = append(got, option.Name)
	}
	if want := list(names); !slices.Equal(got, want) {
		return fmt.Errorf("the %s filter offers %v, want %v", filter, got, want)
	}
	return nil
}

func slug(name string) string { return strings.ToLower(strings.ReplaceAll(name, " ", "_")) }

// sessionIsUser takes the session token as the ID of the user.
type sessionIsUser struct{ authinput.Authenticator }

func (sessionIsUser) Authenticate(_ context.Context, token string) (authdomain.User, error) {
	if token == "" {
		return authdomain.User{}, authdomain.ErrSessionNotFound
	}
	return authdomain.User{ID: token}, nil
}

// namesFromEmails names each user after the part of their email before the
// at sign, so that the scenarios can name the applicant.
type namesFromEmails struct{ authinput.People }

func (namesFromEmails) Names(_ context.Context, ids []string) (map[string]string, error) {
	names := map[string]string{}
	for _, id := range ids {
		names[id] = nameOf(id)
	}
	return names, nil
}

func nameOf(email string) string {
	name, _, _ := strings.Cut(email, "@")
	return name
}
