package rest

import (
	"context"
	"errors"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/ports/input"
)

var _ api.ProcedureServer = (*ProcedureHandler)(nil)

// The use case decides who sees a procedure: this handler only finds who is
// behind the session and the names of the people it shows.
type ProcedureHandler struct {
	auth        authinput.Authenticator
	people      authinput.People
	procedures  input.ProcedureDetails
	submissions input.Submissions
	steps       input.Procedures
}

func NewProcedureHandler(auth authinput.Authenticator, people authinput.People, procedures input.ProcedureDetails, submissions input.Submissions, steps input.Procedures) *ProcedureHandler {
	return &ProcedureHandler{auth: auth, people: people, procedures: procedures, submissions: submissions, steps: steps}
}

func (h *ProcedureHandler) GetProcedure(ctx context.Context, req api.GetProcedureRequestObject) (api.GetProcedureResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.GetProcedure401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	d, err := h.procedures.Detail(ctx, user.ID, req.ProcedureId)
	if errors.Is(err, domain.ErrProcedureNotFound) {
		return api.GetProcedure404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	}
	if err != nil {
		return nil, err
	}
	p := d.Procedure
	people := []string{p.Applicant}
	for _, s := range d.Steps {
		for _, v := range s.Values {
			people = append(people, v.EnteredBy)
		}
		for _, d := range s.Files {
			people = append(people, d.UploadedBy)
		}
	}
	for _, e := range d.History {
		people = append(people, e.By)
	}
	names, err := h.people.Names(ctx, people)
	if err != nil {
		return nil, err
	}
	response := api.GetProcedure200JSONResponse{
		Id:            p.ID,
		ProcedureType: p.ProcedureType,
		Version:       int64(p.Version),
		Name:          d.Name,
		State:         api.ProcedureState(d.State),
		Applicant:     person(p.Applicant, names),
		CanAct:        d.CanAct,
		Steps:         make([]api.ProcedureStep, 0, len(d.Steps)),
		History:       make([]api.ProcedureEvent, 0, len(d.History)),
	}
	if !p.Closed() {
		response.OpenStep = &p.OpenStep
	}
	if r := d.Referral; r != nil {
		response.Referral = &api.ProcedureReferral{
			Area:              api.NamedOption{Id: r.Area.ID, Name: r.Area.Name},
			From:              api.NamedOption{Id: r.From.ID, Name: r.From.Name},
			DaysWithoutReturn: int64(r.DaysWithoutReturn),
		}
	}
	for _, s := range d.Steps {
		response.Steps = append(response.Steps, procedureStep(s, names))
	}
	for _, e := range d.History {
		response.History = append(response.History, procedureEvent(e, names))
	}
	return response, nil
}

func procedureStep(s detail.Step, names map[string]string) api.ProcedureStep {
	step := api.ProcedureStep{
		Id:        s.Step.ID,
		Name:      s.Step.Name,
		Action:    api.ProcedureStepAction(s.Step.Action),
		Role:      api.StepRole(s.Step.Role),
		Fields:    make([]api.FieldRequirement, 0, len(s.Fields)),
		Documents: make([]api.DocumentRequirement, 0, len(s.Documents)),
		Outcomes:  make([]api.StepOutcome, 0, len(s.Step.Outcomes)),
		Values:    make([]api.ProcedureValue, 0, len(s.Values)),
		Files:     make([]api.ProcedureDocument, 0, len(s.Files)),
	}
	if s.Area != nil {
		step.Area = &api.NamedOption{Id: s.Area.ID, Name: s.Area.Name}
	}
	if s.Step.Next != "" {
		step.Next = &s.Step.Next
	}
	for _, f := range s.Fields {
		step.Fields = append(step.Fields, fieldRequirement(f))
	}
	for _, d := range s.Documents {
		step.Documents = append(step.Documents, documentRequirement(d))
	}
	for _, o := range s.Step.Outcomes {
		step.Outcomes = append(step.Outcomes, api.StepOutcome{Id: o.ID, Name: o.Name, Next: o.Next})
	}
	for _, v := range s.Values {
		step.Values = append(step.Values, procedureValue(v, names))
	}
	for _, d := range s.Files {
		step.Files = append(step.Files, procedureDocument(d, true, names))
	}
	return step
}

func procedureValue(v domain.FieldValue, names map[string]string) api.ProcedureValue {
	value := api.ProcedureValue{Field: v.Field, Step: v.Step, EnteredBy: person(v.EnteredBy, names), EnteredAt: v.EnteredAt}
	if v.Items != nil {
		value.Items = &v.Items
	} else {
		value.Text = &v.Text
	}
	return value
}

func (h *ProcedureHandler) ListMyProcedures(ctx context.Context, req api.ListMyProceduresRequestObject) (api.ListMyProceduresResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.ListMyProcedures401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	mine, err := h.procedures.Mine(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	response := make(api.ListMyProcedures200JSONResponse, 0, len(mine))
	for _, m := range mine {
		response = append(response, api.MyProcedure{
			Id:                m.Procedure.ID,
			ProcedureType:     m.Procedure.ProcedureType,
			ProcedureTypeName: m.Name,
			StepName:          m.StepName,
			State:             api.ProcedureState(m.State),
			StartedAt:         m.Procedure.StartedAt,
			LastMovedAt:       m.Procedure.MovedAt,
		})
	}
	return response, nil
}
