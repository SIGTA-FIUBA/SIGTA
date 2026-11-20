package main

import (
	"slices"
	"testing"
	"time"

	authmem "backend/src/auth/adapters/memory"
	"backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
)

// Someone who signed in before keeps the id they got then, and their
// membership and administration go to that id.
func TestAddPeopleKeepsTheIDOfWhoSignedInBefore(t *testing.T) {
	ctx := t.Context()
	users, memberships := authmem.NewUserRepository(), authmem.NewMembershipRepository()
	if _, err := users.Register(ctx, domain.User{ID: "signed-in", Email: "ana@fi.uba.ar", Name: "Ana"}); err != nil {
		t.Fatal(err)
	}
	p := people{
		users:          []domain.User{{ID: "ana", Email: "ana@fi.uba.ar", Name: "Ana"}},
		administrators: []string{"ana"},
		memberships:    []domain.Membership{{ID: "ana-1", UserID: "ana", Area: "office", Permissions: []string{"review"}, GrantedBy: "dev-admin"}},
	}
	administrators, err := addPeople(ctx, users, memberships, p, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(administrators, []string{"signed-in"}) {
		t.Errorf("administrators = %v, want [signed-in]", administrators)
	}
	if got, err := memberships.ListByUser(ctx, "signed-in"); err != nil || len(got) != 1 {
		t.Errorf("memberships of signed-in = %+v, %v; want ana-1", got, err)
	}
}

// The memberships of development keep their ids and are added once.
func TestAddPeopleAddsTheDevelopmentMembershipsOnce(t *testing.T) {
	ctx := t.Context()
	users, memberships := authmem.NewUserRepository(), authmem.NewMembershipRepository()
	for range 2 {
		if _, err := addPeople(ctx, users, memberships, devPeople(), time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	got, err := memberships.ListByUser(ctx, "dev-operative")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "dev-operative-1" || got[1].ID != "dev-operative-2" {
		t.Errorf("memberships of dev-operative = %+v, want dev-operative-1 and dev-operative-2", got)
	}
}

// A membership of development that an administrator ended is not granted
// again on the next start.
func TestAddPeopleKeepsWhatAnAdministratorEnded(t *testing.T) {
	ctx := t.Context()
	users, memberships := authmem.NewUserRepository(), authmem.NewMembershipRepository()
	now := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	if _, err := addPeople(ctx, users, memberships, devPeople(), now); err != nil {
		t.Fatal(err)
	}
	m, err := memberships.Get(ctx, "dev-operative-1")
	if err != nil {
		t.Fatal(err)
	}
	m.Until, m.EndedBy = now.Add(time.Minute), "dev-admin"
	if err := memberships.Save(ctx, m); err != nil {
		t.Fatal(err)
	}
	if _, err := addPeople(ctx, users, memberships, devPeople(), now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if got, err := memberships.Get(ctx, "dev-operative-1"); err != nil || got.ActiveAt(now.Add(time.Hour)) || got.EndedBy != "dev-admin" {
		t.Errorf("dev-operative-1 = %+v, %v; want it still ended by dev-admin", got, err)
	}
	if got, err := memberships.ListByUser(ctx, "dev-operative"); err != nil || len(got) != 2 {
		t.Errorf("memberships of dev-operative = %+v, %v; want dev-operative-1 and dev-operative-2", got, err)
	}
}

// Outside development only the Secretariat is added: people register by
// signing in, and there is no administrator until one is given.
func TestAddDirectoryAddsNoPeopleOutsideDevelopment(t *testing.T) {
	ctx := t.Context()
	s := stores{users: authmem.NewUserRepository(), areas: authmem.NewAreaRepository(), memberships: authmem.NewMembershipRepository()}
	administrators, err := addDirectory(ctx, config{env: envProduction}, s, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(administrators) != 0 {
		t.Errorf("administrators = %v, want none", administrators)
	}
	for _, synthetic := range devPeople().users {
		if u, err := s.users.GetByEmail(ctx, synthetic.Email); err == nil {
			t.Errorf("found %+v, want no synthetic user", u)
		}
		if m, err := s.memberships.ListByUser(ctx, synthetic.ID); err != nil || len(m) != 0 {
			t.Errorf("memberships of %s = %+v, %v; want none", synthetic.ID, m, err)
		}
	}
	if all, err := s.areas.List(ctx); err != nil || len(all) != 1 || all[0].ID != authinput.SecretariatArea {
		t.Errorf("areas = %+v, %v; want only the Secretariat", all, err)
	}
}

// An area that is already stored keeps what the administrators changed.
func TestAddAreasKeepsTheStoredOnes(t *testing.T) {
	ctx := t.Context()
	areas := authmem.NewAreaRepository()
	if err := addAreas(ctx, areas, devAreas()); err != nil {
		t.Fatal(err)
	}
	office, err := areas.Get(ctx, "office")
	if err != nil {
		t.Fatal(err)
	}
	office.Name = "Área renombrada"
	if err := areas.Save(ctx, office); err != nil {
		t.Fatal(err)
	}
	if err := addAreas(ctx, areas, devAreas()); err != nil {
		t.Fatal(err)
	}
	all, err := areas.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != len(devAreas()) || all[0].Name != "Área renombrada" {
		t.Errorf("areas = %+v, want %d with office renamed", all, len(devAreas()))
	}
}
