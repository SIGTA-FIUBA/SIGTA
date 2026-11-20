package memory

import (
	"context"
	"sync"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var _ output.AdministratorRepository = (*AdministratorRepository)(nil)

type AdministratorRepository struct {
	mu  sync.Mutex
	ids map[string]bool
}

func NewAdministratorRepository(ids ...string) *AdministratorRepository {
	a := &AdministratorRepository{ids: map[string]bool{}}
	for _, id := range ids {
		a.ids[id] = true
	}
	return a
}

func (a *AdministratorRepository) IsAdministrator(_ context.Context, userID string) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ids[userID], nil
}

func (a *AdministratorRepository) GrantFirst(_ context.Context, administrator domain.Administrator) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.ids) > 0 {
		return false, nil
	}
	a.ids[administrator.UserID] = true
	return true, nil
}
