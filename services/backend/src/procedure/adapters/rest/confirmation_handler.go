package rest

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/input"
)

const (
	messageCannotConfirm = "No podés confirmar este paso del trámite."
	messageStepNotOpen   = "El paso ya no está abierto: el trámite cambió mientras tanto."
	messageAwaiting      = "El paso no se puede confirmar mientras el titular tenga una subsanación pendiente."
)

var invalidConfirmations = []struct {
	err     error
	message string
}{
	{engine.ErrNoObservation, "Falta la observación."},
	{engine.ErrInvalidOutcome, "La salida elegida no es una de las del paso."},
	{engine.ErrNotAReferral, "El número de expediente solo va con una derivación."},
	{engine.ErrApplicantStep, "Los pasos del titular los completa el titular."},
}

func (h *ProcedureHandler) ConfirmStep(ctx context.Context, req api.ConfirmStepRequestObject) (api.ConfirmStepResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.ConfirmStep401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return api.ConfirmStep400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: api.InvalidRequestMessage}}, nil
	}
	event, err := h.steps.Confirm(ctx, user.ID, req.ProcedureId, input.StepConfirmation{
		Step:               req.StepId,
		Observation:        req.Body.Observation,
		Outcome:            valueOf(req.Body.Outcome),
		ExternalFileNumber: valueOf(req.Body.ExternalFileNumber),
	})
	if err != nil {
		return confirmationError(err)
	}
	recorded, err := h.named(ctx, event)
	return api.ConfirmStep201JSONResponse(recorded), err
}

func confirmationError(err error) (api.ConfirmStepResponseObject, error) {
	switch {
	case errors.Is(err, domain.ErrProcedureNotFound):
		return api.ConfirmStep404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	case errors.Is(err, domain.ErrAccessDenied):
		return api.ConfirmStep403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageCannotConfirm}}, nil
	case errors.Is(err, engine.ErrStepNotOpen):
		return api.ConfirmStep409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: messageStepNotOpen}}, nil
	case errors.Is(err, engine.ErrAwaitingApplicant):
		return api.ConfirmStep409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: messageAwaiting}}, nil
	}
	var missing *engine.MissingDocumentsError
	if errors.As(err, &missing) {
		return api.ConfirmStep400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: missingDocumentsMessage(missing)}}, nil
	}
	for _, invalid := range invalidConfirmations {
		if errors.Is(err, invalid.err) {
			return api.ConfirmStep400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: invalid.message}}, nil
		}
	}
	return nil, err
}

// named answers an event that was just recorded with the name of whoever did
// it.
func (h *ProcedureHandler) named(ctx context.Context, e domain.Event) (api.ProcedureEvent, error) {
	names, err := h.people.Names(ctx, []string{e.By})
	if err != nil {
		return api.ProcedureEvent{}, err
	}
	return procedureEvent(e, names), nil
}

func procedureEvent(e domain.Event, names map[string]string) api.ProcedureEvent {
	event := api.ProcedureEvent{
		Step:        e.Step,
		By:          personIfAny(e.By, names),
		At:          e.At,
		Observation: e.Observation,
		Declared:    e.Declared,
		SendBack:    e.SendBack,
		Returned:    e.Returned,
	}
	if e.Outcome != "" {
		event.Outcome = &e.Outcome
	}
	if e.Next != "" {
		event.Next = &e.Next
	}
	if e.Referral != nil {
		event.Referral = &api.Referral{From: e.Referral.From, To: e.Referral.To}
		if e.Referral.ExternalFileNumber != "" {
			event.Referral.ExternalFileNumber = &e.Referral.ExternalFileNumber
		}
	}
	if e.Document != nil {
		event.Document = &api.HistoryDocument{Id: e.Document.ID, Document: e.Document.Requirement, FileName: e.Document.FileName}
	}
	if e.Value != nil {
		event.Value = &api.HistoryValue{Field: e.Value.Field}
	}
	if r := e.CorrectionAnswer; r != nil {
		event.CorrectionAnswer = &api.HistoryCorrectionRequest{Fields: orEmpty(r.Fields), Documents: orEmpty(r.Documents)}
	}
	if r := e.CorrectionRequest; r != nil {
		event.CorrectionRequest = &api.HistoryCorrectionRequest{Fields: orEmpty(r.Fields), Documents: orEmpty(r.Documents)}
	}
	return event
}

// missingDocumentsMessage names each missing document, so that whoever
// confirms knows what to attach.
func missingDocumentsMessage(e *engine.MissingDocumentsError) string {
	names := make([]string, 0, len(e.Documents))
	for _, d := range e.Documents {
		names = append(names, nameOrID(d.Name, d.ID))
	}
	if len(names) == 1 {
		return "Falta adjuntar el documento " + quotedList(names) + "."
	}
	return "Falta adjuntar los documentos " + quotedList(names) + "."
}

// quotedList names each one in quotes, the last one after "y".
func quotedList(names []string) string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = fmt.Sprintf("%q", name)
	}
	if len(quoted) == 1 {
		return quoted[0]
	}
	return strings.Join(quoted[:len(quoted)-1], ", ") + " y " + quoted[len(quoted)-1]
}

// nameOrID is the name of a field or document, or its id if it has none.
func nameOrID(name, id string) string {
	if name == "" {
		return id
	}
	return name
}
