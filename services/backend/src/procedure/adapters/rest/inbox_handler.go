package rest

import (
	"context"
	"errors"
	"strconv"

	"backend/src/api"
	"backend/src/procedure/domain/inbox"
	ports "backend/src/procedure/ports/input"
)

// ErrNoSession is what an Authenticate returns when the request has no valid
// session.
var ErrNoSession = errors.New("no valid session")

// Authenticate tells who has the session, as the auth module knows them. The
// procedure module does not depend on the auth module: the composition root
// joins them.
type Authenticate func(ctx context.Context, sessionToken string) (inbox.Person, error)

type InboxHandler struct {
	inbox        ports.Inbox
	authenticate Authenticate
}

func NewInboxHandler(inbox ports.Inbox, authenticate Authenticate) *InboxHandler {
	return &InboxHandler{inbox: inbox, authenticate: authenticate}
}

func (h *InboxHandler) ListProcedures(ctx context.Context, req api.ListProceduresRequestObject) (api.ListProceduresResponseObject, error) {
	token := ""
	if req.Params.SigtaSession != nil {
		token = *req.Params.SigtaSession
	}
	person, err := h.authenticate(ctx, token)
	if errors.Is(err, ErrNoSession) {
		return api.ListProcedures401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: "Iniciá sesión para continuar."}}, nil
	}
	if err != nil {
		return nil, err
	}
	if !validParams(req.Params) {
		return api.ListProcedures400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: api.InvalidRequestMessage}}, nil
	}
	list, err := h.inbox.List(ctx, person, queryOf(req.Params))
	if errors.Is(err, ports.ErrNotStaff) {
		return api.ListProcedures403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: "La bandeja es solo para el personal de la Secretaría."}}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.ListProcedures200JSONResponse(procedureList(list)), nil
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
		q.State = inbox.State(*p.State)
	}
	if p.Area != nil {
		q.Area = *p.Area
	}
	if p.Sort != nil {
		q.Sort = inbox.Sort(*p.Sort)
	}
	return q
}

func procedureList(list inbox.List) api.ProcedureList {
	response := api.ProcedureList{
		Items:          make([]api.ProcedureListItem, 0, len(list.Items)),
		ProcedureTypes: make([]api.NamedOption, 0, len(list.ProcedureTypes)),
		Areas:          make([]api.NamedOption, 0, len(list.Areas)),
	}
	for _, item := range list.Items {
		response.Items = append(response.Items, api.ProcedureListItem{
			Id:                item.ID,
			ProcedureType:     item.ProcedureType,
			ProcedureTypeName: item.ProcedureTypeName,
			ApplicantName:     item.ApplicantName,
			StepName:          item.StepName,
			Area:              item.Area.ID,
			AreaName:          item.Area.Name,
			State:             api.ProcedureState(item.State),
			Actionable:        item.Actionable,
			StartedAt:         item.StartedAt,
			LastMovedAt:       item.LastMovedAt,
			AgeDays:           item.AgeDays,
			DaysReferred:      item.DaysReferred,
		})
	}
	for _, option := range list.ProcedureTypes {
		response.ProcedureTypes = append(response.ProcedureTypes, api.NamedOption{Id: idText(option.ID), Name: option.Name})
	}
	for _, option := range list.Areas {
		response.Areas = append(response.Areas, api.NamedOption{Id: option.ID, Name: option.Name})
	}
	return response
}

func idText(id int64) string {
	return strconv.FormatInt(id, 10)
}
