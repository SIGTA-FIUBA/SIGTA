package main

import (
	"testing"
	"time"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain/inbox"
)

// The procedures of the inbox follow the steps of the example type, so the
// staff of development finds work in it: the steps of the area are theirs,
// and those of the applicant only show in the full list.
func TestDevInboxShowsWorkToTheStaff(t *testing.T) {
	ctx := t.Context()
	versions := memory.NewProcedureVersions()
	types := application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now)
	if err := publishDevProcedureType(ctx, versions, types); err != nil {
		t.Fatal(err)
	}
	procedures := memory.NewProcedureRepository()
	if err := addDevInbox(ctx, versions, procedures, time.Now()); err != nil {
		t.Fatal(err)
	}
	directory := devPeople()
	areas := authmem.NewAreaRepository(devAreas()...)
	memberships := authapp.NewMembershipService(areas, authmem.NewAdministrators(directory.administrators...), authmem.NewMembershipRepository(directory.memberships...), func() string { return "id" }, time.Now)
	service := application.NewInboxService(application.NewProcedureService(types, procedures, memberships, func() string { return "id" }, time.Now), time.Now)

	mine, err := service.List(ctx, "dev-operative", inbox.Query{View: inbox.ViewMine})
	if err != nil {
		t.Fatal(err)
	}
	all, err := service.List(ctx, "dev-operative", inbox.Query{View: inbox.ViewAll})
	if err != nil {
		t.Fatal(err)
	}
	if len(mine.Items) != 4 || len(all.Items) != 6 {
		t.Errorf("the operative has %d to do of %d, want 4 of 6", len(mine.Items), len(all.Items))
	}
	known := map[string]bool{}
	for _, user := range directory.users {
		known[user.ID] = true
	}
	for _, item := range all.Items {
		if !known[item.ApplicantID] {
			t.Errorf("%s is of %s, who is not an account of development", item.ID, item.ApplicantID)
		}
	}
}
