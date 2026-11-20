package rest

import (
	"context"
	"errors"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/input"
)

const messageCannotRequestCorrection = "No podés pedir una subsanación sobre este paso del trámite."

var invalidCorrectionRequests = []struct {
	err     error
	message string
}{
	{engine.ErrNoReason, "Falta el motivo."},
	{engine.ErrNothingRequested, "Falta indicar qué datos o documentos se piden."},
	{engine.ErrNotARequirement, "Se pide un dato o documento que el tipo de trámite no tiene."},
	{engine.ErrRepeatedRequirement, "Se pide dos veces el mismo dato o documento."},
	{engine.ErrCorrectionOnApplicant, "El paso abierto es del titular: no hay nada que pedirle sobre él."},
}

func (h *ProcedureHandler) RequestCorrection(ctx context.Context, req api.RequestCorrectionRequestObject) (api.RequestCorrectionResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.RequestCorrection401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return api.RequestCorrection400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: api.InvalidRequestMessage}}, nil
	}
	event, err := h.steps.RequestCorrection(ctx, user.ID, req.ProcedureId, correctionRequest(req.StepId, *req.Body))
	if err != nil {
		return correctionRequestError(err)
	}
	recorded, err := h.named(ctx, event)
	return api.RequestCorrection201JSONResponse(recorded), err
}

func correctionRequest(step string, body api.NewCorrectionRequest) input.CorrectionRequest {
	return input.CorrectionRequest{Step: step, Reason: body.Reason, Fields: valueOf(body.Fields), Documents: valueOf(body.Documents)}
}

func correctionRequestError(err error) (api.RequestCorrectionResponseObject, error) {
	switch {
	case errors.Is(err, domain.ErrProcedureNotFound):
		return api.RequestCorrection404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	case errors.Is(err, domain.ErrAccessDenied):
		return api.RequestCorrection403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageCannotRequestCorrection}}, nil
	case errors.Is(err, engine.ErrStepNotOpen):
		return api.RequestCorrection409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: messageStepNotOpen}}, nil
	}
	for _, invalid := range invalidCorrectionRequests {
		if errors.Is(err, invalid.err) {
			return api.RequestCorrection400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: invalid.message}}, nil
		}
	}
	return nil, err
}
