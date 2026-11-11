package memory

import (
	"context"
	"slices"
	"sync"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var _ output.AreaRepository = (*AreaRepository)(nil)

type AreaRepository struct {
	mu    sync.Mutex
	areas []domain.Area
}

func NewAreaRepository(areas ...domain.Area) *AreaRepository {
	r := &AreaRepository{}
	for _, a := range areas {
		r.areas = append(r.areas, cloneArea(a))
	}
	return r
}

func cloneArea(a domain.Area) domain.Area {
	a.Permissions = slices.Clone(a.Permissions)
	return a
}

func (r *AreaRepository) List(context.Context) ([]domain.Area, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var active []domain.Area
	for _, a := range r.areas {
		if !a.Deleted() {
			active = append(active, cloneArea(a))
		}
	}
	return active, nil
}

func (r *AreaRepository) Get(_ context.Context, id string) (domain.Area, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i := slices.IndexFunc(r.areas, func(a domain.Area) bool { return a.ID == id })
	if i < 0 {
		return domain.Area{}, domain.ErrAreaNotFound
	}
	return cloneArea(r.areas[i]), nil
}

func (r *AreaRepository) Save(_ context.Context, a domain.Area) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i := slices.IndexFunc(r.areas, func(q domain.Area) bool { return q.ID == a.ID }); i >= 0 {
		r.areas[i] = cloneArea(a)
		return nil
	}
	r.areas = append(r.areas, cloneArea(a))
	return nil
}
