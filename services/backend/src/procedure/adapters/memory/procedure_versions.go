package memory

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureVersions = (*ProcedureVersions)(nil)

type ProcedureVersions struct {
	mu       sync.Mutex
	versions map[int64][]domain.Version
}

func NewProcedureVersions() *ProcedureVersions {
	return &ProcedureVersions{versions: map[int64][]domain.Version{}}
}

func (s *ProcedureVersions) Latest(_ context.Context, procedureType int64) (domain.Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.versions[procedureType]
	if len(stored) == 0 {
		return domain.Version{}, output.ErrNoVersion
	}
	return clone(stored[len(stored)-1]), nil
}

func (s *ProcedureVersions) Get(_ context.Context, procedureType int64, number int) (domain.Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.versions[procedureType]
	if number < 1 || number > len(stored) {
		return domain.Version{}, output.ErrNoVersion
	}
	return clone(stored[number-1]), nil
}

func (s *ProcedureVersions) Add(_ context.Context, v domain.Version) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.versions[v.ProcedureType]
	next := len(stored) + 1
	switch {
	case v.Number >= 1 && v.Number < next:
		return output.ErrVersionExists
	case v.Number != next:
		return fmt.Errorf("version %d is not the next one, %d", v.Number, next)
	}
	s.versions[v.ProcedureType] = append(stored, clone(v))
	return nil
}

func clone(v domain.Version) domain.Version {
	v.Encoded = slices.Clone(v.Encoded)
	return v
}
