package acceptance

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/detail"
)

// registerDetailSteps adds to the world of the documents what the detail of
// a procedure shows: the detail reads the same procedure type and files.
func registerDetailSteps(sc *godog.ScenarioContext, w *documentWorld) {
	sc.When(`^"([^"]*)" opens the procedure$`, w.open)

	sc.Then(`^they see these steps:$`, w.thenSteps)
	sc.Then(`^the open step is "([^"]*)", done by (.+)$`, w.thenOpenStep)
	sc.Then(`^it is up to them$`, func() error { return w.thenTurn(true) })
	sc.Then(`^it is not up to them$`, func() error { return w.thenTurn(false) })
	sc.Then(`^the "([^"]*)" shows the file "([^"]*)"$`, w.thenShowsFile)
	sc.Then(`^the "([^"]*)" shows no file$`, func(requirement string) error { return w.thenShowsFile(requirement, "") })
}

func (w *documentWorld) open(user string) {
	w.viewer = user
	w.view, w.err = w.opener.Open(context.Background(), user, w.procedure.ID)
}

func doneBy(s detail.Step) string {
	if s.Step.Role == definition.RoleApplicant {
		return "the applicant"
	}
	return s.AreaName
}

func (w *documentWorld) thenSteps(table *godog.Table) error {
	if w.err != nil {
		return w.err
	}
	steps := w.view.Steps
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

func (w *documentWorld) thenOpenStep(step, by string) error {
	if w.err != nil {
		return w.err
	}
	for _, s := range w.view.Steps {
		if s.Step.ID != w.view.Procedure.OpenStep {
			continue
		}
		if s.Step.ID != step || doneBy(s) != by {
			return fmt.Errorf("the open step is %s, done by %s", s.Step.ID, doneBy(s))
		}
		return nil
	}
	return fmt.Errorf("the open step %s is not among the steps", w.view.Procedure.OpenStep)
}

func (w *documentWorld) thenTurn(want bool) error {
	if w.err != nil {
		return w.err
	}
	if w.view.CanAct() != want {
		return fmt.Errorf("%s can act on the open step: %t, want %t", w.viewer, w.view.CanAct(), want)
	}
	return nil
}

// thenShowsFile looks at the documents as whoever opened the procedure sees
// them, which is what the detail puts next to each requirement.
func (w *documentWorld) thenShowsFile(requirement, name string) error {
	if w.err != nil {
		return w.err
	}
	attached, err := w.service.List(context.Background(), w.viewer, w.procedure.ID)
	if err != nil {
		return err
	}
	var got string
	for _, d := range attached {
		if d.Current && d.Requirement == requirement {
			got = d.FileName
		}
	}
	if got != name {
		return fmt.Errorf("the %s shows %q, want %q", requirement, got, name)
	}
	return nil
}
