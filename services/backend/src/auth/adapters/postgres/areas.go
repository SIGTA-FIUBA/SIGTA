package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/src/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var _ output.AreaRepository = (*AreaRepository)(nil)

// AreaRepository saves an area with its permissions, roles and revisions in
// one transaction, and reads it in one snapshot.
type AreaRepository struct {
	pool *pgxpool.Pool
}

func NewAreaRepository(pool *pgxpool.Pool) *AreaRepository {
	return &AreaRepository{pool: pool}
}

func (r *AreaRepository) List(ctx context.Context) ([]domain.Area, error) {
	areas, err := r.read(ctx, "WHERE deleted_at IS NULL")
	if err != nil {
		return nil, fmt.Errorf("list areas: %w", err)
	}
	return areas, nil
}

func (r *AreaRepository) Get(ctx context.Context, id string) (domain.Area, error) {
	areas, err := r.read(ctx, "WHERE id = $1", id)
	switch {
	case err != nil:
		return domain.Area{}, fmt.Errorf("read area %s: %w", id, err)
	case len(areas) == 0:
		return domain.Area{}, fmt.Errorf("%w: %s", domain.ErrAreaNotFound, id)
	}
	return areas[0], nil
}

func (r *AreaRepository) Save(ctx context.Context, a domain.Area) error {
	if err := database.InReadCommitted(ctx, r.pool, func(tx pgx.Tx) error { return save(ctx, tx, a) }); err != nil {
		return fmt.Errorf("save area %s: %w", a.ID, err)
	}
	return nil
}

// save numbers what is new after what is stored, and keeps what is
// deleted: a may have been read before another save.
func save(ctx context.Context, tx pgx.Tx, a domain.Area) error {
	if err := database.Lock(ctx, tx, areasLockKey); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO areas (id, number, name, system, deleted_at, deleted_by)
		SELECT $1, coalesce(max(number), 0) + 1, $2, $3, $4, $5 FROM areas
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, system = excluded.system,
			deleted_at = coalesce(areas.deleted_at, excluded.deleted_at),
			deleted_by = coalesce(areas.deleted_by, excluded.deleted_by)`,
		a.ID, a.Name, a.System, database.NullableTime(a.DeletedAt), database.Nullable(a.DeletedBy)); err != nil {
		return err
	}
	for _, p := range a.Permissions {
		if _, err := tx.Exec(ctx, `INSERT INTO area_permissions (area_id, id, number, name, system, deleted_at, deleted_by)
			SELECT $1, $2, coalesce(max(number), 0) + 1, $3, $4, $5, $6 FROM area_permissions WHERE area_id = $1
			ON CONFLICT (area_id, id) DO UPDATE SET name = excluded.name, system = excluded.system,
				deleted_at = coalesce(area_permissions.deleted_at, excluded.deleted_at),
				deleted_by = coalesce(area_permissions.deleted_by, excluded.deleted_by)`,
			a.ID, p.ID, p.Name, p.System, database.NullableTime(p.DeletedAt), database.Nullable(p.DeletedBy)); err != nil {
			return err
		}
	}
	for _, role := range a.Roles {
		if _, err := tx.Exec(ctx, `INSERT INTO area_roles (area_id, id, number, name, deleted_at, deleted_by)
			SELECT $1, $2, coalesce(max(number), 0) + 1, $3, $4, $5 FROM area_roles WHERE area_id = $1
			ON CONFLICT (area_id, id) DO UPDATE SET name = excluded.name,
				deleted_at = coalesce(area_roles.deleted_at, excluded.deleted_at),
				deleted_by = coalesce(area_roles.deleted_by, excluded.deleted_by)`,
			a.ID, role.ID, role.Name, database.NullableTime(role.DeletedAt), database.Nullable(role.DeletedBy)); err != nil {
			return err
		}
		for _, revision := range role.Revisions {
			if _, err := tx.Exec(ctx, `INSERT INTO area_role_revisions (area_id, role_id, number, permissions, revised_at, revised_by)
				SELECT $1, $2, coalesce(max(number), 0) + 1, $3, $4, $5 FROM area_role_revisions
				WHERE area_id = $1 AND role_id = $2
				HAVING NOT coalesce(bool_or(permissions = $3 AND revised_at = $4 AND revised_by = $5), false)`,
				a.ID, role.ID, database.Array(revision.Permissions), revision.At, revision.By); err != nil {
				return err
			}
		}
	}
	return nil
}

// read returns the areas that where selects, in the order they were added,
// with their permissions, roles and revisions as of the same snapshot.
func (r *AreaRepository) read(ctx context.Context, where string, args ...any) ([]domain.Area, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	rows, err := tx.Query(ctx, "SELECT id, name, system, deleted_at, deleted_by FROM areas "+where+" ORDER BY number", args...)
	if err != nil {
		return nil, err
	}
	areas, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Area, error) {
		var a domain.Area
		var deletedAt *time.Time
		var deletedBy *string
		err := row.Scan(&a.ID, &a.Name, &a.System, &deletedAt, &deletedBy)
		a.DeletedAt, a.DeletedBy = database.TimeOf(deletedAt), database.ValueOf(deletedBy)
		return a, err
	})
	if err != nil || len(areas) == 0 {
		return nil, err
	}
	index := make(map[string]*domain.Area, len(areas))
	ids := make([]string, len(areas))
	for i := range areas {
		index[areas[i].ID] = &areas[i]
		ids[i] = areas[i].ID
	}

	rows, err = tx.Query(ctx, `SELECT area_id, id, name, system, deleted_at, deleted_by FROM area_permissions
		WHERE area_id = ANY($1) ORDER BY area_id, number`, ids)
	if err != nil {
		return nil, err
	}
	var areaID string
	var p domain.Permission
	var deletedAt *time.Time
	var deletedBy *string
	if _, err := pgx.ForEachRow(rows, []any{&areaID, &p.ID, &p.Name, &p.System, &deletedAt, &deletedBy}, func() error {
		p.DeletedAt, p.DeletedBy = database.TimeOf(deletedAt), database.ValueOf(deletedBy)
		index[areaID].Permissions = append(index[areaID].Permissions, p)
		return nil
	}); err != nil {
		return nil, err
	}

	rows, err = tx.Query(ctx, `SELECT area_id, id, name, deleted_at, deleted_by FROM area_roles
		WHERE area_id = ANY($1) ORDER BY area_id, number`, ids)
	if err != nil {
		return nil, err
	}
	var role domain.Role
	if _, err := pgx.ForEachRow(rows, []any{&areaID, &role.ID, &role.Name, &deletedAt, &deletedBy}, func() error {
		role.DeletedAt, role.DeletedBy = database.TimeOf(deletedAt), database.ValueOf(deletedBy)
		index[areaID].Roles = append(index[areaID].Roles, role)
		return nil
	}); err != nil {
		return nil, err
	}

	rows, err = tx.Query(ctx, `SELECT area_id, role_id, permissions, revised_at, revised_by FROM area_role_revisions
		WHERE area_id = ANY($1) ORDER BY area_id, role_id, number`, ids)
	if err != nil {
		return nil, err
	}
	var roleID string
	var revision domain.RoleRevision
	if _, err := pgx.ForEachRow(rows, []any{&areaID, &roleID, &revision.Permissions, &revision.At, &revision.By}, func() error {
		roles := index[areaID].Roles
		for i := range roles {
			if roles[i].ID == roleID {
				roles[i].Revisions = append(roles[i].Revisions, domain.RoleRevision{
					Permissions: database.List(revision.Permissions), At: revision.At.UTC(), By: revision.By,
				})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return nil, err
	}
	return areas, nil
}
