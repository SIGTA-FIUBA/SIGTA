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
	a.Roles = slices.Clone(a.Roles)
	for i, role := range a.Roles {
		a.Roles[i].Revisions = slices.Clone(role.Revisions)
		for j, revision := range role.Revisions {
			a.Roles[i].Revisions[j].Permissions = slices.Clone(revision.Permissions)
		}
	}
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
		r.areas[i] = merge(r.areas[i], cloneArea(a))
		return nil
	}
	r.areas = append(r.areas, cloneArea(a))
	return nil
}

// merge applies to stored the changes of a, which may have been read before
// stored was saved.
func merge(stored, a domain.Area) domain.Area {
	a.CreatedAt, a.CreatedBy = stored.CreatedAt, stored.CreatedBy
	if stored.Deleted() {
		a.DeletedAt, a.DeletedBy = stored.DeletedAt, stored.DeletedBy
	}
	permissions := stored.Permissions
	for _, p := range a.Permissions {
		i := slices.IndexFunc(permissions, func(q domain.Permission) bool { return q.ID == p.ID })
		switch {
		case i < 0:
			permissions = append(permissions, p)
		case permissions[i].Deleted() && !p.System:
			permissions[i].Name, permissions[i].System = p.Name, p.System
		default:
			p.CreatedAt, p.CreatedBy = permissions[i].CreatedAt, permissions[i].CreatedBy
			permissions[i] = p
		}
	}
	roles := stored.Roles
	for _, role := range a.Roles {
		i := slices.IndexFunc(roles, func(q domain.Role) bool { return q.ID == role.ID })
		if i < 0 {
			roles = append(roles, role)
			continue
		}
		revisions := roles[i].Revisions
		for _, revision := range role.Revisions {
			if !slices.ContainsFunc(revisions, func(r domain.RoleRevision) bool { return sameRevision(r, revision) }) {
				revisions = append(revisions, revision)
			}
		}
		role.Revisions = revisions
		if roles[i].Deleted() {
			role.DeletedAt, role.DeletedBy = roles[i].DeletedAt, roles[i].DeletedBy
		}
		roles[i] = role
	}
	a.Permissions, a.Roles = permissions, roles
	return a
}

func sameRevision(a, b domain.RoleRevision) bool {
	return a.At.Equal(b.At) && a.By == b.By && slices.Equal(a.Permissions, b.Permissions)
}
