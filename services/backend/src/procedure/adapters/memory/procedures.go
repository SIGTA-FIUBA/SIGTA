package memory

import (
	"cmp"
	"context"
	"slices"
	"sync"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureRepository = (*ProcedureRepository)(nil)

type ProcedureRepository struct {
	mu         sync.Mutex
	procedures map[string]domain.Procedure
	values     map[string][]domain.FieldValue
}

func NewProcedureRepository() *ProcedureRepository {
	return &ProcedureRepository{procedures: map[string]domain.Procedure{}, values: map[string][]domain.FieldValue{}}
}

func (s *ProcedureRepository) Add(_ context.Context, p domain.Procedure, values ...domain.FieldValue) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.procedures[p.ID] = p
	s.values[p.ID] = append(s.values[p.ID], cloneValues(values)...)
	return nil
}

func (s *ProcedureRepository) Get(_ context.Context, id string) (domain.Procedure, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.procedures[id]
	if !ok {
		return domain.Procedure{}, domain.ErrProcedureNotFound
	}
	return p, nil
}

func (s *ProcedureRepository) List(context.Context) ([]domain.Procedure, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := make([]domain.Procedure, 0, len(s.procedures))
	for _, p := range s.procedures {
		all = append(all, p)
	}
	slices.SortFunc(all, func(a, b domain.Procedure) int { return cmp.Compare(a.ID, b.ID) })
	return all, nil
}

func (s *ProcedureRepository) Values(_ context.Context, procedureID string) ([]domain.FieldValue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneValues(s.values[procedureID]), nil
}

// cloneValues keeps the caller from changing the stored items.
func cloneValues(values []domain.FieldValue) []domain.FieldValue {
	cloned := make([]domain.FieldValue, len(values))
	for i, v := range values {
		v.Items = slices.Clone(v.Items)
		cloned[i] = v
	}
	return cloned
}
