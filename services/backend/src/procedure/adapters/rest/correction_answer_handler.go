package rest

import (
	"context"
	"errors"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/submission"
)

const (
	messageOnlyApplicant    = "Solo el titular responde una subsanación."
	messageNoCorrection     = "El trámite no tiene una subsanación pendiente."
	messageValueNotAsked    = "Ninguna subsanación pendiente pide ese dato."
	messageInvalidValue     = "El valor no es válido para ese dato."
	messageCorrectionsMoved = "Las subsanaciones pendientes cambiaron mientras tanto. Volvé a revisarlas."
)

func (h *ProcedureHandler) EnterProcedureValue(ctx context.Context, req api.EnterProcedureValueRequestObject) (api.EnterProcedureValueResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.EnterProcedureValue401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return api.EnterProcedureValue400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: api.InvalidRequestMessage}}, nil
	}
	v, err := h.steps.EnterValue(ctx, user.ID, req.ProcedureId, req.Body.Field, submission.Value{Text: valueOf(req.Body.Text), Items: valueOf(req.Body.Items)})
	switch {
	case errors.Is(err, domain.ErrProcedureNotFound):
		return api.EnterProcedureValue404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	case errors.Is(err, domain.ErrAccessDenied):
		return api.EnterProcedureValue403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageOnlyApplicant}}, nil
	case errors.Is(err, engine.ErrNoCorrectionPending):
		return api.EnterProcedureValue409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: messageNoCorrection}}, nil
	case errors.Is(err, engine.ErrNotRequested):
		return api.EnterProcedureValue400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageValueNotAsked}}, nil
	case errors.Is(err, submission.ErrInvalidValue):
		return api.EnterProcedureValue400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageInvalidValue}}, nil
	case err != nil:
		return nil, err
	}
	names, err := h.people.Names(ctx, []string{v.EnteredBy})
	if err != nil {
		return nil, err
	}
	return api.EnterProcedureValue201JSONResponse(procedureValue(v, names)), nil
}

func (h *ProcedureHandler) AnswerCorrection(ctx context.Context, req api.AnswerCorrectionRequestObject) (api.AnswerCorrectionResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.AnswerCorrection401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	event, err := h.steps.AnswerCorrection(ctx, user.ID, req.ProcedureId)
	var missing *engine.MissingAnswerError
	switch {
	case errors.Is(err, domain.ErrProcedureNotFound):
		return api.AnswerCorrection404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	case errors.Is(err, domain.ErrAccessDenied):
		return api.AnswerCorrection403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageOnlyApplicant}}, nil
	case errors.Is(err, engine.ErrNoCorrectionPending):
		return api.AnswerCorrection409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: messageNoCorrection}}, nil
	case errors.Is(err, engine.ErrCorrectionsChanged):
		return api.AnswerCorrection409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: messageCorrectionsMoved}}, nil
	case errors.As(err, &missing):
		return api.AnswerCorrection400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: missingAnswerMessage(missing)}}, nil
	case err != nil:
		return nil, err
	}
	recorded, err := h.named(ctx, event)
	return api.AnswerCorrection201JSONResponse(recorded), err
}

// missingAnswerMessage names, by their names, what the applicant still has to
// load again.
func missingAnswerMessage(e *engine.MissingAnswerError) string {
	names := make([]string, 0, len(e.Fields)+len(e.Documents))
	for _, f := range e.Fields {
		names = append(names, nameOrID(f.Name, f.ID))
	}
	for _, d := range e.Documents {
		names = append(names, nameOrID(d.Name, d.ID))
	}
	return "Falta cargar de nuevo " + quotedList(names) + "."
}
