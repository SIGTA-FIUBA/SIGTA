// Package postgres stores the auth module in PostgreSQL, with the tables of
// src/migrations and the privileges of the application role.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

const (
	serializationFailure = "40001"

	// attempts bounds how many times a write is tried after a serialization
	// failure.
	attempts = 3
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

// Register reads the user again when the insert finds the email taken: the
// statement cannot return it, since its snapshot may be older than the rival
// insert. At SERIALIZABLE that rival fails the insert with a serialization
// failure instead, and Register tries again.
func (r *UserRepository) Register(ctx context.Context, user domain.User) (domain.User, error) {
	user.Email = domain.NormalizeEmail(user.Email)
	for attempt := 1; ; attempt++ {
		tag, err := r.pool.Exec(ctx, "INSERT INTO users (id, email, name) VALUES ($1, $2, $3) ON CONFLICT (email) DO NOTHING",
			user.ID, user.Email, user.Name)
		switch {
		case hasCode(err, serializationFailure) && attempt < attempts:
			continue
		case err != nil:
			return domain.User{}, fmt.Errorf("register user %s: %w", user.Email, err)
		case tag.RowsAffected() == 1:
			return user, nil
		}
		return r.GetByEmail(ctx, user.Email)
	}
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

func hasCode(err error, code string) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == code
}
