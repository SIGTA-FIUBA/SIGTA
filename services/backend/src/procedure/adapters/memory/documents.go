package memory

import (
	"context"
	"slices"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var _ output.DocumentRepository = (*DocumentRepository)(nil)

// DocumentRepository reads the documents that a ProcedureRepository records.
type DocumentRepository struct {
	procedures *ProcedureRepository
}

func NewDocumentRepository(procedures *ProcedureRepository) *DocumentRepository {
	return &DocumentRepository{procedures: procedures}
}

func (r *DocumentRepository) List(_ context.Context, procedureID string) ([]domain.Document, error) {
	r.procedures.mu.Lock()
	defer r.procedures.mu.Unlock()
	var documents []domain.Document
	for _, d := range r.procedures.documents {
		if d.Procedure == procedureID {
			documents = append(documents, d)
		}
	}
	return documents, nil
}

func (r *DocumentRepository) Get(_ context.Context, procedureID, documentID string) (domain.Document, error) {
	r.procedures.mu.Lock()
	defer r.procedures.mu.Unlock()
	i := slices.IndexFunc(r.procedures.documents, func(d domain.Document) bool {
		return d.Procedure == procedureID && d.ID == documentID
	})
	if i < 0 {
		return domain.Document{}, domain.ErrDocumentNotFound
	}
	return r.procedures.documents[i], nil
}

func (r *DocumentRepository) Digests(_ context.Context) ([]string, error) {
	r.procedures.mu.Lock()
	defer r.procedures.mu.Unlock()
	seen := map[string]bool{}
	var digests []string
	for _, d := range r.procedures.documents {
		if !seen[d.Digest] {
			seen[d.Digest] = true
			digests = append(digests, d.Digest)
		}
	}
	return digests, nil
}
