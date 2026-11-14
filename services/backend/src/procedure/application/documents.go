package application

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.Documents = (*DocumentService)(nil)

const (
	actionAttach = "attach"
	// sniffLength is all that http.DetectContentType reads.
	sniffLength = 512
	// maxFileName keeps a long name from filling the record and the download
	// header.
	maxFileName     = 200
	defaultFileName = "documento"
	// maxFilesPerDocument bounds how many files a procedure keeps for one
	// document, since every replaced file stays recorded: without it, one
	// person could fill the disk by uploading the same document again and
	// again. The number is a guess, to confirm with the Secretaría.
	maxFilesPerDocument = 20
)

// DocumentService decides who sees a procedure and who acts on its open step
// as ProcedureService does, so both answer the same to the same person.
type DocumentService struct {
	procedures *ProcedureService
	documents  output.DocumentRepository
	files      output.Files
	newID      func() string
	now        func() time.Time
}

func NewDocumentService(procedures *ProcedureService, documents output.DocumentRepository, files output.Files, newID func() string, now func() time.Time) *DocumentService {
	return &DocumentService{procedures: procedures, documents: documents, files: files, newID: newID, now: now}
}

func (s *DocumentService) Attach(ctx context.Context, userID string, u input.Upload) (domain.Document, error) {
	p, content, err := s.procedures.visible(ctx, userID, u.ProcedureID)
	if err != nil {
		return domain.Document{}, err
	}
	allowed, err := s.procedures.canAct(ctx, userID, p, content, p.OpenStep)
	if err != nil {
		return domain.Document{}, err
	}
	if !allowed {
		return domain.Document{}, denied(userID, actionAttach+" on "+p.OpenStep)
	}
	requirement, err := requested(content.Definition, p.OpenStep, u.Requirement)
	if err != nil {
		return domain.Document{}, fmt.Errorf("procedure %s: %w", p.ID, err)
	}
	if err := s.belowMaxFiles(ctx, p.ID, requirement.ID); err != nil {
		return domain.Document{}, err
	}

	file := bufio.NewReaderSize(u.Content, sniffLength)
	head, err := file.Peek(sniffLength)
	if err != nil && !errors.Is(err, io.EOF) {
		return domain.Document{}, fmt.Errorf("read the upload: %w", err)
	}
	if len(head) == 0 {
		return domain.Document{}, domain.ErrEmptyDocument
	}
	mediaType := detectType(head)
	if !slices.Contains(requirement.AcceptedTypes, mediaType) {
		return domain.Document{}, fmt.Errorf("%s for %s: %w", mediaType, requirement.ID, domain.ErrDocumentType)
	}
	// If recording the document fails after this, no record names the stored
	// file: it is harmless, and a retry with the same content reuses it.
	digest, size, err := s.files.Put(ctx, &sizeLimit{reader: file, left: requirement.MaxSizeBytes})
	if err != nil {
		return domain.Document{}, fmt.Errorf("store the file: %w", err)
	}
	d := domain.Document{
		ID:          s.newID(),
		Procedure:   p.ID,
		Requirement: requirement.ID,
		Step:        p.OpenStep,
		FileName:    withExtension(cleanFileName(u.FileName), mediaType),
		MediaType:   mediaType,
		Size:        size,
		Digest:      digest,
		UploadedBy:  userID,
		UploadedAt:  s.now(),
	}
	if err := s.documents.Add(ctx, d); err != nil {
		return domain.Document{}, fmt.Errorf("record document: %w", err)
	}
	return d, nil
}

func (s *DocumentService) List(ctx context.Context, userID, procedureID string) ([]input.AttachedDocument, error) {
	if _, _, err := s.procedures.visible(ctx, userID, procedureID); err != nil {
		return nil, err
	}
	documents, err := s.documents.List(ctx, procedureID)
	if err != nil {
		return nil, fmt.Errorf("documents of %s: %w", procedureID, err)
	}
	current := map[string]bool{}
	for _, d := range domain.Current(documents) {
		current[d.ID] = true
	}
	attached := make([]input.AttachedDocument, 0, len(documents))
	for _, d := range documents {
		attached = append(attached, input.AttachedDocument{Document: d, Current: current[d.ID]})
	}
	return attached, nil
}

func (s *DocumentService) Open(ctx context.Context, userID, procedureID, documentID string) (domain.Document, io.ReadCloser, error) {
	if _, _, err := s.procedures.visible(ctx, userID, procedureID); err != nil {
		return domain.Document{}, nil, err
	}
	d, err := s.documents.Get(ctx, procedureID, documentID)
	if err != nil {
		return domain.Document{}, nil, fmt.Errorf("document %s of %s: %w", documentID, procedureID, err)
	}
	content, err := s.files.Open(ctx, d.Digest)
	if err != nil {
		return domain.Document{}, nil, fmt.Errorf("document %s of %s: %w", documentID, procedureID, err)
	}
	return d, content, nil
}

// belowMaxFiles checks before storing anything. Two uploads at the same time
// can both pass it, which goes past the maximum by a few files at most.
func (s *DocumentService) belowMaxFiles(ctx context.Context, procedureID, requirement string) error {
	documents, err := s.documents.List(ctx, procedureID)
	if err != nil {
		return fmt.Errorf("documents of %s: %w", procedureID, err)
	}
	n := 0
	for _, d := range documents {
		if d.Requirement == requirement {
			n++
		}
	}
	if n >= maxFilesPerDocument {
		return fmt.Errorf("%s of %s: %w", requirement, procedureID, domain.ErrTooManyDocuments)
	}
	return nil
}

// RemoveUnreferencedFiles removes the stored files that no document names,
// and returns how many it removed. They come from an upload whose record
// failed or, while the records are in memory, from every document of the
// previous run, which no one can reach anymore and which hold personal data.
// It runs at startup, before serving: during an upload the file is stored
// before its record, so it would remove a file about to be named.
func (s *DocumentService) RemoveUnreferencedFiles(ctx context.Context) (int, error) {
	digests, err := s.documents.Digests(ctx)
	if err != nil {
		return 0, fmt.Errorf("digests of the documents: %w", err)
	}
	named := make(map[string]bool, len(digests))
	for _, digest := range digests {
		named[digest] = true
	}
	stored, err := s.files.Digests(ctx)
	if err != nil {
		return 0, fmt.Errorf("stored files: %w", err)
	}
	removed := 0
	for _, digest := range stored {
		if named[digest] {
			continue
		}
		if err := s.files.Remove(ctx, digest); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

// requested returns the document if the step asks for it, as required or
// optional.
func requested(def definition.ProcedureTypeDefinition, stepID, documentID string) (definition.Document, error) {
	i := slices.IndexFunc(def.Steps, func(step definition.Step) bool { return step.ID == stepID })
	if i < 0 {
		return definition.Document{}, fmt.Errorf("open step %q is not in the definition", stepID)
	}
	step := def.Steps[i]
	if !slices.Contains(step.Documents, documentID) && !slices.Contains(step.OptionalDocuments, documentID) {
		return definition.Document{}, fmt.Errorf("%s on %s: %w", documentID, stepID, domain.ErrDocumentNotRequested)
	}
	j := slices.IndexFunc(def.Documents, func(d definition.Document) bool { return d.ID == documentID })
	if j < 0 {
		return definition.Document{}, fmt.Errorf("step %q asks for %q, which is not in the definition", stepID, documentID)
	}
	return def.Documents[j], nil
}

// detectType reads the type from the content, without the parameters such as
// the charset, which the accepted types do not have.
func detectType(head []byte) string {
	detected := http.DetectContentType(head)
	mediaType, _, err := mime.ParseMediaType(detected)
	if err != nil {
		return detected
	}
	return mediaType
}

// cleanFileName keeps only the base name the browser sent, without control
// or format characters, since it is shown and sent back in the download
// header. A format character such as U+202E reverses the text, so that a name
// ending in "fdp.exe" reads as one ending in "exe.pdf".
func cleanFileName(name string) string {
	name = path.Base(strings.ReplaceAll(name, `\`, "/"))
	name = strings.Map(func(r rune) rune {
		if r == utf8.RuneError || unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == "/" || name == ".." {
		return defaultFileName
	}
	if runes := []rune(name); len(runes) > maxFileName {
		name = string(runes[len(runes)-maxFileName:])
	}
	return name
}

// withExtension makes the name end in an extension of the detected type, so
// that the downloaded file opens as what it is and not as what its name says.
func withExtension(name, mediaType string) string {
	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil || len(extensions) == 0 || slices.Contains(extensions, strings.ToLower(path.Ext(name))) {
		return name
	}
	return name + extensions[0]
}

// sizeLimit fails the upload as soon as it goes past the maximum size, so the
// store discards it without reading the rest.
type sizeLimit struct {
	reader io.Reader
	left   int64
}

func (l *sizeLimit) Read(p []byte) (int, error) {
	n, err := l.reader.Read(p)
	l.left -= int64(n)
	if l.left < 0 {
		return 0, domain.ErrDocumentTooLarge
	}
	return n, err
}
