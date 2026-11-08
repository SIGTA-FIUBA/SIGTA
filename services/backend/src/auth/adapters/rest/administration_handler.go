package rest

import (
	"context"
	"errors"
	"time"

	"backend/src/api"
	"backend/src/auth/domain"
	ports "backend/src/auth/ports/input"
)

var _ api.AdministrationServer = (*AdministrationHandler)(nil)

// The use cases decide who is an administrator: this handler only finds who
// is behind the session.
type AdministrationHandler struct {
	auth        ports.Authenticator
	areas       ports.Areas
	memberships ports.Memberships
}

func NewAdministrationHandler(auth ports.Authenticator, areas ports.Areas, memberships ports.Memberships) *AdministrationHandler {
	return &AdministrationHandler{auth: auth, areas: areas, memberships: memberships}
}

const (
	messageSignIn       = "Iniciá sesión para continuar."
	messageNotAdmin     = "Solo una persona administradora puede hacer esto."
	messageInvalidArea  = "No se pudo crear el área: revisá que el identificador esté en minúsculas, empiece con una letra y no esté en uso, y que el nombre no esté vacío ni repetido."
	messageInvalidGrant = "No se pudo otorgar la membresía: revisá el área, el perfil, las fechas y que la persona no tenga ya una membresía vigente en esa área con el mismo alcance."
)

func (h *AdministrationHandler) CreateArea(ctx context.Context, req api.CreateAreaRequestObject) (api.CreateAreaResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if isUnauthenticated(err) {
		return api.CreateArea401JSONResponse{UnauthorizedJSONResponse: unauthorized(messageSignIn)}, nil
	}
	if err != nil {
		return nil, err
	}
	area, err := h.areas.Create(ctx, user.ID, domain.Area{ID: req.Body.Id, Name: req.Body.Name})
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.CreateArea403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrInvalidArea):
		return api.CreateArea400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageInvalidArea}}, nil
	case err != nil:
		return nil, err
	}
	return api.CreateArea201JSONResponse{Id: area.ID, Name: area.Name}, nil
}

func (h *AdministrationHandler) GrantMembership(ctx context.Context, req api.GrantMembershipRequestObject) (api.GrantMembershipResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if isUnauthenticated(err) {
		return api.GrantMembership401JSONResponse{UnauthorizedJSONResponse: unauthorized(messageSignIn)}, nil
	}
	if err != nil {
		return nil, err
	}
	m, err := h.memberships.Grant(ctx, user.ID, ports.MembershipGrant{
		UserID:  req.Body.UserId,
		Area:    req.Body.Area,
		Profile: domain.Profile(req.Body.Profile),
		Scope:   valueOrEmpty(req.Body.Scope),
		From:    timeOrZero(req.Body.From),
		Until:   timeOrZero(req.Body.Until),
	})
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.GrantMembership403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrInvalidMembership):
		return api.GrantMembership400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageInvalidGrant}}, nil
	case err != nil:
		return nil, err
	}
	response := api.GrantMembership201JSONResponse{
		Id:        m.ID,
		UserId:    m.UserID,
		Area:      m.Area,
		Profile:   api.Profile(m.Profile),
		Scope:     valueOrEmpty(&m.Scope),
		From:      m.From,
		GrantedBy: m.GrantedBy,
	}
	if !m.Until.IsZero() {
		response.Until = &m.Until
	}
	return response, nil
}

func valueOrEmpty(scope *[]int64) []int64 {
	if scope == nil || *scope == nil {
		return []int64{}
	}
	return *scope
}

func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
