package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	ports "backend/src/procedure/ports/input"
)

var _ api.DocumentServer = (*DocumentHandler)(nil)

const (
	messageSignIn       = "Iniciá sesión para continuar."
	messageNoProcedure  = "No existe el trámite."
	messageNoDocument   = "No existe el documento."
	messageCannotAttach = "No podés adjuntar documentos en este paso del trámite."
	messageNotRequested = "El paso abierto del trámite no pide ese documento."
	messageType         = "El archivo no es de un tipo que acepte ese documento."
	messageTooLarge     = "El archivo supera el tamaño máximo de ese documento."
	messageEmpty        = "El archivo está vacío."
)

const (
	formDocument          = "document"
	formFile              = "file"
	maxDocumentIDLength   = 256
	dispositionAttachment = "attachment"
)

var errMalformedUpload = errors.New("the upload must have the document and then the file")

// The use cases decide who sees a procedure and who acts on it: this handler
// only finds who is behind the session.
type DocumentHandler struct {
	auth      authinput.Authenticator
	documents ports.Documents
}

func NewDocumentHandler(auth authinput.Authenticator, documents ports.Documents) *DocumentHandler {
	return &DocumentHandler{auth: auth, documents: documents}
}

func (h *DocumentHandler) ListProcedureDocuments(ctx context.Context, req api.ListProcedureDocumentsRequestObject) (api.ListProcedureDocumentsResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.ListProcedureDocuments401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	attached, err := h.documents.List(ctx, user.ID, req.ProcedureId)
	if errors.Is(err, domain.ErrProcedureNotFound) {
		return api.ListProcedureDocuments404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	}
	if err != nil {
		return nil, err
	}
	current := map[string]bool{}
	for _, d := range domain.Current(attached) {
		current[d.ID] = true
	}
	response := make(api.ListProcedureDocuments200JSONResponse, 0, len(attached))
	for _, d := range attached {
		response = append(response, procedureDocument(d, current[d.ID]))
	}
	return response, nil
}

func (h *DocumentHandler) AttachProcedureDocument(ctx context.Context, req api.AttachProcedureDocumentRequestObject) (api.AttachProcedureDocumentResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.AttachProcedureDocument401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	requirement, file, err := uploadParts(req.Body)
	if err != nil {
		return api.AttachProcedureDocument400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: api.InvalidRequestMessage}}, nil
	}
	defer func() { _ = file.Close() }()
	d, err := h.documents.Attach(ctx, user.ID, ports.Upload{
		ProcedureID: req.ProcedureId,
		Requirement: requirement,
		FileName:    file.FileName(),
		Content:     file,
	})
	if message, ok := rejection(err); ok {
		return api.AttachProcedureDocument400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: message}}, nil
	}
	switch {
	case errors.Is(err, domain.ErrProcedureNotFound):
		return api.AttachProcedureDocument404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	case errors.Is(err, domain.ErrAccessDenied):
		return api.AttachProcedureDocument403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse{Message: messageCannotAttach}}, nil
	case err != nil:
		return nil, err
	}
	return api.AttachProcedureDocument201JSONResponse(procedureDocument(d, true)), nil
}

func (h *DocumentHandler) DownloadProcedureDocument(ctx context.Context, req api.DownloadProcedureDocumentRequestObject) (api.DownloadProcedureDocumentResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.DownloadProcedureDocument401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	d, content, err := h.documents.Open(ctx, user.ID, req.ProcedureId, req.DocumentId)
	switch {
	case errors.Is(err, domain.ErrProcedureNotFound):
		return api.DownloadProcedureDocument404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	case errors.Is(err, domain.ErrDocumentNotFound):
		return api.DownloadProcedureDocument404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoDocument}}, nil
	case err != nil:
		return nil, err
	}
	// The generated response closes the content once it is sent.
	return api.DownloadProcedureDocument200AsteriskResponse{
		Body:          content,
		ContentType:   d.MediaType,
		ContentLength: d.Size,
		Headers: api.DownloadProcedureDocument200ResponseHeaders{
			ContentDisposition:  contentDisposition(d.FileName),
			XContentTypeOptions: "nosniff",
		},
	}, nil
}

// uploadParts reads the document first and leaves the file unread, so that
// the use case streams it to the store and stops at the maximum size.
func uploadParts(body *multipart.Reader) (string, *multipart.Part, error) {
	part, err := body.NextPart()
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", errMalformedUpload, err)
	}
	if part.FormName() != formDocument {
		return "", nil, errMalformedUpload
	}
	value, err := io.ReadAll(io.LimitReader(part, maxDocumentIDLength+1))
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", errMalformedUpload, err)
	}
	if len(value) > maxDocumentIDLength {
		return "", nil, errMalformedUpload
	}
	file, err := body.NextPart()
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", errMalformedUpload, err)
	}
	if file.FormName() != formFile {
		_ = file.Close()
		return "", nil, errMalformedUpload
	}
	return string(value), file, nil
}

func rejection(err error) (string, bool) {
	switch {
	case errors.Is(err, domain.ErrDocumentNotRequested):
		return messageNotRequested, true
	case errors.Is(err, domain.ErrDocumentType):
		return messageType, true
	case errors.Is(err, domain.ErrDocumentTooLarge):
		return messageTooLarge, true
	case errors.Is(err, domain.ErrEmptyDocument):
		return messageEmpty, true
	}
	return "", false
}

// contentDisposition encodes a name that is not ASCII as RFC 6266 asks.
func contentDisposition(fileName string) string {
	if value := mime.FormatMediaType(dispositionAttachment, map[string]string{"filename": fileName}); value != "" {
		return value
	}
	return dispositionAttachment
}

func procedureDocument(d domain.Document, current bool) api.ProcedureDocument {
	return api.ProcedureDocument{
		Id:         d.ID,
		Document:   d.Requirement,
		Step:       d.Step,
		FileName:   d.FileName,
		MediaType:  d.MediaType,
		Size:       d.Size,
		UploadedBy: d.UploadedBy,
		UploadedAt: d.UploadedAt,
		Current:    current,
	}
}

func valueOf(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
