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

var _ output.ProcedureVersions = (*ProcedureVersions)(nil)

type ProcedureVersions struct {
	mu       sync.Mutex
	versions map[int64][]version.ProcedureTypeVersion
}

func NewProcedureVersions() *ProcedureVersions {
	return &ProcedureVersions{versions: map[int64][]version.ProcedureTypeVersion{}}
}

func (s *ProcedureVersions) Latest(_ context.Context, procedureType int64) (version.ProcedureTypeVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.versions[procedureType]
	if len(stored) == 0 {
		return version.ProcedureTypeVersion{}, output.ErrNoVersion
	}
	return clone(stored[len(stored)-1]), nil
}

func (s *ProcedureVersions) LatestOfEach(context.Context) ([]version.ProcedureTypeVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	latest := make([]version.ProcedureTypeVersion, 0, len(s.versions))
	for _, stored := range s.versions {
		if len(stored) > 0 {
			latest = append(latest, clone(stored[len(stored)-1]))
		}
	}
	slices.SortFunc(latest, func(left, right version.ProcedureTypeVersion) int {
		return cmp.Compare(left.ProcedureTypeID, right.ProcedureTypeID)
	})
	return latest, nil
}

func (s *ProcedureVersions) Get(_ context.Context, procedureType int64, number int) (version.ProcedureTypeVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.versions[procedureType]
	if number < 1 || number > len(stored) {
		return version.ProcedureTypeVersion{}, output.ErrNoVersion
	}
	return clone(stored[number-1]), nil
}

func (s *ProcedureVersions) Add(_ context.Context, v version.ProcedureTypeVersion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.versions[v.ProcedureTypeID]
	next := len(stored) + 1
	switch {
	case v.Number >= 1 && v.Number < next:
		return output.ErrVersionExists
	case v.Number != next:
		return fmt.Errorf("version %d is not the next one, %d", v.Number, next)
	}
	s.versions[v.ProcedureTypeID] = append(stored, clone(v))
	return nil
}

func clone(v version.ProcedureTypeVersion) version.ProcedureTypeVersion {
	v.EncodedContent = slices.Clone(v.EncodedContent)
	return v
}
