package postgres

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"backend/src/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var _ output.MembershipRepository = (*MembershipRepository)(nil)

type MembershipRepository struct {
	pool *pgxpool.Pool
}

func NewMembershipRepository(pool *pgxpool.Pool) *MembershipRepository {
	return &MembershipRepository{pool: pool}
}

const membershipColumns = "id, user_id, area_id, roles, permissions, scope, valid_from, valid_until, granted_by, granted_at, ended_by"

func (r *MembershipRepository) ListByUser(ctx context.Context, userID string) ([]domain.Membership, error) {
	return r.list(ctx, "memberships of "+userID, "WHERE user_id = $1", userID)
}

func (r *MembershipRepository) ListByArea(ctx context.Context, areaID string) ([]domain.Membership, error) {
	return r.list(ctx, "memberships of area "+areaID, "WHERE area_id = $1", areaID)
}

func (r *MembershipRepository) Get(ctx context.Context, id string) (domain.Membership, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+membershipColumns+" FROM memberships WHERE id = $1", id)
	if err != nil {
		return domain.Membership{}, fmt.Errorf("read membership %s: %w", id, err)
	}
	m, err := pgx.CollectExactlyOneRow(rows, scanMembership)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.Membership{}, fmt.Errorf("%w: %s", domain.ErrMembershipNotFound, id)
	case err != nil:
		return domain.Membership{}, fmt.Errorf("read membership %s: %w", id, err)
	}
	return m, nil
}

func (r *MembershipRepository) Save(ctx context.Context, m domain.Membership) error {
	if err := database.InReadCommitted(ctx, r.pool, func(tx pgx.Tx) error { return saveMembership(ctx, tx, m) }); err != nil {
		return fmt.Errorf("save membership %s: %w", m.ID, err)
	}
	return nil
}

func (r *MembershipRepository) Replace(ctx context.Context, ended, granted domain.Membership) error {
	err := database.InReadCommitted(ctx, r.pool, func(tx pgx.Tx) error {
		if err := saveMembership(ctx, tx, ended); err != nil {
			return err
		}
		return saveMembership(ctx, tx, granted)
	})
	if err != nil {
		return fmt.Errorf("replace membership %s with %s: %w", ended.ID, granted.ID, err)
	}
	return nil
}

// saveMembership checks, under the locks that every save takes, what Save
// says, so that two changes at once cannot both pass it. It takes the lock of
// the areas first, as the areas do, so that an area, permission or role
// removed meanwhile is not granted.
func saveMembership(ctx context.Context, tx pgx.Tx, m domain.Membership) error {
	if err := database.Lock(ctx, tx, areasLockKey); err != nil {
		return err
	}
	if err := database.Lock(ctx, tx, membershipsLockKey); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, "SELECT "+membershipColumns+" FROM memberships WHERE user_id = $1 OR id = $2", m.UserID, m.ID)
	if err != nil {
		return err
	}
	memberships, err := pgx.CollectRows(rows, scanMembership)
	if err != nil {
		return err
	}
	if i := slices.IndexFunc(memberships, func(q domain.Membership) bool { return q.ID == m.ID }); i >= 0 {
		if !memberships[i].EndsAfter(m.Until) {
			return fmt.Errorf("%w: %s", domain.ErrMembershipEnded, m.ID)
		}
		_, err := tx.Exec(ctx, "UPDATE memberships SET valid_from = $2, valid_until = $3, ended_by = $4 WHERE id = $1",
			m.ID, m.From, database.NullableTime(m.Until), database.Nullable(m.EndedBy))
		return err
	}
	if m.OverlapsAny(memberships) {
		return fmt.Errorf("%w: %s in area %s", domain.ErrMembershipOverlaps, m.UserID, m.Area)
	}
	if err := checkGrantable(ctx, tx, m); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO memberships (id, number, user_id, area_id, roles, permissions, scope, valid_from, valid_until, granted_by, granted_at, ended_by)
		SELECT $1, coalesce(max(number), 0) + 1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 FROM memberships`,
		m.ID, m.UserID, m.Area, database.Array(m.Roles), database.Array(m.Permissions), database.Array(m.Scope), m.From, database.NullableTime(m.Until), m.GrantedBy, database.NullableTime(m.GrantedAt), database.Nullable(m.EndedBy))
	return err
}

func (r *MembershipRepository) list(ctx context.Context, what, where string, args ...any) ([]domain.Membership, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+membershipColumns+" FROM memberships "+where+" ORDER BY number", args...)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", what, err)
	}
	memberships, err := pgx.CollectRows(rows, scanMembership)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", what, err)
	}
	return memberships, nil
}

func scanMembership(row pgx.CollectableRow) (domain.Membership, error) {
	var (
		m         domain.Membership
		until     *time.Time
		grantedAt *time.Time
		endedBy   *string
	)
	if err := row.Scan(&m.ID, &m.UserID, &m.Area, &m.Roles, &m.Permissions, &m.Scope, &m.From, &until, &m.GrantedBy, &grantedAt, &endedBy); err != nil {
		return domain.Membership{}, err
	}
	m.Roles, m.Permissions, m.Scope = database.List(m.Roles), database.List(m.Permissions), database.List(m.Scope)
	m.From, m.Until, m.GrantedAt, m.EndedBy = m.From.UTC(), database.TimeOf(until), database.TimeOf(grantedAt), database.ValueOf(endedBy)
	return m, nil
}

// checkGrantable checks that the area of m and its roles and permissions have
// not been removed.
func checkGrantable(ctx context.Context, tx pgx.Tx, m domain.Membership) error {
	var area, permissions, roles bool
	err := tx.QueryRow(ctx, `SELECT
			EXISTS (SELECT FROM areas WHERE id = $1 AND deleted_at IS NULL),
			NOT EXISTS (SELECT FROM unnest($2::text[]) AS p (id) WHERE NOT EXISTS (
				SELECT FROM area_permissions WHERE area_id = $1 AND id = p.id AND deleted_at IS NULL)),
			NOT EXISTS (SELECT FROM unnest($3::text[]) AS r (id) WHERE NOT EXISTS (
				SELECT FROM area_roles WHERE area_id = $1 AND id = r.id AND deleted_at IS NULL))`,
		m.Area, database.Array(m.Permissions), database.Array(m.Roles)).Scan(&area, &permissions, &roles)
	switch {
	case err != nil:
		return err
	case !area:
		return fmt.Errorf("%w: %w: %s", domain.ErrInvalidMembership, domain.ErrAreaNotFound, m.Area)
	case !permissions:
		return fmt.Errorf("%w: %w: in area %s", domain.ErrInvalidMembership, domain.ErrPermissionNotFound, m.Area)
	case !roles:
		return fmt.Errorf("%w: %w: in area %s", domain.ErrInvalidMembership, domain.ErrRoleNotFound, m.Area)
	}
	return nil
}
