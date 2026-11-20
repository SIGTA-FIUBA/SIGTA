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
	"backend/src/migrations"
)

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
