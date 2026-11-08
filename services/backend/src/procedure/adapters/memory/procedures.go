package memory

import (
	"cmp"
	"context"
	"slices"
	"sync"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var (
	_ output.ProcedureRepository = (*ProcedureRepository)(nil)
	_ output.Access              = (*Access)(nil)
	_ output.DeniedAttempts      = (*DeniedAttempts)(nil)
)

type ProcedureRepository struct {
	mu         sync.Mutex
	procedures map[string]domain.Procedure
}

func NewProcedureRepository() *ProcedureRepository {
	return &ProcedureRepository{procedures: map[string]domain.Procedure{}}
}

func (s *ProcedureRepository) Add(_ context.Context, p domain.Procedure) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.procedures[p.ID] = p
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

type Access struct {
	mu             sync.Mutex
	administrators map[string]bool
	memberships    []domain.Membership
}

func NewAccess(administrators ...string) *Access {
	a := &Access{administrators: map[string]bool{}}
	for _, id := range administrators {
		a.administrators[id] = true
	}
	return a
}

func (s *Access) Administrator(_ context.Context, userID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.administrators[userID], nil
}

func (s *Access) Memberships(_ context.Context, userID string) ([]domain.Membership, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var found []domain.Membership
	for _, m := range s.memberships {
		if m.UserID == userID {
			found = append(found, cloneMembership(m))
		}
	}
	return found, nil
}

func (s *Access) Membership(_ context.Context, id string) (domain.Membership, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := slices.IndexFunc(s.memberships, func(m domain.Membership) bool { return m.ID == id })
	if i < 0 {
		return domain.Membership{}, domain.ErrMembershipNotFound
	}
	return cloneMembership(s.memberships[i]), nil
}

func (s *Access) SaveMembership(_ context.Context, m domain.Membership) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m = cloneMembership(m)
	if i := slices.IndexFunc(s.memberships, func(q domain.Membership) bool { return q.ID == m.ID }); i >= 0 {
		s.memberships[i] = m
		return nil
	}
	s.memberships = append(s.memberships, m)
	return nil
}

func cloneMembership(m domain.Membership) domain.Membership {
	m.Scope = slices.Clone(m.Scope)
	return m
}

type DeniedAttempts struct {
	mu       sync.Mutex
	attempts []domain.DeniedAttempt
}

func NewDeniedAttempts() *DeniedAttempts {
	return &DeniedAttempts{}
}

func (s *DeniedAttempts) Record(_ context.Context, a domain.DeniedAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts = append(s.attempts, a)
	return nil
}

func (s *DeniedAttempts) All() []domain.DeniedAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.attempts)
}
