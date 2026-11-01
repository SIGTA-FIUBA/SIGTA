package memory

import (
	"backend/src/domain"
	"bytes"
	"context"
	"io"
	"sync"
)

type DocumentRepository struct {
	mu        sync.RWMutex
	documents map[string]domain.Document
}

func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{documents: make(map[string]domain.Document)}
}

func (r *DocumentRepository) Save(_ context.Context, document domain.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.documents[document.ID] = document
	return nil
}

func (r *DocumentRepository) FindByID(_ context.Context, id string) (domain.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.documents[id]
	if !ok {
		return domain.Document{}, domain.ErrDocumentNotFound
	}
	return d, nil
}

// Procedure is the little the document storage needs to know about a
// procedure until the procedure model (SIG-12) exists.
type Procedure struct {
	ID          string
	ApplicantID string
	// Requirements maps each document requirement ID to its rule.
	Requirements map[string]domain.DocumentRule
}

// Procedures stands in for the procedure model (SIG-12) and authorization
// (SIG-14). Its access policy is a placeholder: the applicant reaches their
// own procedures and the Secretaría reaches every procedure.
type Procedures struct {
	mu         sync.RWMutex
	procedures map[string]Procedure
}

func NewProcedures(procedures ...Procedure) *Procedures {
	p := &Procedures{procedures: make(map[string]Procedure)}
	for _, procedure := range procedures {
		p.procedures[procedure.ID] = procedure
	}
	return p
}

func (p *Procedures) CanAccess(_ context.Context, user domain.User, procedureID string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	procedure, ok := p.procedures[procedureID]
	if !ok {
		return false, nil
	}
	return user.IsStaff() || procedure.ApplicantID == user.ID, nil
}

func (p *Procedures) DocumentRule(_ context.Context, procedureID, requirementID string) (domain.DocumentRule, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	rule, ok := p.procedures[procedureID].Requirements[requirementID]
	if !ok {
		return domain.DocumentRule{}, domain.ErrUnknownRequirement
	}
	return rule, nil
}

// DocumentStore keeps document contents in memory, for tests.
type DocumentStore struct {
	mu       sync.RWMutex
	contents map[string][]byte
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{contents: make(map[string][]byte)}
}

func (s *DocumentStore) Put(_ context.Context, key string, content io.Reader) (int64, error) {
	b, err := io.ReadAll(content)
	if err != nil {
		return int64(len(b)), err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contents[key] = b
	return int64(len(b)), nil
}

func (s *DocumentStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.contents[key]
	if !ok {
		return nil, domain.ErrDocumentNotFound
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (s *DocumentStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.contents, key)
	return nil
}

// Len is how many contents are stored.
func (s *DocumentStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.contents)
}
