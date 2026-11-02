package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// InvalidRequestMessage answers a request that does not follow the contract.
const InvalidRequestMessage = "El pedido no es válido."

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
