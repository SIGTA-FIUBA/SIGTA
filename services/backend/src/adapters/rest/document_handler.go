package rest

import (
	"backend/src/domain"
	ports "backend/src/ports/input"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
)

const fileField = "file"

type DocumentHandler struct {
	auth      ports.Authenticator
	documents ports.Documents
}

func NewDocumentHandler(auth ports.Authenticator, documents ports.Documents) *DocumentHandler {
	return &DocumentHandler{auth: auth, documents: documents}
}

func (h *DocumentHandler) UploadDocument(ctx context.Context, req UploadDocumentRequestObject) (UploadDocumentResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if isUnauthenticated(err) {
		return UploadDocument401JSONResponse{unauthorized("Iniciá sesión para continuar.")}, nil
	}
	if err != nil {
		return nil, err
	}

	file, err := filePart(req.Body)
	if errors.Is(err, errNoFile) {
		return UploadDocument400JSONResponse{BadRequestJSONResponse{Message: "Falta el archivo."}}, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	document, err := h.documents.Upload(ctx, user, ports.Upload{
		ProcedureID:   req.ProcedureId,
		RequirementID: valueOf(req.Params.RequirementId),
		Filename:      file.FileName(),
		Content:       file,
	})
	var rejected *domain.FileRejectedError
	switch {
	case errors.Is(err, domain.ErrNoProcedureAccess):
		return UploadDocument403JSONResponse{ForbiddenJSONResponse{Message: noAccessMessage}}, nil
	case errors.Is(err, domain.ErrUnknownRequirement):
		return UploadDocument400JSONResponse{BadRequestJSONResponse{Message: "El trámite no tiene ese requisito."}}, nil
	case errors.Is(err, domain.ErrRequirementRequired):
		return UploadDocument400JSONResponse{BadRequestJSONResponse{Message: "Indicá qué requisito cubre el documento."}}, nil
	case errors.Is(err, domain.ErrEmptyFile):
		return UploadDocument400JSONResponse{BadRequestJSONResponse{Message: "El archivo está vacío."}}, nil
	case errors.As(err, &rejected) && errors.Is(err, domain.ErrFileTooLarge):
		return UploadDocument413JSONResponse{Message: ruleMessage(rejected.Rule)}, nil
	case errors.As(err, &rejected):
		return UploadDocument415JSONResponse{Message: ruleMessage(rejected.Rule)}, nil
	case err != nil:
		return nil, err
	}
	return UploadDocument201JSONResponse(documentBody(document)), nil
}

func (h *DocumentHandler) GetDocumentContent(ctx context.Context, req GetDocumentContentRequestObject) (GetDocumentContentResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if isUnauthenticated(err) {
		return GetDocumentContent401JSONResponse{unauthorized("Iniciá sesión para continuar.")}, nil
	}
	if err != nil {
		return nil, err
	}

	document, content, err := h.documents.Open(ctx, user, req.ProcedureId, req.DocumentId)
	if errors.Is(err, domain.ErrNoProcedureAccess) {
		return GetDocumentContent403JSONResponse{ForbiddenJSONResponse{Message: noAccessMessage}}, nil
	}
	if errors.Is(err, domain.ErrDocumentNotFound) {
		return GetDocumentContent404JSONResponse{NotFoundJSONResponse{Message: "El documento no existe."}}, nil
	}
	if err != nil {
		return nil, err
	}

	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": document.Filename})
	return documentContentResponse{GetDocumentContent200ApplicationoctetStreamResponse{
		Body:          content,
		Headers:       GetDocumentContent200ResponseHeaders{ContentDisposition: &disposition},
		ContentLength: document.Size,
	}}, nil
}

const noAccessMessage = "No tenés acceso a ese trámite."

// documentContentResponse stops browsers from guessing a type for a file a
// person uploaded, which could otherwise run as a page of SIGTA.
type documentContentResponse struct {
	GetDocumentContent200ApplicationoctetStreamResponse
}

func (r documentContentResponse) VisitGetDocumentContentResponse(w http.ResponseWriter) error {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	return r.GetDocumentContent200ApplicationoctetStreamResponse.VisitGetDocumentContentResponse(w)
}

var errNoFile = errors.New("no file part in the request")

// filePart returns the first part named file, which the caller reads as a
// stream instead of buffering the whole request.
func filePart(body *multipart.Reader) (*multipart.Part, error) {
	for {
		part, err := body.NextPart()
		if errors.Is(err, io.EOF) {
			return nil, errNoFile
		}
		if err != nil {
			return nil, err
		}
		if part.FormName() == fileField {
			return part, nil
		}
		_ = part.Close()
	}
}

func documentBody(d domain.Document) Document {
	body := Document{
		Id:          d.ID,
		ProcedureId: d.ProcedureID,
		Filename:    d.Filename,
		ContentType: d.ContentType,
		Size:        d.Size,
		UploadedBy:  d.UploadedBy,
		UploadedAt:  d.UploadedAt,
	}
	if d.RequirementID != "" {
		body.RequirementId = &d.RequirementID
	}
	return body
}

func ruleMessage(rule domain.DocumentRule) string {
	return fmt.Sprintf("El requisito admite archivos %s de hasta %s.",
		strings.Join(rule.AcceptedTypes, ", "), formatSize(rule.MaxSize))
}

func formatSize(bytes int64) string {
	const mb = 1 << 20
	if bytes%mb == 0 {
		return fmt.Sprintf("%d MB", bytes/mb)
	}
	return fmt.Sprintf("%d KB", bytes/1024)
}
