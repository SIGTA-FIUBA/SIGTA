package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/migrations"
	proceduremem "backend/src/procedure/adapters/memory"
	"backend/src/procedure/adapters/postgres"
	"backend/src/procedure/ports/output"
)

// openProcedureVersions returns where the published versions are kept and a
// function that closes it.
func openProcedureVersions(ctx context.Context, cfg config, logger *slog.Logger) (output.ProcedureVersions, func(), error) {
	if cfg.databaseURL == "" {
		// loadConfig asks for a database everywhere but development.
		logger.Warn("SIGTA_DATABASE_URL is not set: the procedure types are lost when the server stops")
		return proceduremem.NewProcedureVersions(), func() {}, nil
	}
	// The pool outlives ctx, which only bounds the checks.
	pool, err := pgxpool.New(context.WithoutCancel(ctx), cfg.databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("SIGTA_DATABASE_URL: %w", err)
	}
	if err := checkDatabase(ctx, pool); err != nil {
		pool.Close()
		return nil, nil, err
	}
	return postgres.NewProcedureVersions(pool), pool.Close, nil
}

// checkDatabase refuses a schema that is not migrated, and a role that can
// change the published versions. The server connects as the application
// role, which cannot migrate: that is the job of the migrate command.
func checkDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	err := migrations.Check(ctx, pool)
	if errors.Is(err, migrations.ErrPending) {
		return fmt.Errorf("%w: run the migrate command first", err)
	}
	if err != nil {
		return fmt.Errorf("check the migrations: %w", err)
	}
	return migrations.CheckApplicationRole(ctx, pool)
}

// migrateCommand applies the migrations as the role that owns the tables. It
// runs apart from the server, so the process that serves requests never has
// that credential.
func migrateCommand(getenv func(string) string, logger *slog.Logger) error {
	url := getenv("SIGTA_MIGRATIONS_DATABASE_URL")
	if url == "" {
		return errors.New("missing SIGTA_MIGRATIONS_DATABASE_URL")
	}
	// Each lock a migration waits for is bounded apart, in migrations.Apply.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return fmt.Errorf("SIGTA_MIGRATIONS_DATABASE_URL: %w", err)
	}
	defer pool.Close()
	applied, err := applyMigrations(ctx, pool)
	if err != nil {
		return err
	}
	if applied == 0 {
		logger.Info("schema up to date")
	} else {
		logger.Info("migrations applied", "count", applied)
	}
	return nil
}

// SQLSTATE codes that a database created before the migrations role existed
// gives to the migrate command: the role does not exist or cannot log in, or
// the schema belongs to another role, which hides it from the migrations role
// or does not let it create there.
const (
	invalidAuthorization  = "28000"
	invalidPassword       = "28P01"
	invalidSchemaName     = "3F000"
	insufficientPrivilege = "42501"
)

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	applied, err := migrations.Apply(ctx, pool)
	if err == nil {
		return applied, nil
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case invalidAuthorization, invalidPassword, invalidSchemaName, insufficientPrivilege:
			return 0, fmt.Errorf("apply the migrations: %w (%s)", err, migrations.OldDatabaseHint)
		}
	}
	return 0, fmt.Errorf("apply the migrations: %w", err)
}
