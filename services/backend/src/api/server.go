package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// InvalidRequestMessage answers a request that does not follow the contract.
const InvalidRequestMessage = "El pedido no es válido."

// AuthServer is the part of the contract that the auth module answers.
type AuthServer interface {
	BeginLogin(ctx context.Context, request BeginLoginRequestObject) (BeginLoginResponseObject, error)
	CompleteLogin(ctx context.Context, request CompleteLoginRequestObject) (CompleteLoginResponseObject, error)
	Logout(ctx context.Context, request LogoutRequestObject) (LogoutResponseObject, error)
	GetCurrentUser(ctx context.Context, request GetCurrentUserRequestObject) (GetCurrentUserResponseObject, error)
}

// CatalogServer is the part of the contract that the procedure module answers.
type CatalogServer interface {
	ListProcedureTypes(ctx context.Context, request ListProcedureTypesRequestObject) (ListProcedureTypesResponseObject, error)
}

// InboxServer is the part of the contract that answers the staff's inbox, in
// the procedure module.
type InboxServer interface {
	ListProcedures(ctx context.Context, request ListProceduresRequestObject) (ListProceduresResponseObject, error)
}

// Server joins what each module answers into the one interface the contract
// generates. A test that only calls one module can leave the others nil.
type Server struct {
	AuthServer
	CatalogServer
	InboxServer
}

var _ StrictServerInterface = Server{}

func NewServer(auth AuthServer, catalog CatalogServer, inbox InboxServer) Server {
	return Server{AuthServer: auth, CatalogServer: catalog, InboxServer: inbox}
}

// Mount registers the API on mux. Errors raised outside the handlers answer
// with the contract's Error too, without exposing their internal message.
func Mount(mux *http.ServeMux, server StrictServerInterface, logger *slog.Logger) http.Handler {
	badRequest := func(w http.ResponseWriter, _ *http.Request, _ error) {
		writeError(w, http.StatusBadRequest, InvalidRequestMessage)
	}
	internalError := func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		writeError(w, http.StatusInternalServerError, "Ocurrió un error inesperado.")
	}
	strict := NewStrictHandlerWithOptions(server, nil, StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  badRequest,
		ResponseErrorHandlerFunc: internalError,
	})
	return HandlerWithOptions(strict, StdHTTPServerOptions{BaseRouter: mux, ErrorHandlerFunc: badRequest})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The status is already sent, so a failed write has nowhere to be reported.
	_ = json.NewEncoder(w).Encode(Error{Message: message})
}
