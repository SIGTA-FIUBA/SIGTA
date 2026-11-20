package rest

import (
	"context"
	"errors"
	"time"

	"backend/src/api"
	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
)

var _ api.AdministrationServer = (*AdministrationHandler)(nil)

// The use cases decide who is an administrator: this handler only finds who
// is behind the session.
type AdministrationHandler struct {
	auth        input.Authenticator
	areas       input.Areas
	memberships input.Memberships
}

func NewAdministrationHandler(auth input.Authenticator, areas input.Areas, memberships input.Memberships) *AdministrationHandler {
	return &AdministrationHandler{auth: auth, areas: areas, memberships: memberships}
}

const (
	messageSignIn            = "Iniciá sesión para continuar."
	messageNotAdmin          = "Solo una persona administradora puede hacer esto."
	messageInvalidArea       = "No se pudo crear el área: revisá que el identificador esté en minúsculas, empiece con una letra y no esté en uso ni sea el de la Secretaría, y que el nombre no esté vacío ni repetido."
	messageInvalidPermission = "No se pudo crear el permiso: revisá que el identificador esté en minúsculas, empiece con una letra y no esté en uso en el área, y que el nombre no esté vacío ni repetido."
	messageInvalidRole       = "No se pudo guardar el rol: revisá que el identificador esté en minúsculas, empiece con una letra y no esté en uso en el área, que el nombre no esté vacío ni repetido, y que tenga al menos un permiso del área."
	messageUnknownArea       = "No existe esa área."
	messageUnknownRole       = "No existe ese rol en esa área."
	messageInvalidGrant      = "No se pudo otorgar la membresía: revisá el área, sus roles y permisos, las fechas y que la persona no tenga ya una membresía vigente en esa área con el mismo alcance."
)

func (h *AdministrationHandler) CreateArea(ctx context.Context, req api.CreateAreaRequestObject) (api.CreateAreaResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.CreateArea401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
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
	return api.CreateArea201JSONResponse{Id: area.ID, Name: area.Name, Permissions: []api.Permission{}, Roles: []api.Role{}}, nil
}

func (h *AdministrationHandler) GrantMembership(ctx context.Context, req api.GrantMembershipRequestObject) (api.GrantMembershipResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.GrantMembership401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	m, err := h.memberships.Grant(ctx, user.ID, input.MembershipGrant{
		UserID:      req.Body.UserId,
		Area:        req.Body.Area,
		Roles:       listOrEmpty(req.Body.Roles),
		Permissions: listOrEmpty(req.Body.Permissions),
		Scope:       valueOrEmpty(req.Body.Scope),
		From:        timeOrZero(req.Body.From),
		Until:       timeOrZero(req.Body.Until),
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
		Id:          m.ID,
		UserId:      m.UserID,
		Area:        m.Area,
		Roles:       nonNil(m.Roles),
		Permissions: nonNil(m.Permissions),
		Scope:       valueOrEmpty(&m.Scope),
		From:        m.From,
		GrantedBy:   m.GrantedBy,
	}
	if !m.Until.IsZero() {
		response.Until = &m.Until
	}
	return response, nil
}

func (h *AdministrationHandler) CreatePermission(ctx context.Context, req api.CreatePermissionRequestObject) (api.CreatePermissionResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.CreatePermission401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	permission, err := h.areas.CreatePermission(ctx, user.ID, req.AreaId, domain.Permission{ID: req.Body.Id, Name: req.Body.Name})
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.CreatePermission403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrAreaNotFound):
		return api.CreatePermission400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageUnknownArea}}, nil
	case errors.Is(err, domain.ErrInvalidPermission):
		return api.CreatePermission400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageInvalidPermission}}, nil
	case err != nil:
		return nil, err
	}
	return api.CreatePermission201JSONResponse{Id: permission.ID, Name: permission.Name}, nil
}

func (h *AdministrationHandler) CreateRole(ctx context.Context, req api.CreateRoleRequestObject) (api.CreateRoleResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.CreateRole401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	role, err := h.areas.CreateRole(ctx, user.ID, req.AreaId, input.NewRole{ID: req.Body.Id, Name: req.Body.Name, Permissions: req.Body.Permissions})
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.CreateRole403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrAreaNotFound):
		return api.CreateRole400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageUnknownArea}}, nil
	case errors.Is(err, domain.ErrInvalidRole):
		return api.CreateRole400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageInvalidRole}}, nil
	case err != nil:
		return nil, err
	}
	return api.CreateRole201JSONResponse(roleOf(role)), nil
}

func (h *AdministrationHandler) UpdateRolePermissions(ctx context.Context, req api.UpdateRolePermissionsRequestObject) (api.UpdateRolePermissionsResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.UpdateRolePermissions401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	role, err := h.areas.UpdateRolePermissions(ctx, user.ID, req.AreaId, req.RoleId, req.Body.Permissions)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.UpdateRolePermissions403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrAreaNotFound), errors.Is(err, domain.ErrRoleNotFound):
		return api.UpdateRolePermissions404JSONResponse{Message: messageUnknownRole}, nil
	case errors.Is(err, domain.ErrInvalidRole):
		return api.UpdateRolePermissions400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageInvalidRole}}, nil
	case err != nil:
		return nil, err
	}
	return api.UpdateRolePermissions200JSONResponse(roleOf(role)), nil
}

func roleOf(r domain.Role) api.Role {
	return api.Role{Id: r.ID, Name: r.Name, Permissions: nonNil(r.Permissions())}
}

func listOrEmpty(ids *[]string) []string {
	if ids == nil {
		return nil
	}
	return *ids
}

func nonNil(ids []string) []string {
	if ids == nil {
		return []string{}
	}
	return ids
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
