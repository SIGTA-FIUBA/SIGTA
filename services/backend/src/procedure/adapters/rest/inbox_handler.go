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

const (
	signInMessage    = "Iniciá sesión para continuar."
	staffOnlyMessage = "La bandeja es solo para el personal de la Secretaría."
)

func (h *InboxHandler) person(ctx context.Context, session *string) (inbox.Person, error) {
	token := ""
	if session != nil {
		token = *session
	}
	return h.authenticate(ctx, token)
}

func (h *InboxHandler) ListProcedures(ctx context.Context, req api.ListProceduresRequestObject) (api.ListProceduresResponseObject, error) {
	person, err := h.person(ctx, req.Params.SigtaSession)
	if errors.Is(err, ErrNoSession) {
		return api.ListProcedures401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: signInMessage}}, nil
	}
	if err != nil {
		return nil, err
	}
	if !validParams(req.Params) {
		return api.ListProcedures400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: api.InvalidRequestMessage}}, nil
	}
	list, err := h.inbox.List(ctx, person, queryOf(req.Params))
	if errors.Is(err, ports.ErrNotStaff) {
		return api.ListProcedures403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: staffOnlyMessage}}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.ListProcedures200JSONResponse(procedureList(list)), nil
}

func (h *InboxHandler) MarkProcedureSeen(ctx context.Context, req api.MarkProcedureSeenRequestObject) (api.MarkProcedureSeenResponseObject, error) {
	person, err := h.person(ctx, req.Params.SigtaSession)
	if errors.Is(err, ErrNoSession) {
		return api.MarkProcedureSeen401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: signInMessage}}, nil
	}
	if err != nil {
		return nil, err
	}
	err = h.inbox.MarkSeen(ctx, person, req.Id)
	switch {
	case errors.Is(err, ports.ErrNotStaff):
		return api.MarkProcedureSeen403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: staffOnlyMessage}}, nil
	case errors.Is(err, ports.ErrProcedureNotFound):
		return api.MarkProcedureSeen404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "No se encontró el trámite."}}, nil
	case err != nil:
		return nil, err
	}
	return api.MarkProcedureSeen204Response{}, nil
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
		Items:           make([]api.ProcedureListItem, 0, len(list.Items)),
		ProcedureTypes:  make([]api.NamedOption, 0, len(list.ProcedureTypes)),
		Areas:           make([]api.NamedOption, 0, len(list.Areas)),
		ClosedThisMonth: list.ClosedThisMonth,
	}
	for _, item := range list.Items {
		response.Items = append(response.Items, api.ProcedureListItem{
			Id:                item.ID,
			ProcedureType:     item.ProcedureType,
			ProcedureTypeName: item.ProcedureTypeName,
			ApplicantName:     item.ApplicantName,
			ApplicantProgram:  optional(item.ApplicantProgram),
			FileNumber:        optional(item.FileNumber),
			StepName:          item.StepName,
			StepNumber:        item.StepNumber,
			StepCount:         item.StepCount,
			Area:              item.Area.ID,
			AreaName:          item.Area.Name,
			State:             api.ProcedureState(item.State),
			Actionable:        item.Actionable,
			StartedAt:         item.StartedAt,
			LastMovedAt:       item.LastMovedAt,
			AgeDays:           item.AgeDays,
			DaysReferred:      item.DaysReferred,
			HasNews:           item.HasNews,
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

// optional leaves out an empty text, which the contract does not list.
func optional(text string) *string {
	if text == "" {
		return nil
	}
	return &text
}

func idText(id int64) string {
	return strconv.FormatInt(id, 10)
}
