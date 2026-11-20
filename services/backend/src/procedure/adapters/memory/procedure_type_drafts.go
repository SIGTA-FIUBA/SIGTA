package memory

import (
	"cmp"
	"context"
	"encoding/json"
	"slices"
	"sync"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/draft"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureTypeDraftRepository = (*ProcedureTypeDraftRepository)(nil)

type ProcedureTypeDraftRepository struct {
	mu     sync.Mutex
	drafts map[int64]draft.Draft
}

func NewProcedureTypeDraftRepository() *ProcedureTypeDraftRepository {
	return &ProcedureTypeDraftRepository{drafts: map[int64]draft.Draft{}}
}

func (s *ProcedureTypeDraftRepository) Get(_ context.Context, procedureType int64) (draft.Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.drafts[procedureType]
	if !ok {
		return draft.Draft{}, output.ErrDraftNotFound
	}
	return cloneDraft(d)
}

func (s *ProcedureTypeDraftRepository) List(context.Context) ([]draft.Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	drafts := make([]draft.Draft, 0, len(s.drafts))
	for _, d := range s.drafts {
		c, err := cloneDraft(d)
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, c)
	}
	slices.SortFunc(drafts, func(left, right draft.Draft) int {
		return cmp.Compare(left.ProcedureTypeID, right.ProcedureTypeID)
	})
	return drafts, nil
}

func (s *ProcedureTypeDraftRepository) Add(_ context.Context, d draft.Draft) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.drafts[d.ProcedureTypeID]; ok {
		return output.ErrDraftExists
	}
	return s.put(d)
}

func (s *ProcedureTypeDraftRepository) Update(_ context.Context, d draft.Draft) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.at(d.ProcedureTypeID, d.Revision-1); err != nil {
		return err
	}
	return s.put(d)
}

func (s *ProcedureTypeDraftRepository) Remove(_ context.Context, procedureType int64, revision int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.at(procedureType, revision); err != nil {
		return err
	}
	delete(s.drafts, procedureType)
	return nil
}

func (s *ProcedureTypeDraftRepository) at(procedureType int64, revision int) error {
	stored, ok := s.drafts[procedureType]
	switch {
	case !ok:
		return output.ErrDraftNotFound
	case stored.Revision != revision:
		return output.ErrDraftChanged
	}
	return nil
}

func (s *ProcedureTypeDraftRepository) put(d draft.Draft) error {
	c, err := cloneDraft(d)
	if err != nil {
		return err
	}
	s.drafts[d.ProcedureTypeID] = c
	return nil
}

// cloneDraft copies the definition through JSON, as PostgreSQL stores it, and
// the editors, so that neither the caller nor the store shares its slices.
func cloneDraft(d draft.Draft) (draft.Draft, error) {
	d.Editors = slices.Clone(d.Editors)
	encoded, err := json.Marshal(d.Definition)
	if err != nil {
		return draft.Draft{}, err
	}
	d.Definition = definition.ProcedureTypeDefinition{}
	return d, json.Unmarshal(encoded, &d.Definition)
}
