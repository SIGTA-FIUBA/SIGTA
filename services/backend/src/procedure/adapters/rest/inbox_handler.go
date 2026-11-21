package rest

import (
	"context"
	"errors"
	"strconv"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/input"
)

var _ api.InboxServer = (*InboxHandler)(nil)

// The use case decides what the inbox shows: this handler only finds who is
// behind the session and the names of the applicants.
type InboxHandler struct {
	auth   authinput.Authenticator
	people authinput.People
	inbox  input.Inbox
}

func NewInboxHandler(auth authinput.Authenticator, people authinput.People, inbox input.Inbox) *InboxHandler {
	return &InboxHandler{auth: auth, people: people, inbox: inbox}
}

func (h *InboxHandler) ListProcedures(ctx context.Context, req api.ListProceduresRequestObject) (api.ListProceduresResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.ListProcedures401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	if !validParams(req.Params) {
		return api.ListProcedures400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: api.InvalidRequestMessage}}, nil
	}
	list, err := h.inbox.List(ctx, user.ID, queryOf(req.Params))
	if errors.Is(err, input.ErrNotStaff) {
		return api.ListProcedures403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: "La bandeja es solo para el personal."}}, nil
	}
	if err != nil {
		return nil, err
	}
	applicants := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		applicants = append(applicants, item.ApplicantID)
	}
	names, err := h.people.Names(ctx, applicants)
	if err != nil {
		return nil, err
	}
	return api.ListProcedures200JSONResponse(procedureList(list, names)), nil
}

// validParams rejects values the contract does not list: the generated code
// does not check enums.
func validParams(p api.ListProceduresParams) bool {
	return (p.View == nil || p.View.Valid()) &&
		(p.State == nil || p.State.Valid()) &&
		(p.Sort == nil || p.Sort.Valid())
}

func queryOf(p api.ListProceduresParams) inbox.Query {
	q := inbox.Query{View: inbox.ViewMine, Sort: inbox.SortOldest}
	if p.View != nil {
		q.View = inbox.View(*p.View)
	}
	if p.ProcedureType != nil {
		q.ProcedureType = *p.ProcedureType
	}
	if p.State != nil {
		q.State = domain.State(*p.State)
	}
	if p.Area != nil {
		q.Area = *p.Area
	}
	if p.Sort != nil {
		q.Sort = inbox.Sort(*p.Sort)
	}
	return q
}

func procedureList(list inbox.List, names map[string]string) api.ProcedureList {
	response := api.ProcedureList{
		Items:          make([]api.ProcedureListItem, 0, len(list.Items)),
		ProcedureTypes: make([]api.NamedOption, 0, len(list.ProcedureTypes)),
		Areas:          make([]api.NamedOption, 0, len(list.Areas)),
	}
	for _, item := range list.Items {
		listItem := api.ProcedureListItem{
			Id:                item.ID,
			ProcedureType:     item.ProcedureType,
			ProcedureTypeName: item.ProcedureTypeName,
			Applicant:         person(item.ApplicantID, names),
			StepName:          item.StepName,
			State:             api.ProcedureState(item.State),
			Actionable:        item.Actionable,
			StartedAt:         item.StartedAt,
			LastMovedAt:       item.LastMovedAt,
			AgeDays:           item.AgeDays,
		}
		if item.Area != nil {
			listItem.Area = &api.NamedOption{Id: item.Area.ID, Name: item.Area.Name}
		}
		response.Items = append(response.Items, listItem)
	}
	for _, option := range list.ProcedureTypes {
		response.ProcedureTypes = append(response.ProcedureTypes, api.NamedOption{Id: strconv.FormatInt(option.ID, 10), Name: option.Name})
	}
	for _, option := range list.Areas {
		response.Areas = append(response.Areas, api.NamedOption{Id: option.ID, Name: option.Name})
	}
	return response
}
