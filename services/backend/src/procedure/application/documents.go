package application

import (
	"bufio"
	"cmp"
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
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/version"
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

	// unreferencedAge is how old a file that no document names has to be
	// for RemoveUnreferencedFiles to remove it: a younger one can be from an
	// upload, of this instance or another, that has not been recorded yet.
	unreferencedAge = time.Hour
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
	if p.Closed() {
		return domain.Document{}, fmt.Errorf("procedure %s: %w", p.ID, engine.ErrStepNotOpen)
	}
	step, requirement, err := s.whereToAttach(ctx, userID, p, content, u.Requirement)
	if err != nil {
		return domain.Document{}, err
	}
	if err := s.belowMaxFilesFor(ctx, userID, p, requirement.ID); err != nil {
		return domain.Document{}, err
	}

	// If recording the document fails after this, no record names the stored
	// file: it is harmless, and a retry with the same content reuses it.
	stored, err := storeFile(ctx, s.files, requirement, u.FileName, u.Content)
	if err != nil {
		return domain.Document{}, err
	}
	d := domain.Document{
		ID:          s.newID(),
		Procedure:   p.ID,
		Requirement: requirement.ID,
		Step:        step,
		FileName:    stored.fileName,
		MediaType:   stored.mediaType,
		Size:        stored.size,
		Digest:      stored.digest,
		UploadedBy:  userID,
		UploadedAt:  s.now(),
	}
	err = s.procedures.procedures.Attach(ctx, d, attachedEvent(d))
	if errors.Is(err, engine.ErrStepNotOpen) && step != p.OpenStep {
		// An answer to the corrections came first: the applicant is where
		// whereToAttach leaves them with nothing pending, and is told the same.
		err = denied(userID, actionAttach+" on "+step)
	}
	if err != nil {
		return domain.Document{}, fmt.Errorf("record document: %w", err)
	}
	return d, nil
}

func attachedEvent(d domain.Document) domain.Event {
	return domain.Event{
		ProcedureID: d.Procedure,
		Step:        d.Step,
		Action:      definition.ActionUploadDocument,
		By:          d.UploadedBy,
		At:          d.UploadedAt,
		Document:    &domain.DocumentRecord{ID: d.ID, Requirement: d.Requirement, FileName: d.FileName},
	}
}

// List shows the applicant only the documents they see (detail.Applicant),
// and which of those is current for each requirement.
func (s *DocumentService) List(ctx context.Context, userID, procedureID string) ([]input.AttachedDocument, error) {
	p, content, err := s.procedures.visible(ctx, userID, procedureID)
	if err != nil {
		return nil, err
	}
	documents, err := s.documents.List(ctx, procedureID)
	if err != nil {
		return nil, fmt.Errorf("documents of %s: %w", procedureID, err)
	}
	if p.Applicant == userID {
		documents = detail.ForApplicant(content, p).Documents(documents)
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

// Open answers to the applicant, for a document they do not see, as for one
// that does not exist.
func (s *DocumentService) Open(ctx context.Context, userID, procedureID, documentID string) (domain.Document, io.ReadCloser, error) {
	p, procedureContent, err := s.procedures.visible(ctx, userID, procedureID)
	if err != nil {
		return domain.Document{}, nil, err
	}
	d, err := s.documents.Get(ctx, procedureID, documentID)
	if err != nil {
		return domain.Document{}, nil, fmt.Errorf("document %s of %s: %w", documentID, procedureID, err)
	}
	if p.Applicant == userID && !detail.ForApplicant(procedureContent, p).Sees(d) {
		return domain.Document{}, nil, fmt.Errorf("document %s of %s: %w", documentID, procedureID, domain.ErrDocumentNotFound)
	}
	content, err := s.files.Open(ctx, d.Digest)
	if err != nil {
		return domain.Document{}, nil, fmt.Errorf("document %s of %s: %w", documentID, procedureID, err)
	}
	return d, content, nil
}

// belowMaxFilesFor bounds the files of a document. For the applicant answering
// a correction, it counts only what they attached since the last request that
// asks for the document: each request of staff opens room again, so a
// document that reached the bound can still be answered, and the applicant
// alone still cannot fill the disk.
func (s *DocumentService) belowMaxFilesFor(ctx context.Context, userID string, p domain.Procedure, requirement string) error {
	if p.Applicant != userID || !p.AwaitingApplicant() {
		return s.belowMaxFiles(ctx, p.ID, requirement)
	}
	history, err := s.procedures.procedures.Events(ctx, p.ID)
	if err != nil {
		return fmt.Errorf("history of %s: %w", p.ID, err)
	}
	if engine.AttachedSinceRequest(history, requirement) >= maxFilesPerDocument {
		return fmt.Errorf("%s of %s: %w", requirement, p.ID, domain.ErrTooManyDocuments)
	}
	return nil
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

// RemoveUnreferencedFiles removes the stored files that no document names
// and that are older than unreferencedAge, and returns how many it removed.
// They come from an upload whose record failed or, when the records are in
// memory, from every document of a previous run, which no one can reach
// anymore and which hold personal data. It runs at startup, before serving.
// With the records in memory it assumes a single instance: another one would
// lose the files of its documents.
func (s *DocumentService) RemoveUnreferencedFiles(ctx context.Context) (int, error) {
	// Read before the records, so that a file stored and named meanwhile is
	// not taken for unnamed.
	stored, err := s.files.Digests(ctx, s.now().Add(-unreferencedAge))
	if err != nil {
		return 0, fmt.Errorf("stored files: %w", err)
	}
	digests, err := s.documents.Digests(ctx)
	if err != nil {
		return 0, fmt.Errorf("digests of the documents: %w", err)
	}
	named := make(map[string]bool, len(digests))
	for _, digest := range digests {
		named[digest] = true
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

// whereToAttach returns the step a document goes to: the open step, for
// whoever can act on it, or, for the applicant with a pending correction that
// asks for it, the step of theirs that asks for it (ADR-026).
func (s *DocumentService) whereToAttach(ctx context.Context, userID string, p domain.Procedure, content version.Content, documentID string) (string, definition.Document, error) {
	if p.Applicant == userID && p.AwaitingApplicant() {
		step, err := engine.CorrectionDocument(p, content.Definition, documentID)
		if err != nil {
			return "", definition.Document{}, fmt.Errorf("procedure %s: %w", p.ID, err)
		}
		requirement, err := requested(content.Definition, step.ID, documentID)
		return step.ID, requirement, err
	}
	allowed, err := s.procedures.canAct(ctx, userID, p, content, p.OpenStep)
	if err != nil {
		return "", definition.Document{}, err
	}
	if !allowed {
		return "", definition.Document{}, denied(userID, actionAttach+" on "+p.OpenStep)
	}
	requirement, err := requested(content.Definition, p.OpenStep, documentID)
	if err != nil {
		return "", definition.Document{}, fmt.Errorf("procedure %s: %w", p.ID, err)
	}
	return p.OpenStep, requirement, nil
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

type storedFile struct {
	fileName  string
	mediaType string
	digest    string
	size      int64
}

// storeFile checks the file against the document it covers, by the type read
// from its content and its size, and keeps it in the store. It fails with
// domain.ErrDocumentEmpty, domain.ErrDocumentTypeNotAccepted or
// domain.ErrDocumentTooLarge.
func storeFile(ctx context.Context, files output.Files, requirement definition.Document, fileName string, content io.Reader) (storedFile, error) {
	file := bufio.NewReaderSize(content, sniffLength)
	head, err := file.Peek(sniffLength)
	if err != nil && !errors.Is(err, io.EOF) {
		return storedFile{}, fmt.Errorf("read the upload: %w", err)
	}
	if len(head) == 0 {
		return storedFile{}, domain.ErrDocumentEmpty
	}
	mediaType := detectType(head)
	if !slices.Contains(requirement.AcceptedTypes, mediaType) {
		return storedFile{}, fmt.Errorf("%s for %s: %w", mediaType, requirement.ID, domain.ErrDocumentTypeNotAccepted)
	}
	digest, size, err := files.Put(ctx, &sizeLimit{reader: file, left: requirement.MaxSizeBytes})
	if err != nil {
		return storedFile{}, fmt.Errorf("store the file: %w", err)
	}
	return storedFile{
		fileName:  withExtension(cleanFileName(fileName), mediaType),
		mediaType: mediaType,
		digest:    digest,
		size:      size,
	}, nil
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

// preferredExtensions are the usual extensions of the types for which mime
// lists another one first, such as .jfif for image/jpeg. The list of mime
// depends on the mime.types of the machine.
var preferredExtensions = map[string]string{
	"application/zip": ".zip",
	"image/jpeg":      ".jpg",
	"text/html":       ".html",
	"text/plain":      ".txt",
}

// withExtension makes the name end in an extension of the detected type, so
// that the downloaded file opens as what it is and not as what its name says.
// An extension of another type is replaced, so a PNG named foto.pdf becomes
// foto.png; one that names no type is kept, as in informe.final.pdf.
func withExtension(name, mediaType string) string {
	extensions, err := mime.ExtensionsByType(mediaType)
	ext := path.Ext(name)
	if err != nil || len(extensions) == 0 || slices.Contains(extensions, strings.ToLower(ext)) {
		return name
	}
	if ext != "" && mime.TypeByExtension(ext) != "" {
		name = strings.TrimSuffix(name, ext)
	}
	return name + cmp.Or(preferredExtensions[mediaType], extensions[0])
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
