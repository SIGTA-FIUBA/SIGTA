package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/migrations"
	"backend/src/procedure/adapters/disk"
	proceduremem "backend/src/procedure/adapters/memory"
	"backend/src/procedure/adapters/postgres"
	"backend/src/procedure/ports/output"
)

// procedureStores are where the published versions and the drafts of the
// procedure types are kept.
type procedureStores struct {
	versions output.ProcedureTypeVersionRepository
	drafts   output.ProcedureTypeDraftRepository
}

// openProcedureStores returns the stores and a function that closes them.
func openProcedureStores(ctx context.Context, cfg config, logger *slog.Logger) (procedureStores, func(), error) {
	if cfg.databaseURL == "" {
		// loadConfig asks for a database everywhere but development.
		logger.Warn("SIGTA_DATABASE_URL is not set: the procedure types are lost when the server stops")
		return procedureStores{versions: proceduremem.NewProcedureTypeVersionRepository(), drafts: proceduremem.NewProcedureTypeDraftRepository()}, func() {}, nil
	}
	// The pool outlives ctx, which only bounds the checks.
	pool, err := connect(context.WithoutCancel(ctx), cfg.databaseURL, appRole)
	if err != nil {
		return procedureStores{}, nil, fmt.Errorf("SIGTA_DATABASE_URL: %w", err)
	}
	if err := checkDatabase(ctx, pool); err != nil {
		pool.Close()
		return procedureStores{}, nil, err
	}
	return procedureStores{versions: postgres.NewProcedureTypeVersionRepository(pool), drafts: postgres.NewProcedureTypeDraftRepository(pool)}, pool.Close, nil
}

// openFiles returns where the files of the procedures are kept.
func openFiles(cfg config, logger *slog.Logger) (output.Files, error) {
	dir := cfg.documentsDir
	if dir == "" {
		// loadConfig asks for a directory everywhere but development. A new
		// one each time, since the records of the documents are in memory and
		// another user of the machine could have created a fixed one.
		var err error
		if dir, err = os.MkdirTemp("", "sigta-documents-"); err != nil {
			return nil, fmt.Errorf("documents directory: %w", err)
		}
		logger.Warn("SIGTA_DOCUMENTS_DIR is not set: the documents go to a temporary directory", "dir", dir)
	}
	return disk.NewFiles(dir)
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
	err = migrations.CheckApplicationRole(ctx, pool)
	if errors.Is(err, migrations.ErrPrivileged) {
		return fmt.Errorf("SIGTA_DATABASE_URL: %w; the server connects as the application role (backend) and only the migrate command uses the role that owns the tables (backend_migrations)", err)
	}
	return err
}

// The roles and the schema of deploy/postgres/sql/roles.sql.
const (
	appRole        = "backend"
	migrationsRole = "backend_migrations"
	backendSchema  = "backend"
)

// connect opens a pool that acts as role, in the backend schema unless the URL
// names another search_path.
func connect(ctx context.Context, url, role string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	actAs(poolConfig, role)
	return pgxpool.NewWithConfig(ctx, poolConfig)
}

// actAs makes every connection take role. The login does not have to be role
// itself: it can be a login that is a member of it. Such a login does not get
// the search_path that roles.sql sets on role, and what it creates would
// belong to it and not to role, which has to own the tables.
func actAs(config *pgxpool.Config, role string) {
	if _, ok := config.ConnConfig.RuntimeParams["search_path"]; !ok {
		config.ConnConfig.RuntimeParams["search_path"] = backendSchema
	}
	previous := config.AfterConnect
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if previous != nil {
			if err := previous(ctx, conn); err != nil {
				return err
			}
		}
		_, err := conn.Exec(ctx, "SET ROLE "+pgx.Identifier{role}.Sanitize())
		return err
	}
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
	pool, err := connect(ctx, url, migrationsRole)
	if err != nil {
		return fmt.Errorf("SIGTA_MIGRATIONS_DATABASE_URL: %w", err)
	}
	defer pool.Close()
	applied, err := applyMigrations(ctx, pool, getenv("SIGTA_ENV") == envDevelopment)
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
// or does not let it create there. Only development gets the hint to
// recreate the database: elsewhere the same codes come from a wrong password
// or a missing grant, and recreating it would lose the data.
const (
	invalidAuthorization  = "28000"
	invalidPassword       = "28P01"
	invalidSchemaName     = "3F000"
	insufficientPrivilege = "42501"
)

func applyMigrations(ctx context.Context, pool *pgxpool.Pool, development bool) (int, error) {
	applied, err := migrations.Apply(ctx, pool)
	if err == nil {
		return applied, nil
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && development {
		switch pgErr.Code {
		case invalidAuthorization, invalidPassword, invalidSchemaName, insufficientPrivilege:
			return 0, fmt.Errorf("apply the migrations: %w (%s)", err, migrations.OldDatabaseHint)
		}
	}
	return 0, fmt.Errorf("apply the migrations: %w", err)
}
