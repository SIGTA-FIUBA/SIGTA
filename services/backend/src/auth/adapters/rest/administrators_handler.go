package rest

import (
	"context"
	"errors"

	"backend/src/api"
	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
)

const (
	messageUnknownPerson        = "Esa persona nunca ingresó a SIGTA: tiene que ingresar una vez con su cuenta de la UBA."
	messageAlreadyAdministrator = "La persona ya es administradora."
	messageUnknownAdministrator = "Esa persona no es administradora."
	messageLastAdministrator    = "Es la única persona administradora: otorgá la administración a otra antes de quitársela."
)

// SearchPeople looks for who signed in, so that an administrator can grant
// them a membership or the administration.
func (h *AdministrationHandler) SearchPeople(ctx context.Context, req api.SearchPeopleRequestObject) (api.SearchPeopleResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.SearchPeople401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	people, err := h.directory.Search(ctx, user.ID, req.Params.Q)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.SearchPeople403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case err != nil:
		return nil, err
	}
	response := make(api.SearchPeople200JSONResponse, 0, len(people))
	for _, p := range people {
		response = append(response, personOf(p))
	}
	return response, nil
}

func (h *AdministrationHandler) ListAdministrators(ctx context.Context, req api.ListAdministratorsRequestObject) (api.ListAdministratorsResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.ListAdministrators401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	administrators, err := h.administrators.List(ctx, user.ID)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.ListAdministrators403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case err != nil:
		return nil, err
	}
	ids := make([]string, 0, 2*len(administrators))
	for _, a := range administrators {
		ids = append(ids, a.UserID, a.GrantedBy)
	}
	users, err := h.directory.Users(ctx, ids)
	if err != nil {
		return nil, err
	}
	response := make(api.ListAdministrators200JSONResponse, 0, len(administrators))
	for _, a := range administrators {
		response = append(response, administratorOf(a, users))
	}
	return response, nil
}

func (h *AdministrationHandler) GrantAdministrator(ctx context.Context, req api.GrantAdministratorRequestObject) (api.GrantAdministratorResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.GrantAdministrator401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	a, err := h.administrators.Grant(ctx, user.ID, req.Body.UserId)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.GrantAdministrator403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrUserNotFound):
		return api.GrantAdministrator400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: messageUnknownPerson}}, nil
	case errors.Is(err, domain.ErrAlreadyAdministrator):
		return api.GrantAdministrator409JSONResponse{InUseJSONResponse: api.InUseJSONResponse{Message: messageAlreadyAdministrator}}, nil
	case err != nil:
		return nil, err
	}
	users, err := h.directory.Users(ctx, []string{a.UserID, a.GrantedBy})
	if err != nil {
		return nil, err
	}
	return api.GrantAdministrator201JSONResponse(administratorOf(a, users)), nil
}

//nolint:dupl // Ending a membership and an administration answer alike, each with its own response types.
func (h *AdministrationHandler) EndAdministrator(ctx context.Context, req api.EndAdministratorRequestObject) (api.EndAdministratorResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if input.Unauthenticated(err) {
		return api.EndAdministrator401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	err = h.administrators.End(ctx, user.ID, req.AdministratorId)
	switch {
	case errors.Is(err, domain.ErrAccessDenied):
		return api.EndAdministrator403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageNotAdmin}}, nil
	case errors.Is(err, domain.ErrAdministratorNotFound):
		return api.EndAdministrator404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageUnknownAdministrator}}, nil
	case errors.Is(err, domain.ErrLastAdministrator):
		return api.EndAdministrator409JSONResponse{InUseJSONResponse: api.InUseJSONResponse{Message: messageLastAdministrator}}, nil
	case err != nil:
		return nil, err
	}
	return api.EndAdministrator204Response{}, nil
}

// administratorOf leaves granted_by out for the initial grant of the system,
// which nobody in SIGTA granted.
func administratorOf(a domain.Administrator, users map[string]domain.User) api.Administrator {
	administrator := api.Administrator{Id: a.ID, Person: grantedPerson(a.UserID, users), From: a.From}
	if a.GrantedBy != "" {
		administrator.GrantedBy = &api.Person{Id: a.GrantedBy, Name: users[a.GrantedBy].Name}
	}
	return administrator
}
