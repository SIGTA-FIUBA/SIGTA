package acceptance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
)

// receiptWorld issues the receipt over the procedures and documents of the
// confirmations, so that a scenario can move a procedure before.
type receiptWorld struct {
	*confirmationWorld
	receipt input.Receipt
	readErr error
}

func registerReceiptSteps(sc *godog.ScenarioContext, confirmations *confirmationWorld) {
	w := &receiptWorld{confirmationWorld: confirmations}

	sc.When(`^"([^"]*)" downloads the receipt of the procedure$`, w.download)

	sc.Then(`^the receipt names the procedure by its identifier$`, w.thenIdentifier)
	sc.Then(`^the receipt shows a "([^"]*)" procedure, "([^"]*)"$`, w.thenTypeAndState)
	sc.Then(`^the receipt shows it started on "([^"]*)", last moved on "([^"]*)" and was issued on "([^"]*)"$`, w.thenDates)
	sc.Then(`^the receipt shows the open step "([^"]*)", in "([^"]*)"$`, w.thenOpenStep)
	sc.Then(`^the receipt shows no open step$`, func() error { return w.thenOpenStep("", "") })
	sc.Then(`^the receipt shows the route, in this order:$`, w.thenRoute)
	sc.Then(`^the receipt does not show (.+)$`, w.thenHides)
	sc.Then(`^there is no receipt for them$`, w.thenNoReceipt)
}

func (w *receiptWorld) download(user string) {
	details := application.NewProcedureDetailService(w.service, w.documents)
	w.receipt, w.readErr = details.Receipt(context.Background(), user, w.current)
}

func (w *receiptWorld) thenIdentifier() error {
	if w.readErr != nil {
		return w.readErr
	}
	if w.receipt.Procedure.ID == "" || w.receipt.Procedure.ID != w.current {
		return fmt.Errorf("the receipt names %q, want %q", w.receipt.Procedure.ID, w.current)
	}
	return nil
}

func (w *receiptWorld) thenTypeAndState(procedureType, state string) error {
	if w.readErr != nil {
		return w.readErr
	}
	if idOf(w.receipt.Name) != idOf(procedureType) || string(w.receipt.State) != state {
		return fmt.Errorf("the receipt shows a %q procedure, %q; want %q, %q", w.receipt.Name, w.receipt.State, procedureType, state)
	}
	return nil
}

func (w *receiptWorld) thenDates(started, moved, issued string) error {
	if w.readErr != nil {
		return w.readErr
	}
	p := w.receipt.Procedure
	got := []string{p.StartedAt.Format(time.DateOnly), p.MovedAt.Format(time.DateOnly), w.receipt.IssuedAt.Format(time.DateOnly)}
	if want := []string{started, moved, issued}; fmt.Sprint(got) != fmt.Sprint(want) {
		return fmt.Errorf("started, moved and issued on %v, want %v", got, want)
	}
	return nil
}

func (w *receiptWorld) thenOpenStep(step, area string) error {
	if w.readErr != nil {
		return w.readErr
	}
	if w.receipt.Procedure.OpenStep != step {
		return fmt.Errorf("the open step is %q, want %q", w.receipt.Procedure.OpenStep, step)
	}
	if step != "" && w.areaOf(step) != idOf(area) {
		return fmt.Errorf("the step %q is in %q, want %q", step, w.areaOf(step), area)
	}
	return nil
}

func (w *receiptWorld) areaOf(step string) string {
	for _, s := range w.receipt.Steps {
		if s.Step.ID == step && s.Area != nil {
			return s.Area.ID
		}
	}
	return ""
}

// thenRoute expects, as done by, the person on the steps of the applicant and
// the area on those of staff.
func (w *receiptWorld) thenRoute(table *godog.Table) error {
	if w.readErr != nil {
		return w.readErr
	}
	var got, want []string
	for _, e := range w.receipt.History {
		by := e.By
		if by == "" {
			by = w.areaOf(e.Step)
		}
		got = append(got, fmt.Sprintf("%s %s %s", e.Step, by, e.At.Format(time.DateOnly)))
	}
	for _, row := range tableRows(table) {
		by := row["done by"]
		if !strings.Contains(by, "@") {
			by = idOf(by)
		}
		want = append(want, fmt.Sprintf("%s %s %s", row["step"], by, row["on"]))
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		return fmt.Errorf("route = %v, want %v", got, want)
	}
	return nil
}

// thenHides takes a list like "a", "b" nor "c".
func (w *receiptWorld) thenHides(list string) error {
	if w.readErr != nil {
		return w.readErr
	}
	shown, err := json.Marshal(w.receipt)
	if err != nil {
		return err
	}
	parts := strings.Split(list, `"`)
	for i := 1; i < len(parts); i += 2 {
		if part := parts[i]; strings.Contains(string(shown), part) {
			return fmt.Errorf("the receipt shows %q", part)
		}
	}
	return nil
}

func (w *receiptWorld) thenNoReceipt() error {
	if !errors.Is(w.readErr, domain.ErrProcedureNotFound) {
		return fmt.Errorf("got %w, want %w", w.readErr, domain.ErrProcedureNotFound)
	}
	return nil
}
