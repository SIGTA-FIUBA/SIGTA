package acceptance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"time"

	"github.com/cucumber/godog"

	"backend/src/procedure/adapters/disk"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/submission"
	"backend/src/procedure/ports/input"
)

type historyWorld struct {
	*confirmationWorld
	documents *memory.DocumentRepository
	history   []domain.Event
	before    []domain.Event
	readErr   error
}

func registerHistorySteps(sc *godog.ScenarioContext, base *confirmationWorld) {
	w := &historyWorld{confirmationWorld: base, documents: memory.NewDocumentRepository()}

	sc.Given(`^"([^"]*)" submitted a "([^"]*)" procedure on "([^"]*)"$`, w.givenSubmitted)
	sc.Given(`^"([^"]*)" confirmed the "([^"]*)" step(?: choosing "([^"]*)")? with the observation "([^"]*)"$`, w.givenConfirmedWith)
	sc.Given(`^"([^"]*)" read the history of the procedure$`, w.givenRead)

	sc.When(`^"([^"]*)" uploads "([^"]*)" as the "([^"]*)"$`, w.upload)
	sc.When(`^"([^"]*)" reads the history of the procedure$`, w.read)
	sc.When(`^they change what they read of the history$`, w.changeWhatWasRead)

	sc.Then(`^the history shows, in this order:$`, w.thenShows)
	sc.Then(`^the history starts with what it showed before$`, w.thenStartsAsBefore)
	sc.Then(`^then shows the "([^"]*)" step done by "([^"]*)"$`, w.thenThenShows)
	sc.Then(`^the history shows the "([^"]*)" step$`, w.thenShowsStep)
	sc.Then(`^they cannot read the history$`, w.thenCannotRead)
}

// givenSubmitted fills every field the opening steps ask for.
func (w *historyWorld) givenSubmitted(user, name, date string) error {
	if err := w.itIs(date); err != nil {
		return err
	}
	values := map[string]submission.Value{}
	fields, _ := submission.Requirements(w.definition)
	for _, f := range fields {
		values[f.Field.ID] = submission.Value{Text: "Texto de prueba"}
	}
	clock := func() time.Time { return w.now }
	submissions := application.NewSubmissionService(w.service, memory.NewDocumentRepository(), nil, w.newID, clock)
	started, err := submissions.Submit(context.Background(), user, input.Submission{ProcedureType: w.typeID(name), Values: values})
	if err != nil {
		return err
	}
	w.current = started.Procedure.ID
	w.created = append(w.created, w.current)
	return nil
}

func (w *historyWorld) givenConfirmedWith(user, step, outcome, observation string) error {
	return w.send(user, step, input.StepConfirmation{Observation: observation, Outcome: areaID(outcome)})
}

func (w *historyWorld) givenRead(user string) error {
	w.read(user)
	if w.readErr != nil {
		return w.readErr
	}
	w.before = make([]domain.Event, len(w.history))
	for i, e := range w.history {
		if e.Referral != nil {
			r := *e.Referral
			e.Referral = &r
		}
		w.before[i] = e
	}
	return nil
}

// upload attaches a PDF to the open step. The history only keeps the record,
// so the file is not kept after the upload.
func (w *historyWorld) upload(user, fileName, requirement string) error {
	dir, err := os.MkdirTemp("", "sigta-history-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	files, err := disk.NewFiles(dir)
	if err != nil {
		return err
	}
	documents := application.NewDocumentService(w.service, w.documents, files, w.newID, func() time.Time { return w.now })
	_, err = documents.Attach(context.Background(), user, input.Upload{
		ProcedureID: w.current,
		Requirement: areaID(requirement),
		FileName:    fileName,
		Content:     bytes.NewReader([]byte("%PDF-1.7\n")),
	})
	return err
}

func (w *historyWorld) read(user string) {
	details := application.NewProcedureDetailService(w.service, w.documents)
	d, err := details.Detail(context.Background(), user, w.current)
	w.history, w.readErr = d.History, err
}

func (w *historyWorld) changeWhatWasRead() {
	for i := range w.history {
		w.history[i].By = "someone@fi.uba.ar"
		w.history[i].Observation = "changed"
		w.history[i].At = time.Time{}
		if w.history[i].Referral != nil {
			w.history[i].Referral.To = "elsewhere"
		}
	}
}

func (w *historyWorld) thenShows(table *godog.Table) error {
	if w.readErr != nil {
		return w.readErr
	}
	rows := tableRows(table)
	if len(w.history) != len(rows) {
		return fmt.Errorf("the history has %d steps, want %d: %+v", len(w.history), len(rows), w.history)
	}
	for i, row := range rows {
		e := w.history[i]
		var document, file, fileNumber string
		if e.Document != nil {
			document, file = e.Document.Requirement, e.Document.FileName
		}
		if e.Referral != nil {
			fileNumber = e.Referral.ExternalFileNumber
		}
		got := map[string]string{
			"step":                 e.Step,
			"done by":              e.By,
			"on":                   e.At.Format(time.DateOnly),
			"observation":          e.Observation,
			"outcome":              e.Outcome,
			"action":               string(e.Action),
			"document":             document,
			"file":                 file,
			"external file number": fileNumber,
		}
		for column, want := range row {
			if column == "outcome" || column == "document" {
				want = areaID(want)
			}
			if got[column] != want {
				return fmt.Errorf("entry %d: %s = %q, want %q", i+1, column, got[column], want)
			}
		}
	}
	return nil
}

func (w *historyWorld) thenStartsAsBefore() error {
	if w.readErr != nil {
		return w.readErr
	}
	if len(w.history) < len(w.before) || !reflect.DeepEqual(w.history[:len(w.before)], w.before) {
		return fmt.Errorf("the history is %+v, want it to start with %+v", w.history, w.before)
	}
	return nil
}

func (w *historyWorld) thenThenShows(step, user string) error {
	if len(w.history) != len(w.before)+1 {
		return fmt.Errorf("the history has %d steps, want %d", len(w.history), len(w.before)+1)
	}
	if e := w.history[len(w.before)]; e.Step != step || e.By != user {
		return fmt.Errorf("the last entry is %q by %q, want %q by %q", e.Step, e.By, step, user)
	}
	return nil
}

func (w *historyWorld) thenShowsStep(step string) error {
	if w.readErr != nil {
		return w.readErr
	}
	if !slices.ContainsFunc(w.history, func(e domain.Event) bool { return e.Step == step }) {
		return fmt.Errorf("the history %+v does not show %q", w.history, step)
	}
	return nil
}

func (w *historyWorld) thenCannotRead() error {
	if !errors.Is(w.readErr, domain.ErrProcedureNotFound) {
		return fmt.Errorf("got %w, want %w", w.readErr, domain.ErrProcedureNotFound)
	}
	return nil
}
