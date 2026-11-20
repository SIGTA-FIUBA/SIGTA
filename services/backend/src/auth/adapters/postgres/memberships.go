package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

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

const membershipColumns = "id, user_id, area_id, roles, permissions, scope, valid_from, valid_until, granted_by, ended_by"

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
	err := inReadCommitted(ctx, r.pool, func(tx pgx.Tx) error {
		if err := lock(ctx, tx, membershipsLockKey); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO memberships (id, number, user_id, area_id, roles, permissions, scope, valid_from, valid_until, granted_by, ended_by)
			SELECT $1, coalesce(max(number), 0) + 1, $2, $3, $4, $5, $6, $7, $8, $9, $10 FROM memberships
			ON CONFLICT (id) DO UPDATE SET valid_from = excluded.valid_from, valid_until = excluded.valid_until, ended_by = excluded.ended_by`,
			m.ID, m.UserID, m.Area, array(m.Roles), array(m.Permissions), array(m.Scope), m.From, nullableTime(m.Until), m.GrantedBy, nullable(m.EndedBy))
		return err
	})
	if err != nil {
		return fmt.Errorf("save membership %s: %w", m.ID, err)
	}
	return nil
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
		m       domain.Membership
		until   *time.Time
		endedBy *string
	)
	if err := row.Scan(&m.ID, &m.UserID, &m.Area, &m.Roles, &m.Permissions, &m.Scope, &m.From, &until, &m.GrantedBy, &endedBy); err != nil {
		return domain.Membership{}, err
	}
	m.Roles, m.Permissions, m.Scope = list(m.Roles), list(m.Permissions), list(m.Scope)
	m.From, m.Until, m.EndedBy = m.From.UTC(), timeOf(until), valueOf(endedBy)
	return m, nil
}
