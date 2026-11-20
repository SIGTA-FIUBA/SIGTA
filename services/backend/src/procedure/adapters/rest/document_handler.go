package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/input"
)

var _ api.DocumentServer = (*DocumentHandler)(nil)

const (
	messageSignIn          = "Iniciá sesión para continuar."
	messageNoProcedure     = "No existe el trámite."
	messageNoDocument      = "No existe el documento."
	messageCannotAttach    = "No podés adjuntar documentos en este paso del trámite."
	messageNotRequested    = "El paso abierto del trámite no pide ese documento."
	messageNotInCorrection = "Ninguna subsanación pendiente pide ese documento."
	messageType            = "El archivo no es de un tipo que acepte ese documento."
	messageTooLarge        = "El archivo supera el tamaño máximo de ese documento."
	messageEmpty           = "El archivo está vacío."
	messageTooMany         = "El trámite ya tiene la cantidad máxima de archivos para ese documento."
)

const (
	formDocument          = "document"
	formFile              = "file"
	maxDocumentIDLength   = 256
	dispositionAttachment = "attachment"
	noSniff               = "nosniff"
	sandbox               = "sandbox"
	noStore               = "no-store"
)

var errMalformedUpload = errors.New("the upload must have the document and then the file")

// The use cases decide who sees a procedure and who acts on it: this handler
// only finds who is behind the session.
type DocumentHandler struct {
	auth      authinput.Authenticator
	people    authinput.People
	documents input.Documents
}

func NewDocumentHandler(auth authinput.Authenticator, people authinput.People, documents input.Documents) *DocumentHandler {
	return &DocumentHandler{auth: auth, people: people, documents: documents}
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
	uploaders := make([]string, 0, len(attached))
	for _, d := range attached {
		uploaders = append(uploaders, d.UploadedBy)
	}
	names, err := h.people.Names(ctx, uploaders)
	if err != nil {
		return nil, err
	}
	response := make(api.ListProcedureDocuments200JSONResponse, 0, len(attached))
	for _, d := range attached {
		response = append(response, procedureDocument(d.Document, d.Current, names))
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
		// uploadParts only fails on what the request sent.
		message, _ := rejection(err)
		return api.AttachProcedureDocument400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse{Message: message}}, nil
	}
	d, err := h.documents.Attach(ctx, user.ID, input.Upload{
		ProcedureID: req.ProcedureId,
		Requirement: requirement,
		FileName:    file.FileName(),
		Content:     partReader{file},
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
	return api.AttachProcedureDocument201JSONResponse(procedureDocument(d, true, map[string]string{user.ID: user.Name})), nil
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
			ContentDisposition:    mime.FormatMediaType(dispositionAttachment, map[string]string{"filename": d.FileName}),
			XContentTypeOptions:   noSniff,
			ContentSecurityPolicy: sandbox,
			CacheControl:          noStore,
		},
	}, nil
}

// uploadParts reads the document first and leaves the file unread, so that
// the use case streams it to the store and stops at the maximum size. No part
// is closed: closing one reads the rest of the request.
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
		if errors.Is(err, io.EOF) {
			// io.EOF is not wrapped: a request that ends here is malformed,
			// while one that ends before a document has no more files.
			return "", nil, fmt.Errorf("%w: no file after the document", errMalformedUpload)
		}
		return "", nil, fmt.Errorf("%w: %w", errMalformedUpload, err)
	}
	if file.FormName() != formFile {
		return "", nil, errMalformedUpload
	}
	return string(value), file, nil
}

// partReader marks the errors of reading the request, so that a broken or
// oversized upload is answered as such and not as a failure of the server.
type partReader struct {
	part *multipart.Part
}

func (r partReader) Read(p []byte) (int, error) {
	n, err := r.part.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		return n, fmt.Errorf("%w: %w", errMalformedUpload, err)
	}
	return n, err
}

func rejection(err error) (string, bool) {
	var requestTooLarge *http.MaxBytesError
	switch {
	case errors.Is(err, domain.ErrDocumentTooLarge), errors.As(err, &requestTooLarge):
		return messageTooLarge, true
	case errors.Is(err, errMalformedUpload):
		return api.InvalidRequestMessage, true
	case errors.Is(err, domain.ErrDocumentNotRequested):
		return messageNotRequested, true
	case errors.Is(err, engine.ErrNotRequested):
		return messageNotInCorrection, true
	case errors.Is(err, domain.ErrDocumentTypeNotAccepted):
		return messageType, true
	case errors.Is(err, domain.ErrDocumentEmpty):
		return messageEmpty, true
	case errors.Is(err, domain.ErrTooManyDocuments):
		return messageTooMany, true
	}
	return "", false
}

// procedureDocument shows who uploaded the file by the name in names.
func procedureDocument(d domain.Document, current bool, names map[string]string) api.ProcedureDocument {
	return api.ProcedureDocument{
		Id:         d.ID,
		Document:   d.Requirement,
		Step:       d.Step,
		FileName:   d.FileName,
		MediaType:  d.MediaType,
		Size:       d.Size,
		UploadedBy: personIfAny(d.UploadedBy, names),
		UploadedAt: d.UploadedAt,
		Current:    current,
	}
}

// person leaves the name empty if names does not have it, as for someone who
// is not registered.
func person(id string, names map[string]string) api.Person {
	return api.Person{Id: id, Name: names[id]}
}

// personIfAny is nil when id is empty, as for a member of staff that the
// applicant does not see.
func personIfAny(id string, names map[string]string) *api.Person {
	if id == "" {
		return nil
	}
	p := person(id, names)
	return &p
}

// stringIfAny is nil when s is empty: the contract leaves the key out.
func stringIfAny(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// valueOf reads an optional value of a request as its zero value when absent.
func valueOf[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}
