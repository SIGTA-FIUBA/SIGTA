package memory

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sync"

	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureTypeVersionRepository = (*ProcedureTypeVersionRepository)(nil)

type ProcedureTypeVersionRepository struct {
	mu       sync.Mutex
	versions map[int64][]version.ProcedureTypeVersion
}

func NewProcedureTypeVersionRepository() *ProcedureTypeVersionRepository {
	return &ProcedureTypeVersionRepository{versions: map[int64][]version.ProcedureTypeVersion{}}
}

func (s *ProcedureTypeVersionRepository) Latest(_ context.Context, procedureType int64) (version.ProcedureTypeVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.versions[procedureType]
	if len(stored) == 0 {
		return version.ProcedureTypeVersion{}, output.ErrVersionNotFound
	}
	return cloneVersion(stored[len(stored)-1]), nil
}

func (s *ProcedureTypeVersionRepository) LatestOfEach(context.Context) ([]version.ProcedureTypeVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	latest := make([]version.ProcedureTypeVersion, 0, len(s.versions))
	for _, stored := range s.versions {
		if len(stored) > 0 {
			latest = append(latest, cloneVersion(stored[len(stored)-1]))
		}
	}
	slices.SortFunc(latest, func(left, right version.ProcedureTypeVersion) int {
		return cmp.Compare(left.ProcedureTypeID, right.ProcedureTypeID)
	})
	return latest, nil
}

func (s *ProcedureTypeVersionRepository) Get(_ context.Context, procedureType int64, number int) (version.ProcedureTypeVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.versions[procedureType]
	if number < 1 || number > len(stored) {
		return version.ProcedureTypeVersion{}, output.ErrVersionNotFound
	}
	return cloneVersion(stored[number-1]), nil
}

func (s *ProcedureTypeVersionRepository) Add(_ context.Context, v version.ProcedureTypeVersion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.versions[v.ProcedureTypeID]
	next := len(stored) + 1
	switch {
	case v.Number >= 1 && v.Number < next:
		return output.ErrVersionExists
	case v.Number != next:
		return fmt.Errorf("%w: version %d, the next one is %d", output.ErrVersionGap, v.Number, next)
	}
	s.versions[v.ProcedureTypeID] = append(stored, cloneVersion(v))
	return nil
}

func cloneVersion(v version.ProcedureTypeVersion) version.ProcedureTypeVersion {
	v.EncodedContent = slices.Clone(v.EncodedContent)
	return v
}
