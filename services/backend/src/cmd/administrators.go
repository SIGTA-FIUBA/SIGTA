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
	"backend/src/auth/domain"
	authoutput "backend/src/auth/ports/output"
	"backend/src/migrations"
)

// storedOrListedAdministrators answers for the administrators that the
// stores keep and for the synthetic one of development, which is not stored:
// it is listed first, and it cannot be ended or counted to end the others.
type storedOrListedAdministrators struct {
	authoutput.AdministratorPeriodRepository
	listed authoutput.AdministratorPeriodRepository
}

func (a storedOrListedAdministrators) IsAdministrator(ctx context.Context, userID string) (bool, error) {
	if listed, err := a.listed.IsAdministrator(ctx, userID); err != nil || listed {
		return listed, err
	}
	return a.AdministratorPeriodRepository.IsAdministrator(ctx, userID)
}

func (a storedOrListedAdministrators) ListActive(ctx context.Context, at time.Time) ([]domain.Administrator, error) {
	listed, err := a.listed.ListActive(ctx, at)
	if err != nil {
		return nil, err
	}
	stored, err := a.AdministratorPeriodRepository.ListActive(ctx, at)
	if err != nil {
		return nil, err
	}
	return append(listed, stored...), nil
}

// grantAdminCommand makes the person who signed in with the email of args the
// first administrator of an environment, as the initial grant of the system.
// Like migrate, it connects as the role that owns the tables, which the server
// never has, so it runs as a job apart, and keeping who ran it is up to the
// environment where it is deployed. Once there is an administrator it fails:
// administrators grant the others in SIGTA, which keeps who did it.
func grantAdminCommand(getenv func(string) string, args []string, logger *slog.Logger) error {
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" || strings.HasPrefix(args[0], "-") {
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
	logger.Info("first administrator granted", "user", user.ID, "email", user.Email)
	return nil
}
