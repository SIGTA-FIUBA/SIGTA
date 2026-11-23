package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
	"backend/src/database"
)

var (
	_ output.FirstAdministratorRepository  = (*AdministratorRepository)(nil)
	_ output.AdministratorPeriodRepository = (*AdministratorRepository)(nil)
)

// AdministratorRepository needs the role that owns the tables for GrantFirst:
// the application role can only add and end periods granted by an
// administrator.
type AdministratorRepository struct {
	pool *pgxpool.Pool
}

func NewAdministratorRepository(pool *pgxpool.Pool) *AdministratorRepository {
	return &AdministratorRepository{pool: pool}
}

func (a *AdministratorRepository) IsAdministrator(ctx context.Context, userID string) (bool, error) {
	var administrator bool
	err := a.pool.QueryRow(ctx, "SELECT EXISTS (SELECT FROM administrators WHERE user_id = $1 AND valid_from <= now() AND (valid_until IS NULL OR valid_until > now()))", userID).Scan(&administrator)
	if err != nil {
		return false, fmt.Errorf("read whether %s is an administrator: %w", userID, err)
	}
	return administrator, nil
}

// GrantFirst takes a lock so that two at once cannot both find no
// administrator.
func (a *AdministratorRepository) GrantFirst(ctx context.Context, administrator domain.Administrator) (bool, error) {
	var granted bool
	err := database.InReadCommitted(ctx, a.pool, func(tx pgx.Tx) error {
		if err := database.Lock(ctx, tx, administratorsLockKey); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `INSERT INTO administrators (id, user_id, valid_from, grant_source)
			SELECT $1, $2, $3, 'bootstrap' WHERE NOT EXISTS (SELECT FROM administrators WHERE valid_until IS NULL OR valid_until > $3)`,
			administrator.ID, administrator.UserID, administrator.From)
		granted = tag.RowsAffected() == 1
		return err
	})
	if err != nil {
		return false, fmt.Errorf("make %s the first administrator: %w", administrator.UserID, err)
	}
	return granted, nil
}

func (a *AdministratorRepository) ListActive(ctx context.Context, at time.Time) ([]domain.Administrator, error) {
	rows, err := a.pool.Query(ctx, `SELECT id, user_id, valid_from, coalesce(granted_by, '') FROM administrators
		WHERE valid_from <= $1 AND (valid_until IS NULL OR valid_until > $1) ORDER BY valid_from, id`, at)
	if err != nil {
		return nil, fmt.Errorf("list the administrators: %w", err)
	}
	administrators, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Administrator, error) {
		var p domain.Administrator
		err := row.Scan(&p.ID, &p.UserID, &p.From, &p.GrantedBy)
		p.From = p.From.UTC()
		return p, err
	})
	if err != nil {
		return nil, fmt.Errorf("list the administrators: %w", err)
	}
	return administrators, nil
}

func (a *AdministratorRepository) Grant(ctx context.Context, administrator domain.Administrator) error {
	_, err := a.pool.Exec(ctx, "INSERT INTO administrators (id, user_id, valid_from, grant_source, granted_by) VALUES ($1, $2, $3, 'administrator', $4)",
		administrator.ID, administrator.UserID, administrator.From, administrator.GrantedBy)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "administrators_without_end" {
		return fmt.Errorf("%w: %s", domain.ErrAlreadyAdministrator, administrator.UserID)
	}
	if err != nil {
		return fmt.Errorf("grant administrator to %s: %w", administrator.UserID, err)
	}
	return nil
}

// End takes the same lock as GrantFirst, so that two ending each other at
// once cannot both find the other active.
func (a *AdministratorRepository) End(ctx context.Context, id string, at time.Time, by string) error {
	err := database.InReadCommitted(ctx, a.pool, func(tx pgx.Tx) error {
		if err := database.Lock(ctx, tx, administratorsLockKey); err != nil {
			return err
		}
		var found, others bool
		err := tx.QueryRow(ctx, `SELECT
				EXISTS (SELECT FROM administrators WHERE id = $1 AND valid_from <= $2 AND (valid_until IS NULL OR valid_until > $2)),
				EXISTS (SELECT FROM administrators WHERE id <> $1 AND valid_from <= $2 AND (valid_until IS NULL OR valid_until > $2))`, id, at).Scan(&found, &others)
		switch {
		case err != nil:
			return err
		case !found:
			return fmt.Errorf("%w: %s", domain.ErrAdministratorNotFound, id)
		case !others:
			return fmt.Errorf("%w: %s is the only one", domain.ErrLastAdministrator, id)
		}
		_, err = tx.Exec(ctx, "UPDATE administrators SET valid_until = $2, ended_by = $3 WHERE id = $1", id, at, by)
		return err
	})
	if err != nil {
		return fmt.Errorf("end administrator %s: %w", id, err)
	}
	return nil
}
