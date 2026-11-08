package acceptance

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"backend/src/api"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/application"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
)

// inboxWorld is the state of one scenario of US17 and US37.
type inboxWorld struct {
	summaries *memory.ProcedureSummaries
	service   *application.InboxService
	now       time.Time
	types     map[string]int64
	areas     map[string]string

	list   inbox.List
	status int
}

func registerInboxSteps(sc *godog.ScenarioContext) {
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	summaries := memory.NewProcedureSummaries()
	world := &inboxWorld{
		summaries: summaries,
		service:   application.NewInboxService(summaries, func() time.Time { return now }),
		now:       now,
		types:     map[string]int64{},
		areas:     map[string]string{},
	}

	sc.Given(`^these procedures in progress:$`, world.givenProcedures)

	sc.When(`^a staff member with the "([^"]*)" profile opens their inbox$`, func(profile string) error {
		return world.query(definition.Profile(profile), inbox.Query{View: inbox.ViewMine})
	})
	sc.When(`^a staff member with the "([^"]*)" profile lists all the procedures$`, func(profile string) error {
		return world.query(definition.Profile(profile), inbox.Query{View: inbox.ViewAll})
	})
	sc.When(`^a staff member lists all the procedures of the "([^"]*)" type$`, func(name string) error {
		return world.queryAll(inbox.Query{ProcedureType: world.types[name]})
	})
	sc.When(`^a staff member lists all the procedures in the state "([^"]*)"$`, func(state string) error {
		return world.queryAll(inbox.Query{State: inbox.State(state)})
	})
	sc.When(`^a staff member lists all the procedures in the hands of "([^"]*)"$`, func(area string) error {
		return world.queryAll(inbox.Query{Area: world.areas[area]})
	})
	sc.When(`^a staff member lists all the procedures of the "([^"]*)" type in the state "([^"]*)"$`, func(name, state string) error {
		return world.queryAll(inbox.Query{ProcedureType: world.types[name], State: inbox.State(state)})
	})
	sc.When(`^a staff member lists all the procedures sorted by "([^"]*)"$`, func(sort string) error {
		return world.queryAll(inbox.Query{Sort: inbox.Sort(sort)})
	})
	sc.When(`^an applicant asks for the inbox$`, func() { world.request("applicant", nil) })
	sc.When(`^someone without a session asks for the inbox$`, func() { world.request("", rest.ErrNoSession) })

	sc.Then(`^it lists the procedures "([^"]*)"$`, world.thenLists)
	sc.Then(`^"([^"]*)" shows the procedure type "([^"]*)"$`, world.thenType)
	sc.Then(`^"([^"]*)" shows the pending step "([^"]*)" in "([^"]*)"$`, world.thenStep)
	sc.Then(`^"([^"]*)" is (\d+) days old$`, world.thenAge)
	sc.Then(`^"([^"]*)" is marked as actionable$`, func(id string) error { return world.thenActionable(id, true) })
	sc.Then(`^"([^"]*)" is not marked as actionable$`, func(id string) error { return world.thenActionable(id, false) })
	sc.Then(`^"([^"]*)" is in the state "([^"]*)"$`, world.thenState)
	sc.Then(`^"([^"]*)" has been referred for (\d+) days$`, world.thenReferred)
	sc.Then(`^"([^"]*)" shows no referral days$`, world.thenNotReferred)
	sc.Then(`^they are told the inbox is only for the staff$`, func() error { return world.thenStatus(http.StatusForbidden) })
	sc.Then(`^they are asked to sign in$`, func() error { return world.thenStatus(http.StatusUnauthorized) })
	sc.Then(`^the type filter offers "([^"]*)"$`, world.thenTypeOptions)
}

func (w *inboxWorld) givenProcedures(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		cell := func(i int) string { return row.Cells[i].Value }
		typeName, areaName := cell(1), cell(3)
		if _, ok := w.types[typeName]; !ok {
			w.types[typeName] = int64(len(w.types) + 1)
		}
		w.areas[areaName] = strings.ToLower(strings.ReplaceAll(areaName, " ", "_"))
		started, err := strconv.Atoi(cell(6))
		if err != nil {
			return err
		}
		moved, err := strconv.Atoi(cell(7))
		if err != nil {
			return err
		}
		w.summaries.Add(inbox.Summary{
			ID:                cell(0),
			ProcedureType:     w.types[typeName],
			ProcedureTypeName: typeName,
			ApplicantName:     "Titular de prueba",
			StepName:          cell(2),
			Area:              definition.Area{ID: w.areas[areaName], Name: areaName, Kind: definition.AreaKind(cell(4))},
			Profile:           definition.Profile(cell(5)),
			StartedAt:         w.daysAgo(started),
			LastMovedAt:       w.daysAgo(moved),
		})
	}
	return nil
}

func (w *inboxWorld) daysAgo(days int) time.Time {
	return w.now.Add(-time.Duration(days) * 24 * time.Hour)
}

func (w *inboxWorld) query(viewer definition.Profile, q inbox.Query) error {
	list, err := w.service.List(context.Background(), viewer, q)
	w.list = list
	return err
}

// queryAll lists every procedure, as an operative profile does from the
// list, with the filters and sort of the scenario.
func (w *inboxWorld) queryAll(q inbox.Query) error {
	q.View = inbox.ViewAll
	return w.query(definition.ProfileStaffOperative, q)
}

// request asks for the inbox over HTTP, with whatever role the session has.
func (w *inboxWorld) request(role string, authErr error) {
	authenticate := func(context.Context, string) (string, error) { return role, authErr }
	server := api.NewServer(nil, nil, rest.NewInboxHandler(w.service, authenticate))
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/procedures", http.NoBody))
	w.status = recorder.Code
}

func (w *inboxWorld) item(id string) (inbox.Item, error) {
	for _, item := range w.list.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return inbox.Item{}, fmt.Errorf("%q is not listed", id)
}

func (w *inboxWorld) thenLists(ids string) error {
	var got []string
	for _, item := range w.list.Items {
		got = append(got, item.ID)
	}
	want := strings.Split(ids, ", ")
	if !slices.Equal(got, want) {
		return fmt.Errorf("listed %v, want %v", got, want)
	}
	return nil
}

func (w *inboxWorld) thenType(id, name string) error {
	item, err := w.item(id)
	if err != nil {
		return err
	}
	if item.ProcedureTypeName != name {
		return fmt.Errorf("procedure type = %q, want %q", item.ProcedureTypeName, name)
	}
	return nil
}

func (w *inboxWorld) thenStep(id, step, area string) error {
	item, err := w.item(id)
	if err != nil {
		return err
	}
	if item.StepName != step || item.Area.Name != area {
		return fmt.Errorf("pending step = %q in %q, want %q in %q", item.StepName, item.Area.Name, step, area)
	}
	return nil
}

func (w *inboxWorld) thenAge(id string, days int) error {
	item, err := w.item(id)
	if err != nil {
		return err
	}
	if item.AgeDays != days {
		return fmt.Errorf("age = %d days, want %d", item.AgeDays, days)
	}
	return nil
}

func (w *inboxWorld) thenActionable(id string, want bool) error {
	item, err := w.item(id)
	if err != nil {
		return err
	}
	if item.Actionable != want {
		return fmt.Errorf("actionable = %v, want %v", item.Actionable, want)
	}
	return nil
}

func (w *inboxWorld) thenState(id, state string) error {
	item, err := w.item(id)
	if err != nil {
		return err
	}
	if string(item.State) != state {
		return fmt.Errorf("state = %q, want %q", item.State, state)
	}
	return nil
}

func (w *inboxWorld) thenReferred(id string, days int) error {
	item, err := w.item(id)
	if err != nil {
		return err
	}
	if item.DaysReferred == nil || *item.DaysReferred != days {
		return fmt.Errorf("days referred = %v, want %d", item.DaysReferred, days)
	}
	return nil
}

func (w *inboxWorld) thenNotReferred(id string) error {
	item, err := w.item(id)
	if err != nil {
		return err
	}
	if item.DaysReferred != nil {
		return fmt.Errorf("days referred = %d, want none", *item.DaysReferred)
	}
	return nil
}

func (w *inboxWorld) thenStatus(want int) error {
	if w.status != want {
		return fmt.Errorf("status = %d, want %d", w.status, want)
	}
	return nil
}

func (w *inboxWorld) thenTypeOptions(names string) error {
	var got []string
	for _, option := range w.list.ProcedureTypes {
		got = append(got, option.Name)
	}
	want := strings.Split(names, ", ")
	if !slices.Equal(got, want) {
		return fmt.Errorf("type options = %v, want %v", got, want)
	}
	return nil
}
