package main

import (
	"fmt"
	"slices"
	"testing"
	"time"

	authmem "backend/src/auth/adapters/memory"
	"backend/src/auth/domain"
)

// fileMember is someone of the users file with a membership in office.
func fileMember(id string, permissions ...string) people {
	return people{
		users:       []domain.User{{ID: id, Email: id + "@fi.uba.ar", Name: id}},
		memberships: []domain.Membership{{UserID: id, Area: "office", Permissions: permissions, GrantedBy: usersFileGranter}},
	}
}

func joined(all ...people) people {
	var p people
	for _, one := range all {
		p.users = append(p.users, one.users...)
		p.administrators = append(p.administrators, one.administrators...)
		p.memberships = append(p.memberships, one.memberships...)
	}
	return p
}

func counter() func() string {
	n := 0
	return func() string {
		n++
		return fmt.Sprintf("m%d", n)
	}
}

// Each start loads the file again into the same stores: what the file still
// lists stays as it was granted, a change grants it anew and ends the old
// one, and whoever leaves the file loses their membership.
func TestAddPeopleFollowsTheUsersFile(t *testing.T) {
	ctx := t.Context()
	users, memberships := authmem.NewUserRepository(), authmem.NewMembershipRepository()
	newID := counter()
	first := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	starts := []people{
		joined(fileMember("ana", "review"), fileMember("beto", "review")),
		joined(fileMember("ana", "review"), fileMember("beto", "review", "approve")),
		fileMember("ana", "review"),
	}
	for i, p := range starts {
		if _, err := addPeople(ctx, users, memberships, testAreas, p, newID, first.Add(time.Duration(i)*time.Hour)); err != nil {
			t.Fatalf("start %d: %v", i+1, err)
		}
	}

	ana, err := memberships.ListByUser(ctx, "ana")
	if err != nil {
		t.Fatal(err)
	}
	if len(ana) != 1 || ana[0].ID != "m1" || !ana[0].Until.IsZero() || !ana[0].From.Equal(first) {
		t.Errorf("memberships of ana = %+v, want only m1, active since the first start", ana)
	}
	beto, err := memberships.ListByUser(ctx, "beto")
	if err != nil {
		t.Fatal(err)
	}
	ended := []struct {
		id    string
		until time.Time
	}{{"m2", first.Add(time.Hour)}, {"m3", first.Add(2 * time.Hour)}}
	if len(beto) != len(ended) {
		t.Fatalf("memberships of beto = %+v, want %d", beto, len(ended))
	}
	for i, want := range ended {
		if m := beto[i]; m.ID != want.id || !m.Until.Equal(want.until) || m.EndedBy != usersFileGranter {
			t.Errorf("membership %d of beto = %+v, want %s ended by the file at %s", i+1, m, want.id, want.until)
		}
	}
}

// The file decides who has which membership: one of its memberships that an
// administrator ended is granted again on the next start.
func TestAddPeopleGrantsAgainWhatTheFileStillLists(t *testing.T) {
	ctx := t.Context()
	users, memberships := authmem.NewUserRepository(), authmem.NewMembershipRepository()
	newID := counter()
	now := time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)
	if _, err := addPeople(ctx, users, memberships, testAreas, fileMember("ana", "review"), newID, now); err != nil {
		t.Fatal(err)
	}
	m, err := memberships.Get(ctx, "m1")
	if err != nil {
		t.Fatal(err)
	}
	m.Until, m.EndedBy = now.Add(time.Minute), "admin"
	if err := memberships.Save(ctx, m); err != nil {
		t.Fatal(err)
	}
	if _, err := addPeople(ctx, users, memberships, testAreas, fileMember("ana", "review"), newID, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	ana, err := memberships.ListByUser(ctx, "ana")
	if err != nil {
		t.Fatal(err)
	}
	if len(ana) != 2 || ana[0].EndedBy != "admin" || ana[1].ID != "m2" || !ana[1].ActiveAt(now.Add(time.Hour)) {
		t.Errorf("memberships of ana = %+v, want m1 ended by the administrator and m2 active", ana)
	}
}

// Someone who signed in before the file listed them keeps the id they got
// then, and their membership and administration go to that id.
func TestAddPeopleKeepsTheIDOfWhoSignedInBefore(t *testing.T) {
	ctx := t.Context()
	users, memberships := authmem.NewUserRepository(), authmem.NewMembershipRepository()
	if _, err := users.Register(ctx, domain.User{ID: "signed-in", Email: "ana@fi.uba.ar", Name: "Ana"}); err != nil {
		t.Fatal(err)
	}
	p := fileMember("ana", "review")
	p.administrators = []string{"ana"}
	administrators, err := addPeople(ctx, users, memberships, testAreas, p, counter(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(administrators, []string{"signed-in"}) {
		t.Errorf("administrators = %v, want [signed-in]", administrators)
	}
	if got, err := memberships.ListByUser(ctx, "signed-in"); err != nil || len(got) != 1 {
		t.Errorf("memberships of signed-in = %+v, %v; want the one of the file", got, err)
	}
}

// The file cannot give a stored id another email.
func TestAddPeopleRejectsAnotherEmailForAStoredID(t *testing.T) {
	ctx := t.Context()
	users, memberships := authmem.NewUserRepository(), authmem.NewMembershipRepository()
	if _, err := addPeople(ctx, users, memberships, testAreas, fileMember("ana", "review"), counter(), time.Now()); err != nil {
		t.Fatal(err)
	}
	changed := fileMember("ana", "review")
	changed.users[0].Email = "ana.perez@fi.uba.ar"
	if _, err := addPeople(ctx, users, memberships, testAreas, changed, counter(), time.Now()); err == nil {
		t.Error("addPeople() with another email for ana = nil, want an error")
	}
}

// The memberships of development keep their ids and are added once.
func TestAddPeopleAddsTheDevelopmentMembershipsOnce(t *testing.T) {
	ctx := t.Context()
	users, memberships := authmem.NewUserRepository(), authmem.NewMembershipRepository()
	for range 2 {
		if _, err := addPeople(ctx, users, memberships, devAreas(), devPeople(), counter(), time.Now()); err != nil {
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
