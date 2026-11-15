package memory

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sync"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureRepository = (*ProcedureRepository)(nil)

type ProcedureRepository struct {
	mu         sync.Mutex
	procedures map[string]domain.Procedure
	events     map[string][]engine.Event
}

func NewProcedureRepository() *ProcedureRepository {
	return &ProcedureRepository{procedures: map[string]domain.Procedure{}, events: map[string][]engine.Event{}}
}

func (s *ProcedureRepository) Add(_ context.Context, p domain.Procedure) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.procedures[p.ID] = cloneProcedure(p)
	return nil
}

func (s *ProcedureRepository) Get(_ context.Context, id string) (domain.Procedure, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.procedures[id]
	if !ok {
		return domain.Procedure{}, domain.ErrProcedureNotFound
	}
	return cloneProcedure(p), nil
}

func (s *ProcedureRepository) List(context.Context) ([]domain.Procedure, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := make([]domain.Procedure, 0, len(s.procedures))
	for _, p := range s.procedures {
		all = append(all, cloneProcedure(p))
	}
	slices.SortFunc(all, func(a, b domain.Procedure) int { return cmp.Compare(a.ID, b.ID) })
	return all, nil
}

func (s *ProcedureRepository) Advance(_ context.Context, p domain.Procedure, e engine.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.procedures[p.ID]
	if !ok {
		return domain.ErrProcedureNotFound
	}
	if stored.Closed() || stored.OpenStep != e.Step {
		return fmt.Errorf("procedure %s: %w", p.ID, engine.ErrStepNotOpen)
	}
	s.procedures[p.ID] = cloneProcedure(p)
	s.events[p.ID] = append(s.events[p.ID], cloneEvent(e))
	return nil
}

func (s *ProcedureRepository) Events(_ context.Context, procedureID string) ([]engine.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.procedures[procedureID]; !ok {
		return nil, domain.ErrProcedureNotFound
	}
	events := make([]engine.Event, 0, len(s.events[procedureID]))
	for _, e := range s.events[procedureID] {
		events = append(events, cloneEvent(e))
	}
	return events, nil
}

// Copies keep callers from changing what is stored through a pointer.
func cloneProcedure(p domain.Procedure) domain.Procedure {
	if p.Referral != nil {
		r := *p.Referral
		p.Referral = &r
	}
	return p
}

func cloneEvent(e engine.Event) engine.Event {
	if e.Referral != nil {
		r := *e.Referral
		e.Referral = &r
	}
	return e
}
