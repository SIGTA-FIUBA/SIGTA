package acceptance

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/ports/input"
)

// detailWorld opens the procedure that the world of the documents sets up,
// with the files attached there. Its errors go to that world, which checks
// that the procedure is not found.
type detailWorld struct {
	*documentWorld
	details *application.ProcedureDetailService

	viewer string
	detail input.ProcedureDetail
}

func registerDetailSteps(sc *godog.ScenarioContext, documents *documentWorld) {
	w := &detailWorld{documentWorld: documents}
	// It runs after the Before of the documents, which builds what it uses.
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		procedures := application.NewProcedureService(w.types, w.procedures, w.documents, w.memberships, w.newID, time.Now)
		w.details = application.NewProcedureDetailService(procedures, w.documents)
		return ctx, nil
	})

	sc.Given(`^the procedure has no open step$`, func() error { return w.givenMovedOn("") })

	sc.When(`^"([^"]*)" opens the procedure$`, w.open)

	sc.Then(`^they see these steps:$`, w.thenSteps)
	sc.Then(`^the open step is "([^"]*)", done by (.+)$`, w.thenOpenStep)
	sc.Then(`^the procedure is (with the applicant|in an office|closed)$`, w.thenState)
	sc.Then(`^it is up to them$`, func() error { return w.thenTurn(true) })
	sc.Then(`^it is not up to them$`, func() error { return w.thenTurn(false) })
	sc.Then(`^the "([^"]*)" shows the file "([^"]*)"$`, w.thenShowsFile)
	sc.Then(`^the "([^"]*)" shows no file$`, func(requirement string) error { return w.thenShowsFile(requirement, "") })
}

func (w *detailWorld) open(user string) {
	w.viewer = user
	w.detail, w.err = w.details.Detail(context.Background(), user, w.procedure.ID)
}

func doneBy(s detail.Step) string {
	if s.Step.Role == definition.RoleApplicant {
		return "the applicant"
	}
	return s.Area.Name
}

func (w *detailWorld) thenSteps(table *godog.Table) error {
	if w.err != nil {
		return w.err
	}
	steps := w.detail.Steps
	if len(steps) != len(table.Rows)-1 {
		return fmt.Errorf("got %d steps, want %d", len(steps), len(table.Rows)-1)
	}
	for i, row := range table.Rows[1:] {
		var documents []string
		for _, d := range steps[i].Documents {
			documents = append(documents, d.Document.ID)
		}
		got := []string{steps[i].Step.ID, doneBy(steps[i]), strings.Join(documents, ", ")}
		want := []string{row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value}
		if strings.Join(got, " | ") != strings.Join(want, " | ") {
			return fmt.Errorf("step %d is %v, want %v", i+1, got, want)
		}
	}
	return nil
}

func (w *detailWorld) thenOpenStep(step, by string) error {
	if w.err != nil {
		return w.err
	}
	for _, s := range w.detail.Steps {
		if s.Step.ID != w.detail.Procedure.OpenStep {
			continue
		}
		if s.Step.ID != step || doneBy(s) != by {
			return fmt.Errorf("the open step is %s, done by %s", s.Step.ID, doneBy(s))
		}
		return nil
	}
	return fmt.Errorf("the open step %s is not among the steps", w.detail.Procedure.OpenStep)
}

var states = map[string]domain.State{
	"with the applicant": domain.StateWithApplicant,
	"in an office":       domain.StateInOffice,
	"closed":             domain.StateClosed,
}

func (w *detailWorld) thenState(state string) error {
	if w.err != nil {
		return w.err
	}
	if w.detail.State != states[state] {
		return fmt.Errorf("the procedure is %s, want %s", w.detail.State, states[state])
	}
	return nil
}

func (w *detailWorld) thenTurn(want bool) error {
	if w.err != nil {
		return w.err
	}
	if w.detail.CanAct != want {
		return fmt.Errorf("%s can act on the open step: %t, want %t", w.viewer, w.detail.CanAct, want)
	}
	return nil
}

// thenShowsFile looks for the file next to the requirement in every step of
// the detail that asks for it.
func (w *detailWorld) thenShowsFile(requirement, name string) error {
	if w.err != nil {
		return w.err
	}
	asked := false
	for _, s := range w.detail.Steps {
		if !slices.ContainsFunc(s.Documents, func(r definition.DocumentRequirement) bool { return r.Document.ID == requirement }) {
			continue
		}
		asked = true
		var got string
		for _, d := range s.Files {
			if d.Requirement == requirement {
				got = d.FileName
			}
		}
		if got != name {
			return fmt.Errorf("the %s of %s shows %q, want %q", requirement, s.Step.ID, got, name)
		}
	}
	if !asked {
		return fmt.Errorf("no step of the detail asks for the %s", requirement)
	}
	return nil
}
