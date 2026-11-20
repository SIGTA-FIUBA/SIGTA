package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	authpostgres "backend/src/auth/adapters/postgres"
	"backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	authoutput "backend/src/auth/ports/output"
	"backend/src/migrations"
)

// storedOrListedAdministrators answers for the administrators that the
// database keeps and for those that the users file or development lists,
// which are not stored.
type storedOrListedAdministrators struct {
	stored, listed authoutput.AdministratorRepository
}

func (a storedOrListedAdministrators) IsAdministrator(ctx context.Context, userID string) (bool, error) {
	if listed, err := a.listed.IsAdministrator(ctx, userID); err != nil || listed {
		return listed, err
	}
	return a.stored.IsAdministrator(ctx, userID)
}

// GrantFirst looks only at the stored administrators: the listed ones come
// from the configuration.
func (a storedOrListedAdministrators) GrantFirst(ctx context.Context, administrator authdomain.Administrator) (bool, error) {
	return a.stored.GrantFirst(ctx, administrator)
}

// grantAdminCommand makes the person who signed in with the email of args the
// first administrator of an environment, as the initial grant of the system.
// Like migrate, it connects as the role that owns the tables, which the server
// never has, so it runs as a job apart, and who ran it is in the audit logs
// of the infrastructure. Once there is an administrator it fails:
// administrators grant the others in SIGTA, which keeps who did it.
func grantAdminCommand(getenv func(string) string, args []string, logger *slog.Logger) error {
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return errors.New("usage: backend grant-admin <email>")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	pool, err := connectAsOwner(ctx, getenv)
	if err != nil {
		return err
	}
	defer pool.Close()
	err = migrations.Check(ctx, pool)
	if errors.Is(err, migrations.ErrPending) {
		return fmt.Errorf("%w: run the migrate command first", err)
	}
	if err != nil {
		return fmt.Errorf("check the migrations: %w", err)
	}
	user, err := application.GrantFirstAdministrator(ctx, authpostgres.NewUserRepository(pool), authpostgres.NewAdministratorRepository(pool), args[0], newID, time.Now())
	if err != nil {
		return err
	}
	logger.Info("first administrator granted", "user", user.ID)
	return nil
}
