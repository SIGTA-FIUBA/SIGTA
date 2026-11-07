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
	mu          sync.Mutex
	kinds       map[string]domain.UserKind
	permissions map[string][]domain.Permission
}

func NewAccess() *Access {
	return &Access{kinds: map[string]domain.UserKind{}, permissions: map[string][]domain.Permission{}}
}

func (s *Access) Register(_ context.Context, userID string, kind domain.UserKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.kinds[userID] = kind
	return nil
}

func (s *Access) Actor(_ context.Context, userID string) (domain.Actor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	kind, ok := s.kinds[userID]
	if !ok {
		return domain.Actor{}, output.ErrUnknownUser
	}
	return domain.Actor{UserID: userID, Kind: kind, Permissions: slices.Clone(s.permissions[userID])}, nil
}

func (s *Access) Grant(_ context.Context, p domain.Permission) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	granted := slices.DeleteFunc(s.permissions[p.UserID], func(q domain.Permission) bool { return q.ProcedureType == p.ProcedureType })
	s.permissions[p.UserID] = append(granted, p)
	return nil
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
