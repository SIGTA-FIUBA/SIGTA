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
	values     map[string][]domain.FieldValue
	events     map[string][]domain.Event
}

func NewProcedureRepository() *ProcedureRepository {
	return &ProcedureRepository{
		procedures: map[string]domain.Procedure{},
		values:     map[string][]domain.FieldValue{},
		events:     map[string][]domain.Event{},
	}
}

func (s *ProcedureRepository) Add(_ context.Context, p domain.Procedure, events []domain.Event, values ...domain.FieldValue) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.procedures[p.ID] = cloneProcedure(p)
	for _, e := range events {
		s.events[p.ID] = append(s.events[p.ID], cloneEvent(e))
	}
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

func (s *ProcedureRepository) Advance(_ context.Context, p domain.Procedure, e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.procedures[p.ID]
	if !ok {
		return domain.ErrProcedureNotFound
	}
	if stored.Closed() || stored.OpenStep != e.Step {
		return fmt.Errorf("procedure %s: %w", p.ID, engine.ErrStepNotOpen)
	}
	if stored.AwaitingApplicant() {
		return fmt.Errorf("procedure %s: %w", p.ID, engine.ErrAwaitingApplicant)
	}
	s.procedures[p.ID] = cloneProcedure(p)
	s.events[p.ID] = append(s.events[p.ID], cloneEvent(e))
	return nil
}

func (s *ProcedureRepository) RequestCorrection(_ context.Context, r domain.CorrectionRequest, e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.procedures[e.ProcedureID]
	if !ok {
		return domain.ErrProcedureNotFound
	}
	if stored.Closed() || stored.OpenStep != r.Step {
		return fmt.Errorf("procedure %s: %w", e.ProcedureID, engine.ErrStepNotOpen)
	}
	stored.CorrectionRequests = append(stored.CorrectionRequests, r)
	s.procedures[e.ProcedureID] = cloneProcedure(stored)
	s.events[e.ProcedureID] = append(s.events[e.ProcedureID], cloneEvent(e))
	return nil
}

func (s *ProcedureRepository) EnterValue(_ context.Context, v domain.FieldValue, e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.procedures[e.ProcedureID]
	if !ok {
		return domain.ErrProcedureNotFound
	}
	if !stored.AwaitingApplicant() {
		return fmt.Errorf("procedure %s: %w", e.ProcedureID, engine.ErrNoCorrectionPending)
	}
	s.values[e.ProcedureID] = append(s.values[e.ProcedureID], cloneValues([]domain.FieldValue{v})...)
	s.events[e.ProcedureID] = append(s.events[e.ProcedureID], cloneEvent(e))
	return nil
}

func (s *ProcedureRepository) AnswerCorrection(_ context.Context, p domain.Procedure, requested int, e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.procedures[p.ID]
	if !ok {
		return domain.ErrProcedureNotFound
	}
	if !stored.AwaitingApplicant() {
		return fmt.Errorf("procedure %s: %w", p.ID, engine.ErrNoCorrectionPending)
	}
	if engine.CorrectionRequestsIn(s.events[p.ID]) != requested || stored.OpenStep != p.OpenStep {
		return fmt.Errorf("procedure %s: %w", p.ID, engine.ErrCorrectionsChanged)
	}
	s.procedures[p.ID] = cloneProcedure(p)
	s.events[p.ID] = append(s.events[p.ID], cloneEvent(e))
	return nil
}

func (s *ProcedureRepository) Record(_ context.Context, e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.procedures[e.ProcedureID]; !ok {
		return domain.ErrProcedureNotFound
	}
	s.events[e.ProcedureID] = append(s.events[e.ProcedureID], cloneEvent(e))
	return nil
}

func (s *ProcedureRepository) Events(_ context.Context, procedureID string) ([]domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.procedures[procedureID]; !ok {
		return nil, domain.ErrProcedureNotFound
	}
	events := make([]domain.Event, 0, len(s.events[procedureID]))
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
	requests := make([]domain.CorrectionRequest, len(p.CorrectionRequests))
	for i, r := range p.CorrectionRequests {
		r.Fields = slices.Clone(r.Fields)
		r.Documents = slices.Clone(r.Documents)
		requests[i] = r
	}
	p.CorrectionRequests = requests
	return p
}

func cloneEvent(e domain.Event) domain.Event {
	if e.Referral != nil {
		r := *e.Referral
		e.Referral = &r
	}
	if e.Document != nil {
		d := *e.Document
		e.Document = &d
	}
	if e.Value != nil {
		v := *e.Value
		e.Value = &v
	}
	if e.CorrectionAnswer != nil {
		e.CorrectionAnswer = &domain.CorrectionRecord{
			Fields:    slices.Clone(e.CorrectionAnswer.Fields),
			Documents: slices.Clone(e.CorrectionAnswer.Documents),
		}
	}
	if e.CorrectionRequest != nil {
		e.CorrectionRequest = &domain.CorrectionRecord{
			Fields:    slices.Clone(e.CorrectionRequest.Fields),
			Documents: slices.Clone(e.CorrectionRequest.Documents),
		}
	}
	return e
}
