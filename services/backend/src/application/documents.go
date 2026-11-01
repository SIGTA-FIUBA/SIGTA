package application

import (
	"backend/src/domain"
	"backend/src/ports/input"
	out "backend/src/ports/output"
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// sniffLen is how much of a file content type detection looks at.
const sniffLen = 512

const maxFilenameLen = 255

type DocumentConfig struct {
	// UnlinkedRule applies to documents that cover no requirement, which only
	// the Secretaría can attach.
	UnlinkedRule domain.DocumentRule
}

type DocumentService struct {
	config       DocumentConfig
	access       out.ProcedureAccess
	requirements out.DocumentRequirements
	documents    out.DocumentRepository
	store        out.DocumentStore
	now          func() time.Time
}

func NewDocumentService(
	config DocumentConfig,
	access out.ProcedureAccess,
	requirements out.DocumentRequirements,
	documents out.DocumentRepository,
	store out.DocumentStore,
	now func() time.Time,
) *DocumentService {
	return &DocumentService{
		config:       config,
		access:       access,
		requirements: requirements,
		documents:    documents,
		store:        store,
		now:          now,
	}
}

// Upload stores a document after checking it against the rule of the
// requirement it covers. The type comes from the content: the name and the
// type the client declares are untrusted.
func (s *DocumentService) Upload(ctx context.Context, user domain.User, upload input.Upload) (domain.Document, error) {
	if err := s.authorize(ctx, user, upload.ProcedureID); err != nil {
		return domain.Document{}, err
	}
	rule, err := s.rule(ctx, user, upload)
	if err != nil {
		return domain.Document{}, err
	}

	head := make([]byte, sniffLen)
	n, err := io.ReadFull(upload.Content, head)
	if errors.Is(err, io.EOF) {
		return domain.Document{}, domain.ErrEmptyFile
	}
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return domain.Document{}, err
	}
	head = head[:n]
	contentType := detectContentType(head)
	if !rule.Accepts(contentType) {
		return domain.Document{}, &domain.FileRejectedError{Reason: domain.ErrFileTypeNotAccepted, Rule: rule}
	}

	id := rand.Text()
	// One byte past the limit is enough to know the file is too large without
	// reading the rest of it.
	content := io.LimitReader(io.MultiReader(bytes.NewReader(head), upload.Content), rule.MaxSize+1)
	size, err := s.store.Put(ctx, id, content)
	if err != nil {
		return domain.Document{}, errors.Join(err, s.store.Delete(ctx, id))
	}
	if size > rule.MaxSize {
		return domain.Document{}, errors.Join(
			&domain.FileRejectedError{Reason: domain.ErrFileTooLarge, Rule: rule},
			s.store.Delete(ctx, id),
		)
	}

	document := domain.Document{
		ID:            id,
		ProcedureID:   upload.ProcedureID,
		RequirementID: upload.RequirementID,
		Filename:      cleanFilename(upload.Filename),
		ContentType:   contentType,
		Size:          size,
		UploadedBy:    user.ID,
		UploadedAt:    s.now(),
	}
	if err := s.documents.Save(ctx, document); err != nil {
		return domain.Document{}, errors.Join(err, s.store.Delete(ctx, id))
	}
	return document, nil
}

// Open checks the permission on the procedure on every access. A document is
// only reachable through the procedure it belongs to.
func (s *DocumentService) Open(ctx context.Context, user domain.User, procedureID, documentID string) (domain.Document, io.ReadCloser, error) {
	if err := s.authorize(ctx, user, procedureID); err != nil {
		return domain.Document{}, nil, err
	}
	document, err := s.documents.FindByID(ctx, documentID)
	if err != nil {
		return domain.Document{}, nil, err
	}
	if document.ProcedureID != procedureID {
		return domain.Document{}, nil, domain.ErrDocumentNotFound
	}
	content, err := s.store.Open(ctx, document.ID)
	if err != nil {
		return domain.Document{}, nil, err
	}
	return document, content, nil
}

func (s *DocumentService) authorize(ctx context.Context, user domain.User, procedureID string) error {
	ok, err := s.access.CanAccess(ctx, user, procedureID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNoProcedureAccess
	}
	return nil
}

func (s *DocumentService) rule(ctx context.Context, user domain.User, upload input.Upload) (domain.DocumentRule, error) {
	if upload.RequirementID != "" {
		return s.requirements.DocumentRule(ctx, upload.ProcedureID, upload.RequirementID)
	}
	if !user.IsStaff() {
		return domain.DocumentRule{}, domain.ErrRequirementRequired
	}
	return s.config.UnlinkedRule, nil
}

// detectContentType returns the bare media type, without parameters such as
// the charset.
func detectContentType(head []byte) string {
	mediaType, _, err := mime.ParseMediaType(http.DetectContentType(head))
	if err != nil {
		return "application/octet-stream"
	}
	return mediaType
}

// cleanFilename keeps only the base name the client sent, without control
// characters and within a length any file system accepts. It is only shown
// to people: the stored file is named after the document ID.
func cleanFilename(name string) string {
	name = path.Base(strings.ReplaceAll(name, `\`, "/"))
	name = strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name))
	for len(name) > maxFilenameLen {
		_, size := utf8.DecodeLastRuneInString(name)
		name = name[:len(name)-size]
	}
	if name == "" || name == "." || name == "/" {
		return "documento"
	}
	return name
}
