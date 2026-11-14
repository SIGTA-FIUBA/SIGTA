package memory

import (
	"context"
	"slices"
	"sync"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var _ output.DocumentRepository = (*DocumentRepository)(nil)

type DocumentRepository struct {
	mu        sync.Mutex
	documents []domain.Document
}

func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{}
}

func (r *DocumentRepository) Add(_ context.Context, d domain.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.documents = append(r.documents, d)
	return nil
}

func (r *DocumentRepository) List(_ context.Context, procedureID string) ([]domain.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var documents []domain.Document
	for _, d := range r.documents {
		if d.Procedure == procedureID {
			documents = append(documents, d)
		}
	}
	return documents, nil
}

func (r *DocumentRepository) Get(_ context.Context, procedureID, documentID string) (domain.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i := slices.IndexFunc(r.documents, func(d domain.Document) bool {
		return d.Procedure == procedureID && d.ID == documentID
	})
	if i < 0 {
		return domain.Document{}, domain.ErrDocumentNotFound
	}
	return r.documents[i], nil
}

func (r *DocumentRepository) Digests(_ context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := map[string]bool{}
	var digests []string
	for _, d := range r.documents {
		if !seen[d.Digest] {
			seen[d.Digest] = true
			digests = append(digests, d.Digest)
		}
	}
	return digests, nil
}
