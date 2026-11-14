package rest

import (
	"context"
	"errors"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	ports "backend/src/procedure/ports/input"
)

var _ api.ProcedureServer = (*ProcedureHandler)(nil)

// The use case decides who sees a procedure: this handler only finds who is
// behind the session.
type ProcedureHandler struct {
	auth       authinput.Authenticator
	procedures ports.Procedures
}

func NewProcedureHandler(auth authinput.Authenticator, procedures ports.Procedures) *ProcedureHandler {
	return &ProcedureHandler{auth: auth, procedures: procedures}
}

func (h *ProcedureHandler) GetProcedure(ctx context.Context, req api.GetProcedureRequestObject) (api.GetProcedureResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.GetProcedure401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	view, err := h.procedures.Open(ctx, user.ID, req.ProcedureId)
	if errors.Is(err, domain.ErrProcedureNotFound) {
		return api.GetProcedure404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	}
	if err != nil {
		return nil, err
	}
	p := view.Procedure
	response := api.GetProcedure200JSONResponse{
		Id:            p.ID,
		ProcedureType: p.ProcedureType,
		Version:       p.Version,
		Name:          view.Name,
		OpenStep:      p.OpenStep,
		CanAct:        view.CanAct(),
		Steps:         make([]api.ProcedureStep, 0, len(view.Steps)),
	}
	for _, s := range view.Steps {
		documents := make([]api.DocumentRequirement, 0, len(s.Documents))
		for _, d := range s.Documents {
			documents = append(documents, documentRequirement(d))
		}
		response.Steps = append(response.Steps, api.ProcedureStep{
			Id:        s.Step.ID,
			Name:      s.Step.Name,
			Action:    api.ProcedureStepAction(s.Step.Action),
			Role:      api.StepRole(s.Step.Role),
			AreaName:  s.AreaName,
			Documents: documents,
		})
	}
	return response, nil
}
