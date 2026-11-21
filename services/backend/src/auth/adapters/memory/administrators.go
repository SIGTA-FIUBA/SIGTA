package memory

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var (
	_ output.FirstAdministratorRepository  = (*AdministratorRepository)(nil)
	_ output.AdministratorPeriodRepository = (*AdministratorRepository)(nil)
)

type AdministratorRepository struct {
	mu      sync.Mutex
	periods []domain.Administrator
}

// NewAdministratorRepository starts with each user of ids as an
// administrator since ever, as the initial grant of the system.
func NewAdministratorRepository(userIDs ...string) *AdministratorRepository {
	a := &AdministratorRepository{}
	for _, id := range userIDs {
		a.periods = append(a.periods, domain.Administrator{ID: "A-" + id, UserID: id})
	}
	return a
}

func (a *AdministratorRepository) IsAdministrator(_ context.Context, userID string) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	return slices.ContainsFunc(a.periods, func(p domain.Administrator) bool { return p.UserID == userID && p.ActiveAt(now) }), nil
}

func (a *AdministratorRepository) ListActive(_ context.Context, at time.Time) ([]domain.Administrator, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var active []domain.Administrator
	for _, p := range a.periods {
		if p.ActiveAt(at) {
			active = append(active, p)
		}
	}
	slices.SortStableFunc(active, func(x, y domain.Administrator) int { return x.From.Compare(y.From) })
	return active, nil
}

func (a *AdministratorRepository) Grant(_ context.Context, administrator domain.Administrator) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if slices.ContainsFunc(a.periods, func(p domain.Administrator) bool { return p.UserID == administrator.UserID && p.Until.IsZero() }) {
		return fmt.Errorf("%w: %s", domain.ErrAlreadyAdministrator, administrator.UserID)
	}
	a.periods = append(a.periods, administrator)
	return nil
}

func (a *AdministratorRepository) End(_ context.Context, id string, at time.Time, by string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	i := slices.IndexFunc(a.periods, func(p domain.Administrator) bool { return p.ID == id && p.ActiveAt(at) })
	if i < 0 {
		return fmt.Errorf("%w: %s", domain.ErrAdministratorNotFound, id)
	}
	if !slices.ContainsFunc(a.periods, func(p domain.Administrator) bool { return p.ID != id && p.ActiveAt(at) }) {
		return fmt.Errorf("%w: %s is the only one", domain.ErrLastAdministrator, id)
	}
	a.periods[i].Until, a.periods[i].EndedBy = at, by
	return nil
}

func (a *AdministratorRepository) GrantFirst(_ context.Context, administrator domain.Administrator) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if slices.ContainsFunc(a.periods, func(p domain.Administrator) bool { return p.Until.IsZero() || p.Until.After(administrator.From) }) {
		return false, nil
	}
	administrator.GrantedBy = ""
	a.periods = append(a.periods, administrator)
	return true, nil
}
