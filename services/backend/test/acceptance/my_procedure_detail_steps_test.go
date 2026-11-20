package acceptance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cucumber/godog"

	"backend/src/procedure/adapters/disk"
	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
)

// myProcedureDetailWorld reads the detail over the procedures, documents and
// service of the confirmations, so that a scenario can move a procedure and
// attach its documents before reading it. It keeps the files it attaches, so
// that they can be downloaded.
type myProcedureDetailWorld struct {
	*confirmationWorld
	dir        string
	files      *disk.Files
	detail     input.ProcedureDetail
	downloaded []byte
	readErr    error
}

func registerMyProcedureDetailSteps(sc *godog.ScenarioContext, confirmations *confirmationWorld) *myProcedureDetailWorld {
	w := &myProcedureDetailWorld{confirmationWorld: confirmations}
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		if w.dir != "" {
			return ctx, errors.Join(err, os.RemoveAll(w.dir))
		}
		return ctx, err
	})

	sc.Given(`^"([^"]*)" attached "([^"]*)" as the "([^"]*)"$`, w.givenAttachedFile)
	sc.Given(`^"([^"]*)" confirmed the "([^"]*)" step choosing "([^"]*)" with the observation "([^"]*)" and the external file number "([^"]*)"$`, w.givenConfirmedWithFileNumber)

	sc.When(`^"([^"]*)" looks at the procedure$`, w.look)
	sc.When(`^"([^"]*)" downloads "([^"]*)" from the procedure$`, w.download)

	sc.Then(`^they see the open step "([^"]*)", in "([^"]*)"$`, w.thenOpenStep)
	sc.Then(`^they see the procedure is "([^"]*)"$`, w.thenState)
	sc.Then(`^they see it is referred to "([^"]*)" from "([^"]*)"$`, w.thenReferred)
	sc.Then(`^they see these files:$`, w.thenFiles)
	sc.Then(`^they see no files$`, func() error { return w.thenFiles(nil) })
	sc.Then(`^they get the file$`, w.thenGetFile)
	sc.Then(`^the file is not found$`, w.thenFileNotFound)
	return w
}

func (w *myProcedureDetailWorld) documentService() (*application.DocumentService, error) {
	if w.files == nil {
		dir, err := os.MkdirTemp("", "sigta-my-procedure-")
		if err != nil {
			return nil, err
		}
		w.dir = dir
		if w.files, err = disk.NewFiles(dir); err != nil {
			return nil, err
		}
	}
	return application.NewDocumentService(w.service, w.documents, w.files, w.newID, func() time.Time { return w.now }), nil
}

func (w *myProcedureDetailWorld) givenAttachedFile(user, fileName, requirement string) error {
	documents, err := w.documentService()
	if err != nil {
		return err
	}
	_, err = documents.Attach(context.Background(), user, input.Upload{
		ProcedureID: w.current,
		Requirement: idOf(requirement),
		FileName:    fileName,
		Content:     bytes.NewReader([]byte("%PDF-1.7\n" + fileName)),
	})
	return err
}

func (w *myProcedureDetailWorld) givenConfirmedWithFileNumber(user, step, outcome, observation, fileNumber string) error {
	return w.send(user, step, input.StepConfirmation{Observation: observation, Outcome: idOf(outcome), ExternalFileNumber: fileNumber})
}

func (w *myProcedureDetailWorld) look(user string) {
	details := application.NewProcedureDetailService(w.service, w.documents)
	w.detail, w.readErr = details.Detail(context.Background(), user, w.current)
}

func (w *myProcedureDetailWorld) download(user, fileName string) error {
	documents, err := w.documentService()
	if err != nil {
		return err
	}
	all, err := w.documents.List(context.Background(), w.current)
	if err != nil {
		return err
	}
	id := ""
	for _, d := range all {
		if d.FileName == fileName {
			id = d.ID
		}
	}
	if id == "" {
		return fmt.Errorf("nobody attached %q", fileName)
	}
	_, content, err := documents.Open(context.Background(), user, w.current, id)
	w.downloaded, w.readErr = nil, err
	if err != nil {
		return nil
	}
	defer content.Close()
	w.downloaded, err = io.ReadAll(content)
	return err
}

func (w *myProcedureDetailWorld) thenOpenStep(step, area string) error {
	if w.readErr != nil {
		return w.readErr
	}
	if w.detail.Procedure.OpenStep != step {
		return fmt.Errorf("the open step is %q, want %q", w.detail.Procedure.OpenStep, step)
	}
	for _, s := range w.detail.Steps {
		if s.Step.ID == step {
			if s.Area == nil || s.Area.ID != idOf(area) {
				return fmt.Errorf("the step %q is in %+v, want %q", step, s.Area, area)
			}
			return nil
		}
	}
	return fmt.Errorf("the detail does not list the step %q", step)
}

func (w *myProcedureDetailWorld) thenState(state string) error {
	if w.readErr != nil {
		return w.readErr
	}
	if string(w.detail.State) != state {
		return fmt.Errorf("the procedure is %q, want %q", w.detail.State, state)
	}
	return nil
}

func (w *myProcedureDetailWorld) thenReferred(to, from string) error {
	if w.readErr != nil {
		return w.readErr
	}
	r := w.detail.Referral
	if r == nil || r.Area.ID != idOf(to) || r.From.ID != idOf(from) {
		return fmt.Errorf("the referral is %+v, want to %q from %q", r, to, from)
	}
	return nil
}

func (w *myProcedureDetailWorld) thenFiles(table *godog.Table) error {
	if w.readErr != nil {
		return w.readErr
	}
	var got []string
	for _, s := range w.detail.Steps {
		for _, d := range s.Files {
			got = append(got, fmt.Sprintf("%s %s %s", s.Step.ID, d.Requirement, d.FileName))
		}
	}
	var want []string
	if table != nil {
		for _, row := range tableRows(table) {
			want = append(want, fmt.Sprintf("%s %s %s", row["step"], idOf(row["document"]), row["file"]))
		}
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		return fmt.Errorf("files = %v, want %v", got, want)
	}
	return nil
}

func (w *myProcedureDetailWorld) thenGetFile() error {
	if w.readErr != nil {
		return w.readErr
	}
	if !bytes.HasPrefix(w.downloaded, []byte("%PDF-1.7\n")) {
		return fmt.Errorf("got %q, want the file that was attached", w.downloaded)
	}
	return nil
}

func (w *myProcedureDetailWorld) thenFileNotFound() error {
	if !errors.Is(w.readErr, domain.ErrDocumentNotFound) {
		return fmt.Errorf("got %w, want %w", w.readErr, domain.ErrDocumentNotFound)
	}
	return nil
}
