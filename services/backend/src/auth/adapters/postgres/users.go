package postgres

import (
	"context"
	"errors"
	"fmt"

	"backend/src/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var _ output.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Get(ctx context.Context, id string) (domain.User, error) {
	return r.one(ctx, "user "+id, "SELECT id, email, name FROM users WHERE id = $1", id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.one(ctx, "user "+email, "SELECT id, email, name FROM users WHERE email = $1", email)
}

// Register runs at READ COMMITTED, where an insert that finds the email
// taken by a rival waits for it and skips the row, instead of failing as it
// does at SERIALIZABLE. The user is read again after the insert: its
// statement cannot return it, since its snapshot may be older than the rival.
func (r *UserRepository) Register(ctx context.Context, user domain.User) (domain.User, error) {
	user.Email = domain.NormalizeEmail(user.Email)
	var inserted bool
	err := database.InReadCommitted(ctx, r.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, "INSERT INTO users (id, email, name) VALUES ($1, $2, $3) ON CONFLICT (email) DO NOTHING",
			user.ID, user.Email, user.Name)
		inserted = tag.RowsAffected() == 1
		return err
	})
	switch {
	case err != nil:
		return domain.User{}, fmt.Errorf("register user %s: %w", user.Email, err)
	case inserted:
		return user, nil
	}
	return r.GetByEmail(ctx, user.Email)
}

// one reads the user that sql selects, described by what.
func (r *UserRepository) one(ctx context.Context, what, sql string, args ...any) (domain.User, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return domain.User{}, fmt.Errorf("read %s: %w", what, err)
	}
	u, err := pgx.CollectExactlyOneRow(rows, func(row pgx.CollectableRow) (domain.User, error) {
		var u domain.User
		err := row.Scan(&u.ID, &u.Email, &u.Name)
		return u, err
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.User{}, fmt.Errorf("%w: %s", domain.ErrUserNotFound, what)
	case err != nil:
		return domain.User{}, fmt.Errorf("read %s: %w", what, err)
	}
	return u, nil
}
