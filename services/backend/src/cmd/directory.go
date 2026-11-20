package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	authmem "backend/src/auth/adapters/memory"
	"backend/src/auth/application"
	"backend/src/auth/domain"
	authoutput "backend/src/auth/ports/output"
)

// people are the synthetic accounts that development adds on every start.
type people struct {
	users          []domain.User
	administrators []string
	memberships    []domain.Membership
}

// addDirectory leaves the Secretariat in the stores on every start and, in
// development, the example areas and the synthetic people. It returns the
// administrators: the stored ones and, when anyone can log in as anyone
// (the dev identity provider), the synthetic one. Outside development
// everyone registers by signing in, and memberships are granted only in SIGTA.
func addDirectory(ctx context.Context, cfg config, s stores, now time.Time) (authoutput.AdministratorRepository, error) {
	if err := application.EnsureSecretariat(ctx, s.areas); err != nil {
		return nil, err
	}
	if !cfg.development() {
		return s.administrators, nil
	}
	if err := addAreas(ctx, s.areas, devAreas()); err != nil {
		return nil, err
	}
	listed, err := addPeople(ctx, s.users, s.memberships, devPeople(), now)
	if err != nil {
		return nil, err
	}
	if cfg.identityProvider != identityDev {
		return s.administrators, nil
	}
	return storedOrListedAdministrators{stored: s.administrators, listed: authmem.NewAdministratorRepository(listed...)}, nil
}

// addAreas adds the areas that are not stored yet. One that is keeps what
// the administrators changed in it.
func addAreas(ctx context.Context, repository authoutput.AreaRepository, areas []domain.Area) error {
	for _, a := range areas {
		_, err := repository.Get(ctx, a.ID)
		if errors.Is(err, domain.ErrAreaNotFound) {
			err = repository.Save(ctx, a)
		}
		if err != nil {
			return fmt.Errorf("area %s: %w", a.ID, err)
		}
	}
	return nil
}

// addPeople registers the people and adds each membership whose id is not
// stored, so one that an administrator ended stays ended. Someone who signed
// in before keeps the id they got then, so it returns the administrators
// with the ids the users have.
func addPeople(ctx context.Context, users authoutput.UserRepository, memberships authoutput.MembershipRepository, p people, now time.Time) ([]string, error) {
	ids := map[string]string{}
	for _, u := range p.users {
		registered, err := users.Register(ctx, u)
		if err != nil {
			return nil, err
		}
		ids[u.ID] = registered.ID
	}
	registeredID := func(id string) string {
		if registered, ok := ids[id]; ok {
			return registered
		}
		return id
	}
	administrators := make([]string, 0, len(p.administrators))
	for _, id := range p.administrators {
		administrators = append(administrators, registeredID(id))
	}
	for _, m := range p.memberships {
		m.UserID, m.From = registeredID(m.UserID), now
		_, err := memberships.Get(ctx, m.ID)
		if errors.Is(err, domain.ErrMembershipNotFound) {
			err = memberships.Save(ctx, m)
		}
		if err != nil {
			return nil, fmt.Errorf("membership %s: %w", m.ID, err)
		}
	}
	return administrators, nil
}
