package rest

import (
	"context"
	"errors"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"backend/src/api"
	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
)

// The areas, their memberships and what removes them. A removal never deletes
// a row: it ends or marks it, and refuses when someone would keep a
// permission that no longer exists, saying why.

func (h *AdministrationHandler) ListAreas(ctx context.Context, req api.ListAreasRequestObject) (api.ListAreasResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.ListAreas401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	areas, err := h.areas.ListToAdminister(ctx, user.ID)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.ListAreas403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case err != nil:
		return nil, err
	}
	response := make(api.ListAreas200JSONResponse, 0, len(areas))
	for _, a := range areas {
		response = append(response, areaOf(a))
	}
	return response, nil
}

func (h *AdministrationHandler) DeleteArea(ctx context.Context, req api.DeleteAreaRequestObject) (api.DeleteAreaResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.DeleteArea401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	err = h.areas.Delete(ctx, user.ID, req.AreaId)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.DeleteArea403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrAreaNotFound):
		return api.DeleteArea404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageUnknownArea}}, nil
	case errors.Is(err, domain.ErrSystemArea):
		return api.DeleteArea409JSONResponse{InUseJSONResponse: api.InUseJSONResponse{Message: messageSystemArea}}, nil
	case errors.Is(err, domain.ErrAreaHasMemberships):
		return api.DeleteArea409JSONResponse{InUseJSONResponse: api.InUseJSONResponse{Message: messageAreaInUse}}, nil
	case err != nil:
		return nil, err
	}
	return api.DeleteArea204Response{}, nil
}

func (h *AdministrationHandler) DeletePermission(ctx context.Context, req api.DeletePermissionRequestObject) (api.DeletePermissionResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.DeletePermission401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	err = h.areas.DeletePermission(ctx, user.ID, req.AreaId, req.PermissionId)
	inUse := func(message string) (api.DeletePermissionResponseObject, error) {
		return api.DeletePermission409JSONResponse{InUseJSONResponse: api.InUseJSONResponse{Message: message}}, nil
	}
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.DeletePermission403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrAreaNotFound), errors.Is(err, domain.ErrPermissionNotFound):
		return api.DeletePermission404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageUnknownPermission}}, nil
	case errors.Is(err, domain.ErrSystemPermission):
		return inUse(messageSystemPermission)
	case errors.Is(err, domain.ErrPermissionInRole):
		return inUse(messagePermissionInRole)
	case errors.Is(err, domain.ErrPermissionHeld):
		return inUse(messagePermissionHeld)
	case err != nil:
		return nil, err
	}
	return api.DeletePermission204Response{}, nil
}

func (h *AdministrationHandler) DeleteRole(ctx context.Context, req api.DeleteRoleRequestObject) (api.DeleteRoleResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.DeleteRole401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	err = h.areas.DeleteRole(ctx, user.ID, req.AreaId, req.RoleId)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.DeleteRole403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrAreaNotFound), errors.Is(err, domain.ErrRoleNotFound):
		return api.DeleteRole404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageUnknownRole}}, nil
	case errors.Is(err, domain.ErrRoleHeld):
		return api.DeleteRole409JSONResponse{InUseJSONResponse: api.InUseJSONResponse{Message: messageRoleHeld}}, nil
	case err != nil:
		return nil, err
	}
	return api.DeleteRole204Response{}, nil
}

func (h *AdministrationHandler) ListAreaMemberships(ctx context.Context, req api.ListAreaMembershipsRequestObject) (api.ListAreaMembershipsResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.ListAreaMemberships401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	memberships, err := h.memberships.ListByArea(ctx, user.ID, req.AreaId)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.ListAreaMemberships403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrAreaNotFound):
		return api.ListAreaMemberships404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageUnknownArea}}, nil
	case err != nil:
		return nil, err
	}
	ids := make([]string, 0, 2*len(memberships))
	for _, m := range memberships {
		ids = append(ids, m.UserID, m.GrantedBy)
	}
	users, err := h.directory.Users(ctx, ids)
	if err != nil {
		return nil, err
	}
	response := make(api.ListAreaMemberships200JSONResponse, 0, len(memberships))
	for _, m := range memberships {
		response = append(response, api.AreaMembership{
			Id:          m.ID,
			Person:      registeredPerson(m.UserID, users),
			Roles:       nonNil(m.Roles),
			Permissions: nonNil(m.Permissions),
			Scope:       valueOrEmpty(&m.Scope),
			From:        m.From,
			Until:       timeOrNil(m.Until),
			GrantedBy:   api.Person{Id: m.GrantedBy, Name: users[m.GrantedBy].Name},
		})
	}
	return response, nil
}

//nolint:dupl // Ending a membership and an administration answer alike, each with its own response types.
func (h *AdministrationHandler) EndMembership(ctx context.Context, req api.EndMembershipRequestObject) (api.EndMembershipResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.EndMembership401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	err = h.memberships.End(ctx, user.ID, req.MembershipId)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.EndMembership403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrMembershipNotFound):
		return api.EndMembership404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageUnknownMembership}}, nil
	case errors.Is(err, domain.ErrMembershipEnded):
		return api.EndMembership409JSONResponse{InUseJSONResponse: api.InUseJSONResponse{Message: messageMembershipEnded}}, nil
	case err != nil:
		return nil, err
	}
	return api.EndMembership204Response{}, nil
}

func (h *AdministrationHandler) ChangeMembershipPermissions(ctx context.Context, req api.ChangeMembershipPermissionsRequestObject) (api.ChangeMembershipPermissionsResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.ChangeMembershipPermissions401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	m, err := h.memberships.Change(ctx, user.ID, req.MembershipId, req.Body.Roles, req.Body.Permissions)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.ChangeMembershipPermissions403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrMembershipNotFound):
		return api.ChangeMembershipPermissions404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageUnknownMembership}}, nil
	case errors.Is(err, domain.ErrMembershipEnded):
		return api.ChangeMembershipPermissions409JSONResponse{InUseJSONResponse: api.InUseJSONResponse{Message: messageMembershipEnded}}, nil
	case errors.Is(err, domain.ErrMembershipOverlaps):
		return api.ChangeMembershipPermissions409JSONResponse{InUseJSONResponse: api.InUseJSONResponse{Message: messageChangeOverlaps}}, nil
	case errors.Is(err, domain.ErrInvalidMembership):
		return api.ChangeMembershipPermissions400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: invalidMembership(err, messageInvalidChange)}}, nil
	case err != nil:
		return nil, err
	}
	return api.ChangeMembershipPermissions200JSONResponse(membershipOf(m)), nil
}

// registeredPerson shows an id that is no longer a user with an empty name
// and email, which the client shows as the id.
func registeredPerson(id string, users map[string]domain.User) api.RegisteredPerson {
	return api.RegisteredPerson{Id: id, Name: users[id].Name, Email: openapi_types.Email(users[id].Email)}
}
