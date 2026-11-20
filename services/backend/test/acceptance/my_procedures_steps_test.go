package acceptance

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/cucumber/godog"

	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
)

// myProceduresWorld lists the procedures of a person over the
// procedures and the service of the confirmations, so that a scenario can
// move a procedure before listing it.
type myProceduresWorld struct {
	*confirmationWorld
	listed []input.MyProcedure
}

func registerMyProceduresSteps(sc *godog.ScenarioContext, confirmations *confirmationWorld) {
	w := &myProceduresWorld{confirmationWorld: confirmations}

	sc.When(`^"([^"]*)" looks at their procedures$`, w.listMine)
	sc.Then(`^they see these procedures:$`, w.thenSee)
	sc.Then(`^they see no procedures$`, func() error { return w.thenSee(nil) })
	sc.Then(`^"([^"]*)" can open the detail of each one$`, w.thenOpensEach)
}

func (w *myProceduresWorld) details() *application.ProcedureDetailService {
	return application.NewProcedureDetailService(w.service, w.documents)
}

func (w *myProceduresWorld) listMine(user string) error {
	listed, err := w.details().Mine(context.Background(), user)
	w.listed = listed
	return err
}

// thenSee compares without order: the procedures of a scenario move at the
// same moment, and the order by last movement is tested in the use case.
func (w *myProceduresWorld) thenSee(table *godog.Table) error {
	stateNames := map[domain.State]string{
		domain.StateWithApplicant: "with the applicant",
		domain.StateInOffice:      "in an office",
		domain.StateReferred:      "referred",
		domain.StateClosed:        "closed",
	}
	var want []string
	if table != nil {
		for _, row := range tableRows(table) {
			want = append(want, strings.Join([]string{row["procedure type"], row["open step"], strings.TrimSpace(row["state"])}, " | "))
		}
	}
	var got []string
	for _, m := range w.listed {
		got = append(got, strings.Join([]string{m.Name, m.StepName, stateNames[m.State]}, " | "))
	}
	slices.Sort(want)
	slices.Sort(got)
	if !slices.Equal(got, want) {
		return fmt.Errorf("listed %q, want %q", got, want)
	}
	return nil
}

func (w *myProceduresWorld) thenOpensEach(user string) error {
	if len(w.listed) == 0 {
		return fmt.Errorf("nothing was listed")
	}
	for _, m := range w.listed {
		if _, err := w.details().Detail(context.Background(), user, m.Procedure.ID); err != nil {
			return fmt.Errorf("open %s: %w", m.Procedure.ID, err)
		}
	}
	return nil
}
