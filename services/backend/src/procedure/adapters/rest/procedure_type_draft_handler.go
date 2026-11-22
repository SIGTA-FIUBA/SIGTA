package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/draft"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ api.ProcedureTypeDraftServer = (*ProcedureTypeDraftHandler)(nil)

// The use cases decide who edits, proposes and approves a procedure type:
// this handler only finds who is behind the session and translates errors.
type ProcedureTypeDraftHandler struct {
	auth   authinput.Authenticator
	people authinput.People
	drafts input.ProcedureTypeDrafts
}

func NewProcedureTypeDraftHandler(auth authinput.Authenticator, people authinput.People, drafts input.ProcedureTypeDrafts) *ProcedureTypeDraftHandler {
	return &ProcedureTypeDraftHandler{auth: auth, people: people, drafts: drafts}
}

const (
	messageCannotEdit     = "No tenés permiso para editar o proponer este tipo de trámite."
	messageCannotApprove  = "No tenés permiso para aprobar o rechazar este tipo de trámite."
	messageCannotSee      = "No tenés permiso para ver este borrador."
	messageSelfApproval   = "La propuesta la tiene que aprobar o rechazar otra persona."
	messageNoDraft        = "El tipo de trámite no tiene un borrador."
	messageNoVersion      = "No existe ese tipo de trámite."
	messageDraftExists    = "El tipo de trámite ya tiene un borrador: editá ese."
	messageStaleDraft     = "El borrador cambió desde que lo abriste. Volvé a cargarlo."
	messageProposed       = "El borrador está propuesto: no se puede cambiar hasta que lo aprueben o lo rechacen."
	messageNotProposed    = "El borrador no está propuesto."
	messageOutdatedBase   = "Se publicó otra versión desde que se abrió el borrador. Descartalo y abrí uno nuevo desde la última."
	messageNoReason       = "Escribí el motivo del rechazo."
	messageInvalidDefault = "La definición tiene problemas: corregilos antes de seguir."
)

// draftFailure is how an error of the use cases is answered.
type draftFailure struct {
	status   int
	message  string
	problems []definition.ValidationProblem
}

// failure translates err, with denied as the message of a missing permission.
// It returns false for an error the contract does not describe.
func failure(err error, denied string) (draftFailure, bool) {
	if invalid, ok := errors.AsType[*definition.ValidationError](err); ok && errors.Is(err, definition.ErrInvalidDefinition) {
		return draftFailure{status: http.StatusUnprocessableEntity, message: messageInvalidDefault, problems: invalid.Problems}, true
	}
	cases := []struct {
		err     error
		status  int
		message string
	}{
		{domain.ErrAccessDenied, http.StatusForbidden, denied},
		{draft.ErrSelfApproval, http.StatusForbidden, messageSelfApproval},
		{output.ErrDraftNotFound, http.StatusNotFound, messageNoDraft},
		{output.ErrVersionNotFound, http.StatusNotFound, messageNoVersion},
		{output.ErrDraftExists, http.StatusConflict, messageDraftExists},
		{output.ErrDraftChanged, http.StatusConflict, messageStaleDraft},
		{draft.ErrDraftProposed, http.StatusConflict, messageProposed},
		{draft.ErrDraftNotProposed, http.StatusConflict, messageNotProposed},
		{input.ErrOutdatedBase, http.StatusConflict, messageOutdatedBase},
		{draft.ErrRejectionReasonMissing, http.StatusBadRequest, messageNoReason},
	}
	for _, c := range cases {
		if errors.Is(err, c.err) {
			return draftFailure{status: c.status, message: c.message}, true
		}
	}
	return draftFailure{}, false
}

func (h *ProcedureTypeDraftHandler) CreateProcedureType(ctx context.Context, req api.CreateProcedureTypeRequestObject) (api.CreateProcedureTypeResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.CreateProcedureType401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	def, err := fromContract(req.Body.Definition)
	if err != nil {
		return nil, err
	}
	view, err := h.drafts.Create(ctx, user.ID, def)
	if err != nil {
		f, ok := failure(err, messageCannotEdit)
		if ok && f.status == http.StatusForbidden {
			return api.CreateProcedureType403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: f.message}}, nil
		}
		return nil, err
	}
	names, err := h.names(ctx, view)
	if err != nil {
		return nil, err
	}
	d, err := draftResponse(view, names)
	if err != nil {
		return nil, err
	}
	return api.CreateProcedureType201JSONResponse(d), nil
}

func (h *ProcedureTypeDraftHandler) ListProcedureTypeDrafts(ctx context.Context, req api.ListProcedureTypeDraftsRequestObject) (api.ListProcedureTypeDraftsResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.ListProcedureTypeDrafts401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	views, err := h.drafts.List(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	names, err := h.names(ctx, views...)
	if err != nil {
		return nil, err
	}
	response := make(api.ListProcedureTypeDrafts200JSONResponse, 0, len(views))
	for _, view := range views {
		d, err := draftResponse(view, names)
		if err != nil {
			return nil, err
		}
		response = append(response, d)
	}
	return response, nil
}

func (h *ProcedureTypeDraftHandler) ListEditableProcedureTypes(ctx context.Context, req api.ListEditableProcedureTypesRequestObject) (api.ListEditableProcedureTypesResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.ListEditableProcedureTypes401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	editable, err := h.drafts.Editable(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	var views []input.DraftView
	for _, t := range editable.ProcedureTypes {
		if t.Draft != nil {
			views = append(views, *t.Draft)
		}
	}
	names, err := h.names(ctx, views...)
	if err != nil {
		return nil, err
	}
	response := api.ListEditableProcedureTypes200JSONResponse{
		CanCreate:      editable.CanCreate,
		ProcedureTypes: make([]api.EditableProcedureType, 0, len(editable.ProcedureTypes)),
		Areas:          make([]api.StepArea, 0, len(editable.Areas)),
	}
	for _, t := range editable.ProcedureTypes {
		item := api.EditableProcedureType{ProcedureTypeId: t.ProcedureTypeID, Name: t.Name, Version: t.Latest, CanOpenDraft: t.CanOpenDraft}
		if t.Latest > 0 {
			item.PublishedAt = &t.PublishedAt
		}
		if t.Draft != nil {
			d, err := draftResponse(*t.Draft, names)
			if err != nil {
				return nil, err
			}
			item.Draft = &d
		}
		response.ProcedureTypes = append(response.ProcedureTypes, item)
	}
	for _, a := range editable.Areas {
		area := api.StepArea{Id: a.ID, Name: a.Name, Permissions: make([]api.StepAreaPermission, 0, len(a.Permissions))}
		for _, p := range a.Permissions {
			area.Permissions = append(area.Permissions, api.StepAreaPermission{Id: p.ID, Name: p.Name})
		}
		response.Areas = append(response.Areas, area)
	}
	return response, nil
}

func (h *ProcedureTypeDraftHandler) GetProcedureTypeDraft(ctx context.Context, req api.GetProcedureTypeDraftRequestObject) (api.GetProcedureTypeDraftResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.GetProcedureTypeDraft401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	view, err := h.drafts.Get(ctx, user.ID, req.ProcedureTypeId)
	if err != nil {
		f, ok := failure(err, messageCannotSee)
		switch {
		case ok && f.status == http.StatusForbidden:
			return api.GetProcedureTypeDraft403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusNotFound:
			return api.GetProcedureTypeDraft404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: f.message}}, nil
		}
		return nil, err
	}
	names, err := h.names(ctx, view)
	if err != nil {
		return nil, err
	}
	d, err := draftResponse(view, names)
	if err != nil {
		return nil, err
	}
	return api.GetProcedureTypeDraft200JSONResponse(d), nil
}

func (h *ProcedureTypeDraftHandler) OpenProcedureTypeDraft(ctx context.Context, req api.OpenProcedureTypeDraftRequestObject) (api.OpenProcedureTypeDraftResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.OpenProcedureTypeDraft401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	view, err := h.drafts.Open(ctx, user.ID, req.ProcedureTypeId)
	if err != nil {
		f, ok := failure(err, messageCannotEdit)
		switch {
		case ok && f.status == http.StatusForbidden:
			return api.OpenProcedureTypeDraft403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusNotFound:
			return api.OpenProcedureTypeDraft404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusConflict:
			return api.OpenProcedureTypeDraft409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: f.message}}, nil
		}
		return nil, err
	}
	names, err := h.names(ctx, view)
	if err != nil {
		return nil, err
	}
	d, err := draftResponse(view, names)
	if err != nil {
		return nil, err
	}
	return api.OpenProcedureTypeDraft201JSONResponse(d), nil
}

func (h *ProcedureTypeDraftHandler) EditProcedureTypeDraft(ctx context.Context, req api.EditProcedureTypeDraftRequestObject) (api.EditProcedureTypeDraftResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.EditProcedureTypeDraft401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	def, err := fromContract(req.Body.Definition)
	if err != nil {
		return nil, err
	}
	view, err := h.drafts.Edit(ctx, user.ID, req.ProcedureTypeId, req.Body.Revision, def)
	if err != nil {
		f, ok := failure(err, messageCannotEdit)
		switch {
		case ok && f.status == http.StatusForbidden:
			return api.EditProcedureTypeDraft403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusNotFound:
			return api.EditProcedureTypeDraft404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusConflict:
			return api.EditProcedureTypeDraft409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: f.message}}, nil
		}
		return nil, err
	}
	names, err := h.names(ctx, view)
	if err != nil {
		return nil, err
	}
	d, err := draftResponse(view, names)
	if err != nil {
		return nil, err
	}
	return api.EditProcedureTypeDraft200JSONResponse(d), nil
}

func (h *ProcedureTypeDraftHandler) DiscardProcedureTypeDraft(ctx context.Context, req api.DiscardProcedureTypeDraftRequestObject) (api.DiscardProcedureTypeDraftResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.DiscardProcedureTypeDraft401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	err = h.drafts.Discard(ctx, user.ID, req.ProcedureTypeId, req.Params.Revision)
	if err != nil {
		f, ok := failure(err, messageCannotEdit)
		switch {
		case ok && f.status == http.StatusForbidden:
			return api.DiscardProcedureTypeDraft403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusNotFound:
			return api.DiscardProcedureTypeDraft404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusConflict:
			return api.DiscardProcedureTypeDraft409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: f.message}}, nil
		}
		return nil, err
	}
	return api.DiscardProcedureTypeDraft204Response{}, nil
}

func (h *ProcedureTypeDraftHandler) ProposeProcedureTypeDraft(ctx context.Context, req api.ProposeProcedureTypeDraftRequestObject) (api.ProposeProcedureTypeDraftResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.ProposeProcedureTypeDraft401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	view, err := h.drafts.Propose(ctx, user.ID, req.ProcedureTypeId, req.Body.Revision)
	if err != nil {
		f, ok := failure(err, messageCannotEdit)
		switch {
		case ok && f.status == http.StatusForbidden:
			return api.ProposeProcedureTypeDraft403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusNotFound:
			return api.ProposeProcedureTypeDraft404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusConflict:
			return api.ProposeProcedureTypeDraft409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusUnprocessableEntity:
			return api.ProposeProcedureTypeDraft422JSONResponse{InvalidDefinitionJSONResponse: invalidDefinition(f)}, nil
		}
		return nil, err
	}
	names, err := h.names(ctx, view)
	if err != nil {
		return nil, err
	}
	d, err := draftResponse(view, names)
	if err != nil {
		return nil, err
	}
	return api.ProposeProcedureTypeDraft200JSONResponse(d), nil
}

func (h *ProcedureTypeDraftHandler) RejectProcedureTypeDraft(ctx context.Context, req api.RejectProcedureTypeDraftRequestObject) (api.RejectProcedureTypeDraftResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.RejectProcedureTypeDraft401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	view, err := h.drafts.Reject(ctx, user.ID, req.ProcedureTypeId, req.Body.Revision, req.Body.Reason)
	if err != nil {
		f, ok := failure(err, messageCannotApprove)
		switch {
		case ok && f.status == http.StatusBadRequest:
			return api.RejectProcedureTypeDraft400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusForbidden:
			return api.RejectProcedureTypeDraft403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusNotFound:
			return api.RejectProcedureTypeDraft404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusConflict:
			return api.RejectProcedureTypeDraft409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: f.message}}, nil
		}
		return nil, err
	}
	names, err := h.names(ctx, view)
	if err != nil {
		return nil, err
	}
	d, err := draftResponse(view, names)
	if err != nil {
		return nil, err
	}
	return api.RejectProcedureTypeDraft200JSONResponse(d), nil
}

func (h *ProcedureTypeDraftHandler) ApproveProcedureTypeDraft(ctx context.Context, req api.ApproveProcedureTypeDraftRequestObject) (api.ApproveProcedureTypeDraftResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.ApproveProcedureTypeDraft401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	v, err := h.drafts.Approve(ctx, user.ID, req.ProcedureTypeId, req.Body.Revision)
	if err != nil {
		f, ok := failure(err, messageCannotApprove)
		switch {
		case ok && f.status == http.StatusForbidden:
			return api.ApproveProcedureTypeDraft403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusNotFound:
			return api.ApproveProcedureTypeDraft404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusConflict:
			return api.ApproveProcedureTypeDraft409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: f.message}}, nil
		case ok && f.status == http.StatusUnprocessableEntity:
			return api.ApproveProcedureTypeDraft422JSONResponse{InvalidDefinitionJSONResponse: invalidDefinition(f)}, nil
		}
		return nil, err
	}
	return api.ApproveProcedureTypeDraft200JSONResponse(publishedVersion(v)), nil
}

// names finds the name of whoever edited, proposed or rejected the drafts, so
// that a screen shows people and not their identifiers.
func (h *ProcedureTypeDraftHandler) names(ctx context.Context, views ...input.DraftView) (map[string]string, error) {
	var ids []string
	for _, view := range views {
		d := view.Draft
		ids = append(ids, d.EditedBy, d.ProposedBy, d.Rejection.By)
		ids = append(ids, d.Editors...)
	}
	return h.people.Names(ctx, slices.DeleteFunc(ids, func(id string) bool { return id == "" }))
}

func draftResponse(view input.DraftView, names map[string]string) (api.ProcedureTypeDraft, error) {
	d := view.Draft
	def, err := toContract(d.Definition)
	if err != nil {
		return api.ProcedureTypeDraft{}, err
	}
	response := api.ProcedureTypeDraft{
		ProcedureTypeId: d.ProcedureTypeID,
		Base:            d.Base,
		Revision:        d.Revision,
		Status:          api.DraftStatus(d.Status),
		Definition:      def,
		EditedBy:        person(d.EditedBy, names),
		EditedAt:        d.EditedAt,
		Editors:         people(d.Editors, names),
		Problems:        problems(view.Problems),
		CanEdit:         view.CanEdit,
		CanReview:       view.CanReview,
	}
	if d.ProposedBy != "" {
		proposer := person(d.ProposedBy, names)
		response.ProposedBy = &proposer
		response.ProposedAt = &d.ProposedAt
	}
	if d.Rejection.By != "" {
		response.Rejection = &api.DraftRejection{By: person(d.Rejection.By, names), At: d.Rejection.At, Reason: d.Rejection.Reason}
	}
	return response, nil
}

func publishedVersion(v version.ProcedureTypeVersion) api.PublishedVersion {
	return api.PublishedVersion{
		ProcedureTypeId: v.ProcedureTypeID,
		Number:          v.Number,
		Digest:          v.Digest,
		ProposedBy:      v.Authorship.ProposedBy,
		ApprovedBy:      v.Authorship.ApprovedBy,
		PublishedAt:     v.PublishedAt,
	}
}

func invalidDefinition(f draftFailure) api.InvalidDefinitionJSONResponse {
	return api.InvalidDefinitionJSONResponse{Message: f.message, Problems: problems(f.problems)}
}

func problems(found []definition.ValidationProblem) []api.DefinitionProblem {
	response := make([]api.DefinitionProblem, 0, len(found))
	for _, p := range found {
		path := make([]api.ProblemElement, 0, len(p.Path))
		for _, e := range p.Path {
			element := api.ProblemElement{Kind: string(e.Kind)}
			if e.ID != "" {
				element.Id = &e.ID
			}
			if e.Position > 0 {
				element.Position = &e.Position
			}
			path = append(path, element)
		}
		problem := api.DefinitionProblem{Code: string(p.Code), Path: path, Text: p.String()}
		if p.Value != "" {
			problem.Value = &p.Value
		}
		if p.Other != "" {
			problem.Other = &p.Other
		}
		response = append(response, problem)
	}
	return response
}

// The contract mirrors the JSON of a definition, so both convert through it.
func fromContract(def api.ProcedureTypeDefinition) (definition.ProcedureTypeDefinition, error) {
	var converted definition.ProcedureTypeDefinition
	err := convert(def, &converted)
	return converted, err
}

func toContract(def definition.ProcedureTypeDefinition) (api.ProcedureTypeDefinition, error) {
	var converted api.ProcedureTypeDefinition
	err := convert(def, &converted)
	return converted, err
}

func convert(from, to any) error {
	encoded, err := json.Marshal(from)
	if err != nil {
		return fmt.Errorf("encode definition: %w", err)
	}
	if err := json.Unmarshal(encoded, to); err != nil {
		return fmt.Errorf("decode definition: %w", err)
	}
	return nil
}

func people(ids []string, names map[string]string) []api.Person {
	list := make([]api.Person, 0, len(ids))
	for _, id := range ids {
		list = append(list, person(id, names))
	}
	return list
}
