package memory

import (
	"context"
	"slices"
	"sync"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var (
	_ output.Administrators       = (*Administrators)(nil)
	_ output.MembershipRepository = (*MembershipRepository)(nil)
	_ output.DeniedAttemptLog     = (*DeniedAttemptLog)(nil)
)

type Administrators struct {
	ids map[string]bool
}

func NewAdministrators(ids ...string) *Administrators {
	a := &Administrators{ids: map[string]bool{}}
	for _, id := range ids {
		a.ids[id] = true
	}
	return a
}

func (a *Administrators) IsAdministrator(_ context.Context, userID string) (bool, error) {
	return a.ids[userID], nil
}

type MembershipRepository struct {
	mu          sync.Mutex
	memberships []domain.Membership
}

func NewMembershipRepository(memberships ...domain.Membership) *MembershipRepository {
	r := &MembershipRepository{}
	for _, m := range memberships {
		r.memberships = append(r.memberships, cloneMembership(m))
	}
	return r
}

func (r *MembershipRepository) ListByUser(_ context.Context, userID string) ([]domain.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var found []domain.Membership
	for _, m := range r.memberships {
		if m.UserID == userID {
			found = append(found, cloneMembership(m))
		}
	}
	return found, nil
}

func (r *MembershipRepository) ListByArea(_ context.Context, areaID string) ([]domain.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var found []domain.Membership
	for _, m := range r.memberships {
		if m.Area == areaID {
			found = append(found, cloneMembership(m))
		}
	}
	return found, nil
}

func (r *MembershipRepository) Get(_ context.Context, id string) (domain.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i := slices.IndexFunc(r.memberships, func(m domain.Membership) bool { return m.ID == id })
	if i < 0 {
		return domain.Membership{}, domain.ErrMembershipNotFound
	}
	return cloneMembership(r.memberships[i]), nil
}

func (r *MembershipRepository) Save(_ context.Context, m domain.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m = cloneMembership(m)
	if i := slices.IndexFunc(r.memberships, func(q domain.Membership) bool { return q.ID == m.ID }); i >= 0 {
		r.memberships[i] = m
		return nil
	}
	r.memberships = append(r.memberships, m)
	return nil
}

func cloneMembership(m domain.Membership) domain.Membership {
	m.Scope = slices.Clone(m.Scope)
	return m
}

type DeniedAttemptLog struct {
	mu       sync.Mutex
	attempts []domain.DeniedAttempt
}

func NewDeniedAttemptLog() *DeniedAttemptLog {
	return &DeniedAttemptLog{}
}

func (l *DeniedAttemptLog) Record(_ context.Context, a domain.DeniedAttempt) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.attempts = append(l.attempts, a)
	return nil
}

func (l *DeniedAttemptLog) All() []domain.DeniedAttempt {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.attempts)
}
