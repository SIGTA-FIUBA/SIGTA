package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"backend/src/auth/application"
	"backend/src/auth/domain"
	authoutput "backend/src/auth/ports/output"
)

// usersFileGranter is who grants and ends the memberships of SIGTA_USERS_FILE.
const usersFileGranter = "SIGTA_USERS_FILE"

// addDirectory leaves in the stores the areas of staff and the people who
// can log in, on every start. It returns the ids of the administrators.
func addDirectory(ctx context.Context, cfg config, s stores, now time.Time) ([]string, error) {
	if err := addAreas(ctx, s.areas, devAreas()); err != nil {
		return nil, err
	}
	if err := application.EnsureSecretariat(ctx, s.areas); err != nil {
		return nil, err
	}
	areas, err := s.areas.List(ctx)
	if err != nil {
		return nil, err
	}
	directory, err := peopleOf(cfg, areas)
	if err != nil {
		return nil, err
	}
	return addPeople(ctx, s.users, s.memberships, areas, directory, newID, now)
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

// addPeople registers the people and grants their memberships. Someone who
// signed in before being listed keeps the id they got then, so it returns
// the administrators with the ids the users have. A membership of
// development is added only if its id is not stored; the ones of the users
// file follow the file, as grantFileMemberships says.
func addPeople(ctx context.Context, users authoutput.UserRepository, memberships authoutput.MembershipRepository, areas []domain.Area, p people, newID func() string, now time.Time) ([]string, error) {
	ids := map[string]string{}
	for _, u := range p.users {
		stored, err := users.Get(ctx, u.ID)
		switch {
		case err == nil && stored.Email != domain.NormalizeEmail(u.Email):
			return nil, fmt.Errorf("user %s is stored with %s, not %s: the users file cannot change an email", u.ID, stored.Email, u.Email)
		case err != nil && !errors.Is(err, domain.ErrUserNotFound):
			return nil, err
		}
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
	var fromFile []domain.Membership
	for _, m := range p.memberships {
		m.UserID, m.From = registeredID(m.UserID), now
		if m.GrantedBy == usersFileGranter {
			fromFile = append(fromFile, m)
			continue
		}
		_, err := memberships.Get(ctx, m.ID)
		if errors.Is(err, domain.ErrMembershipNotFound) {
			err = memberships.Save(ctx, m)
		}
		if err != nil {
			return nil, fmt.Errorf("membership %s: %w", m.ID, err)
		}
	}
	if err := grantFileMemberships(ctx, memberships, areas, fromFile, newID, now); err != nil {
		return nil, err
	}
	return administrators, nil
}

// grantFileMemberships leaves active every membership the users file lists,
// and ends the ones it granted and no longer lists, so that taking someone
// out of the file takes their permissions away on the next start. The file
// prevails over the administrators: a membership of the file that one of
// them ended is granted again.
func grantFileMemberships(ctx context.Context, memberships authoutput.MembershipRepository, areas []domain.Area, listed []domain.Membership, newID func() string, now time.Time) error {
	for _, a := range areas {
		granted, err := memberships.ListByArea(ctx, a.ID)
		if err != nil {
			return err
		}
		for _, m := range granted {
			if m.GrantedBy != usersFileGranter || !m.ActiveAt(now) {
				continue
			}
			if i := slices.IndexFunc(listed, func(l domain.Membership) bool { return sameGrant(l, m) }); i >= 0 {
				listed = slices.Delete(listed, i, i+1)
				continue
			}
			m.Until, m.EndedBy = now, usersFileGranter
			if err := memberships.Save(ctx, m); err != nil {
				return fmt.Errorf("end membership %s: %w", m.ID, err)
			}
		}
	}
	for _, m := range listed {
		m.ID = newID()
		if err := memberships.Save(ctx, m); err != nil {
			return fmt.Errorf("grant membership of %s in %s: %w", m.UserID, m.Area, err)
		}
	}
	return nil
}

func sameGrant(a, b domain.Membership) bool {
	return a.UserID == b.UserID && a.Area == b.Area && sameSet(a.Roles, b.Roles) &&
		sameSet(a.Permissions, b.Permissions) && sameSet(a.Scope, b.Scope)
}

func sameSet[T cmp.Ordered](a, b []T) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}
