package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

const InvalidRequestMessage = "El pedido no es válido."

type AuthServer interface {
	BeginLogin(ctx context.Context, request BeginLoginRequestObject) (BeginLoginResponseObject, error)
	CompleteLogin(ctx context.Context, request CompleteLoginRequestObject) (CompleteLoginResponseObject, error)
	Logout(ctx context.Context, request LogoutRequestObject) (LogoutResponseObject, error)
	GetCurrentUser(ctx context.Context, request GetCurrentUserRequestObject) (GetCurrentUserResponseObject, error)
}

type CatalogServer interface {
	ListProcedureTypes(ctx context.Context, request ListProcedureTypesRequestObject) (ListProcedureTypesResponseObject, error)
}

type AdministrationServer interface {
	CreateArea(ctx context.Context, request CreateAreaRequestObject) (CreateAreaResponseObject, error)
	GrantMembership(ctx context.Context, request GrantMembershipRequestObject) (GrantMembershipResponseObject, error)
	CreatePermission(ctx context.Context, request CreatePermissionRequestObject) (CreatePermissionResponseObject, error)
	CreateRole(ctx context.Context, request CreateRoleRequestObject) (CreateRoleResponseObject, error)
	UpdateRolePermissions(ctx context.Context, request UpdateRolePermissionsRequestObject) (UpdateRolePermissionsResponseObject, error)
}

type DocumentServer interface {
	ListProcedureDocuments(ctx context.Context, request ListProcedureDocumentsRequestObject) (ListProcedureDocumentsResponseObject, error)
	AttachProcedureDocument(ctx context.Context, request AttachProcedureDocumentRequestObject) (AttachProcedureDocumentResponseObject, error)
	DownloadProcedureDocument(ctx context.Context, request DownloadProcedureDocumentRequestObject) (DownloadProcedureDocumentResponseObject, error)
}

type ProcedureServer interface {
	GetProcedure(ctx context.Context, request GetProcedureRequestObject) (GetProcedureResponseObject, error)
	SubmitProcedure(ctx context.Context, request SubmitProcedureRequestObject) (SubmitProcedureResponseObject, error)
	ConfirmStep(ctx context.Context, request ConfirmStepRequestObject) (ConfirmStepResponseObject, error)
	ListMyProcedures(ctx context.Context, request ListMyProceduresRequestObject) (ListMyProceduresResponseObject, error)
}

type InboxServer interface {
	ListProcedures(ctx context.Context, request ListProceduresRequestObject) (ListProceduresResponseObject, error)
}

type ProcedureTypeDraftServer interface {
	CreateProcedureType(ctx context.Context, request CreateProcedureTypeRequestObject) (CreateProcedureTypeResponseObject, error)
	ListProcedureTypeDrafts(ctx context.Context, request ListProcedureTypeDraftsRequestObject) (ListProcedureTypeDraftsResponseObject, error)
	GetProcedureTypeDraft(ctx context.Context, request GetProcedureTypeDraftRequestObject) (GetProcedureTypeDraftResponseObject, error)
	OpenProcedureTypeDraft(ctx context.Context, request OpenProcedureTypeDraftRequestObject) (OpenProcedureTypeDraftResponseObject, error)
	EditProcedureTypeDraft(ctx context.Context, request EditProcedureTypeDraftRequestObject) (EditProcedureTypeDraftResponseObject, error)
	DiscardProcedureTypeDraft(ctx context.Context, request DiscardProcedureTypeDraftRequestObject) (DiscardProcedureTypeDraftResponseObject, error)
	ProposeProcedureTypeDraft(ctx context.Context, request ProposeProcedureTypeDraftRequestObject) (ProposeProcedureTypeDraftResponseObject, error)
	RejectProcedureTypeDraft(ctx context.Context, request RejectProcedureTypeDraftRequestObject) (RejectProcedureTypeDraftResponseObject, error)
	ApproveProcedureTypeDraft(ctx context.Context, request ApproveProcedureTypeDraftRequestObject) (ApproveProcedureTypeDraftResponseObject, error)
}

// A test that only calls one module can leave the others nil.
type Server struct {
	AuthServer
	CatalogServer
	AdministrationServer
	DocumentServer
	ProcedureServer
	InboxServer
	ProcedureTypeDraftServer
}

var _ StrictServerInterface = Server{}

func NewServer(auth AuthServer, catalog CatalogServer, administration AdministrationServer, documents DocumentServer, procedures ProcedureServer, inbox InboxServer, drafts ProcedureTypeDraftServer) Server {
	return Server{AuthServer: auth, CatalogServer: catalog, AdministrationServer: administration, DocumentServer: documents, ProcedureServer: procedures, InboxServer: inbox, ProcedureTypeDraftServer: drafts}
}

// MaxRequestBytes bounds the body of every request, so an upload is cut there
// even if its document allows a larger file.
const MaxRequestBytes = 25 << 20

// ReadTimeout bounds how long the server reads a request, body included, so
// a client that sends an upload a byte at a time does not hold the connection
// forever. Five minutes let MaxRequestBytes through at about 0.7 Mbit/s.
const ReadTimeout = 5 * time.Minute

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
	return http.MaxBytesHandler(HandlerWithOptions(strict, StdHTTPServerOptions{BaseRouter: mux, ErrorHandlerFunc: badRequest}), MaxRequestBytes)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The status is already sent, so a failed write has nowhere to be reported.
	_ = json.NewEncoder(w).Encode(Error{Message: message})
}
