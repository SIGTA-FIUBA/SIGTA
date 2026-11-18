package main

import (
	"bytes"
	"slices"
	"testing"
	"time"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	"backend/src/procedure/adapters/disk"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/application"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/input"
)

// The procedures of the inbox follow the steps of the example type, so the
// staff of development finds work in it: the steps of the area are theirs,
// and those of the applicant only show in the full list.
func TestDevInboxShowsWorkToTheStaff(t *testing.T) {
	ctx := t.Context()
	versions := memory.NewProcedureVersions()
	types := application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now)
	if err := publishIfMissing(ctx, versions, types, devProcedureType()); err != nil {
		t.Fatal(err)
	}
	if err := publishIfMissing(ctx, versions, types, devReferralProcedureType()); err != nil {
		t.Fatal(err)
	}
	if err := publishIfMissing(ctx, versions, types, devRulingProcedureType()); err != nil {
		t.Fatal(err)
	}
	procedures := memory.NewProcedureRepository()
	if err := addDevInbox(ctx, versions, procedures, time.Now()); err != nil {
		t.Fatal(err)
	}
	directory := devPeople()
	areas := authmem.NewAreaRepository(devAreas()...)
	memberships := authapp.NewMembershipService(areas, authmem.NewAdministrators(directory.administrators...), authmem.NewMembershipRepository(directory.memberships...), func() string { return "id" }, time.Now)
	service := application.NewInboxService(application.NewProcedureService(types, procedures, memory.NewDocumentRepository(), memberships, func() string { return "id" }, time.Now))

	mine, err := service.List(ctx, "dev-operative", inbox.Query{View: inbox.ViewMine})
	if err != nil {
		t.Fatal(err)
	}
	all, err := service.List(ctx, "dev-operative", inbox.Query{View: inbox.ViewAll})
	if err != nil {
		t.Fatal(err)
	}
	if len(mine.Items) != 5 || len(all.Items) != 9 {
		t.Errorf("the operative has %d to do of %d, want 5 of 9", len(mine.Items), len(all.Items))
	}
	committee, err := service.List(ctx, "dev-committee", inbox.Query{View: inbox.ViewMine})
	if err != nil {
		t.Fatal(err)
	}
	if len(committee.Items) != 2 || committee.Items[0].State != domain.StateReferred || committee.Items[1].State != domain.StateReferred {
		t.Errorf("the committee has %+v to do, want the two referred procedures", committee.Items)
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

// The staff of development can walk the example with a referral to the end:
// the review refers it to the committee, whose ruling returns it.
func TestDevReferralExampleGoesToTheCommitteeAndBack(t *testing.T) {
	ctx := t.Context()
	versions := memory.NewProcedureVersions()
	types := application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now)
	if err := publishIfMissing(ctx, versions, types, devReferralProcedureType()); err != nil {
		t.Fatal(err)
	}
	if err := publishIfMissing(ctx, versions, types, devRulingProcedureType()); err != nil {
		t.Fatal(err)
	}
	if err := publishIfMissing(ctx, versions, types, devProcedureType()); err != nil {
		t.Fatal(err)
	}
	procedures := memory.NewProcedureRepository()
	if err := addDevInbox(ctx, versions, procedures, time.Now()); err != nil {
		t.Fatal(err)
	}
	directory := devPeople()
	areas := authmem.NewAreaRepository(devAreas()...)
	memberships := authapp.NewMembershipService(areas, authmem.NewAdministrators(directory.administrators...), authmem.NewMembershipRepository(directory.memberships...), func() string { return "id" }, time.Now)
	service := application.NewProcedureService(types, procedures, memory.NewDocumentRepository(), memberships, func() string { return "id" }, time.Now)

	for _, step := range []struct {
		user string
		c    input.StepConfirmation
	}{
		{"dev-operative", input.StepConfirmation{Step: devReviewStep, Observation: "Completo", Outcome: "accept"}},
		{"dev-committee", input.StepConfirmation{Step: devEvaluateStep, Observation: "Dictamen recibido"}},
		{"dev-operative", input.StepConfirmation{Step: "communicate", Observation: "Comunicado"}},
	} {
		if _, err := service.Confirm(ctx, step.user, "dev-07", step.c); err != nil {
			t.Fatalf("%s confirms %s: %v", step.user, step.c.Step, err)
		}
	}
	events, err := procedures.Events(ctx, "dev-07")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range events {
		got = append(got, e.Step)
	}
	want := []string{"request", devDocumentsStep, devReviewStep, devEvaluateStep, "communicate"}
	if !slices.Equal(got, want) {
		t.Errorf("history = %v, want %v", got, want)
	}
	if events[2].Referral == nil || !events[3].Returned {
		t.Errorf("review referral = %+v, ruling returned = %t; want a referral and a return", events[2].Referral, events[3].Returned)
	}
}

// Once the committee's ruling is back, the staff of development attaches it
// on the last step of the example with a ruling.
func TestDevRulingExampleAsksTheStaffForTheRuling(t *testing.T) {
	ctx := t.Context()
	versions := memory.NewProcedureVersions()
	types := application.NewProcedureTypeService(versions, devAreaService(), devStaff(), time.Now)
	for _, example := range []input.Publication{devProcedureType(), devReferralProcedureType(), devRulingProcedureType()} {
		if err := publishIfMissing(ctx, versions, types, example); err != nil {
			t.Fatal(err)
		}
	}
	procedures := memory.NewProcedureRepository()
	if err := addDevInbox(ctx, versions, procedures, time.Now()); err != nil {
		t.Fatal(err)
	}
	directory := devPeople()
	areas := authmem.NewAreaRepository(devAreas()...)
	memberships := authapp.NewMembershipService(areas, authmem.NewAdministrators(directory.administrators...), authmem.NewMembershipRepository(directory.memberships...), func() string { return "id" }, time.Now)
	attached := memory.NewDocumentRepository()
	service := application.NewProcedureService(types, procedures, attached, memberships, func() string { return "id" }, time.Now)
	files, err := disk.NewFiles(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	documents := application.NewDocumentService(service, attached, files, func() string { return "D-1" }, time.Now)

	if _, err := service.Confirm(ctx, "dev-committee", "dev-09", input.StepConfirmation{Step: devEvaluateStep, Observation: "Dictamen recibido"}); err != nil {
		t.Fatal(err)
	}
	ruling := input.Upload{ProcedureID: "dev-09", Requirement: "ruling", FileName: "dictamen.pdf", Content: bytes.NewReader([]byte("%PDF-1.7\n"))}
	if _, err := documents.Attach(ctx, "dev-operative", ruling); err != nil {
		t.Fatalf("attach the ruling: %v", err)
	}
	if values, err := procedures.Values(ctx, "dev-09"); err != nil || len(values) == 0 {
		t.Errorf("values = %v, %v; want what the applicant entered", values, err)
	}
}
