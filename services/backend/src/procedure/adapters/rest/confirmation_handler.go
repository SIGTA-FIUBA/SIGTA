package rest

import (
	"context"
	"errors"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	ports "backend/src/procedure/ports/input"
)

const (
	messageCannotConfirm = "No podés confirmar este paso del trámite."
	messageStepNotOpen   = "El paso ya no está abierto: el trámite cambió mientras tanto."
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
	event, err := h.steps.Confirm(ctx, user.ID, req.ProcedureId, ports.StepConfirmation{
		Step:               req.StepId,
		Observation:        req.Body.Observation,
		Outcome:            valueOf(req.Body.Outcome),
		ExternalFileNumber: valueOf(req.Body.ExternalFileNumber),
	})
	if err != nil {
		return confirmationError(err)
	}
	names, err := h.people.Names(ctx, []string{event.By})
	if err != nil {
		return nil, err
	}
	return api.ConfirmStep201JSONResponse(procedureEvent(event, names)), nil
}

func confirmationError(err error) (api.ConfirmStepResponseObject, error) {
	switch {
	case errors.Is(err, domain.ErrProcedureNotFound):
		return api.ConfirmStep404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	case errors.Is(err, domain.ErrAccessDenied):
		return api.ConfirmStep403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageCannotConfirm}}, nil
	case errors.Is(err, engine.ErrStepNotOpen):
		return api.ConfirmStep409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: messageStepNotOpen}}, nil
	}
	for _, invalid := range invalidConfirmations {
		if errors.Is(err, invalid.err) {
			return api.ConfirmStep400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: invalid.message}}, nil
		}
	}
	return nil, err
}

func procedureEvent(e domain.Event, names map[string]string) api.ProcedureEvent {
	event := api.ProcedureEvent{
		Step:        e.Step,
		By:          person(e.By, names),
		At:          e.At,
		Observation: e.Observation,
		Declared:    e.Declared,
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
	return event
}
