package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
	"backend/src/database"
)

var _ output.FirstAdministratorRepository = (*AdministratorRepository)(nil)

// AdministratorRepository needs the role that owns the tables for GrantFirst:
// the application role can only read them.
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
