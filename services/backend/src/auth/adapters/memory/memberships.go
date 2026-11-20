package memory

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var _ output.MembershipRepository = (*MembershipRepository)(nil)

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
	if err := checkSave(m, r.memberships); err != nil {
		return err
	}
	r.memberships = saved(r.memberships, m)
	return nil
}

func (r *MembershipRepository) Replace(_ context.Context, ended, granted domain.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := checkSave(ended, r.memberships); err != nil {
		return err
	}
	after := saved(slices.Clone(r.memberships), ended)
	if err := checkSave(granted, after); err != nil {
		return err
	}
	r.memberships = saved(after, granted)
	return nil
}

// checkSave tells whether m can be saved over memberships, as Save says.
func checkSave(m domain.Membership, memberships []domain.Membership) error {
	if i := slices.IndexFunc(memberships, func(q domain.Membership) bool { return q.ID == m.ID }); i >= 0 {
		if !memberships[i].EndsAfter(m.Until) {
			return fmt.Errorf("%w: %s", domain.ErrMembershipEnded, m.ID)
		}
		return nil
	}
	if m.OverlapsAny(memberships) {
		return fmt.Errorf("%w: %s in area %s", domain.ErrMembershipOverlaps, m.UserID, m.Area)
	}
	return nil
}

// saved returns memberships with m saved, changing only the validity of an
// existing one. The memberships it returns are not shared with m.
func saved(memberships []domain.Membership, m domain.Membership) []domain.Membership {
	m = cloneMembership(m)
	if i := slices.IndexFunc(memberships, func(q domain.Membership) bool { return q.ID == m.ID }); i >= 0 {
		stored := &memberships[i]
		stored.From, stored.Until, stored.EndedBy = m.From, m.Until, m.EndedBy
		return memberships
	}
	return append(memberships, m)
}

func cloneMembership(m domain.Membership) domain.Membership {
	m.Roles = slices.Clone(m.Roles)
	m.Permissions = slices.Clone(m.Permissions)
	m.Scope = slices.Clone(m.Scope)
	return m
}
