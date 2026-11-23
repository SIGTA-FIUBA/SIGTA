package application

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

func TestDetailPutsTheCurrentFileNextToEachDocument(t *testing.T) {
	f := newDocumentsFixture(t)
	for _, u := range []input.Upload{upload("annex", "anexo.pdf", pdf(100)), upload("statement", "primera.pdf", pdf(100)), upload("statement", "segunda.pdf", pdf(100))} {
		if _, err := f.service.Attach(t.Context(), applicant, u); err != nil {
			t.Fatal(err)
		}
	}
	for _, user := range []string{applicant, reviewer} {
		detail, err := f.details.Detail(t.Context(), user, "P-1")
		if err != nil {
			t.Fatalf("%s: %v", user, err)
		}
		var got []string
		for _, s := range detail.Steps {
			for _, d := range s.Files {
				got = append(got, s.Step.ID+" "+d.Requirement+" "+d.FileName)
			}
		}
		want := []string{"documents statement segunda.pdf", "documents annex anexo.pdf"}
		if !slices.Equal(got, want) {
			t.Errorf("%s: files = %v, want %v", user, got, want)
		}
	}
}

func TestDetailPutsTheValuesNextToTheFieldsInTheOrderOfTheDefinition(t *testing.T) {
	f := newDocumentsFixture(t)
	p := domain.Procedure{ID: "P-2", ProcedureType: 1, Version: 1, VersionDigest: f.digest, Applicant: applicant, OpenStep: "review"}
	values := []domain.FieldValue{
		{Procedure: "P-2", Field: "subjects", Step: "request", Items: []string{"Análisis II", "Física I"}, EnteredBy: applicant, EnteredAt: attachedAt},
		{Procedure: "P-2", Field: "reason", Step: "request", Text: "Pase de carrera", EnteredBy: applicant, EnteredAt: attachedAt},
	}
	if err := f.procedures.Add(t.Context(), output.Start{Procedure: p, Values: values}); err != nil {
		t.Fatal(err)
	}

	detail, err := f.details.Detail(t.Context(), reviewer, "P-2")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := detail.Steps[0].Values, []domain.FieldValue{values[1], values[0]}; !reflect.DeepEqual(got, want) {
		t.Errorf("values of the first step = %+v, want %+v", got, want)
	}
	if got := detail.Steps[1].Values; len(got) != 0 {
		t.Errorf("values of a step that asks for no field = %+v, want none", got)
	}
}

func TestDetailSaysTheStateAndWhoActs(t *testing.T) {
	f := newDocumentsFixture(t)
	f.on(t, "P-2", "review")
	f.on(t, "P-3", "")
	for _, tc := range []struct {
		procedure, user string
		state           domain.State
		canAct          bool
	}{
		{"P-1", applicant, domain.StateWithApplicant, true},
		{"P-1", reviewer, domain.StateWithApplicant, false},
		{"P-2", applicant, domain.StateInOffice, false},
		{"P-2", reviewer, domain.StateInOffice, true},
		{"P-3", applicant, domain.StateClosed, false},
		{"P-3", reviewer, domain.StateClosed, false},
	} {
		detail, err := f.details.Detail(t.Context(), tc.user, tc.procedure)
		if err != nil {
			t.Fatalf("%s on %s: %v", tc.user, tc.procedure, err)
		}
		if detail.State != tc.state || detail.CanAct != tc.canAct || detail.Name != "Trámite con documentos" {
			t.Errorf("%s on %s: %s, can act %t, name %q; want %s and %t", tc.user, tc.procedure, detail.State, detail.CanAct, detail.Name, tc.state, tc.canAct)
		}
	}
}

func TestDetailShowsTheHistoryInTheOrderItWasRecorded(t *testing.T) {
	f := newDocumentsFixture(t)
	f.on(t, "P-2", "review")
	p, err := f.procedures.Get(t.Context(), "P-2")
	if err != nil {
		t.Fatal(err)
	}
	reviewed := domain.Event{ProcedureID: "P-2", Step: "review", By: reviewer, Observation: "ok"}
	p.OpenStep = ""
	if err := f.procedures.Advance(t.Context(), p, reviewed); err != nil {
		t.Fatal(err)
	}

	detail, err := f.details.Detail(t.Context(), reviewer, "P-2")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(detail.History, []domain.Event{reviewed}) || detail.AsApplicant {
		t.Errorf("history = %+v, as applicant %t; want %+v, as staff", detail.History, detail.AsApplicant, []domain.Event{reviewed})
	}

	mine, err := f.details.Detail(t.Context(), applicant, "P-2")
	if err != nil {
		t.Fatal(err)
	}
	reviewed.By, reviewed.Observation = "", ""
	if !slices.Equal(mine.History, []domain.Event{reviewed}) || !mine.AsApplicant {
		t.Errorf("history of the applicant = %+v, as applicant %t; want %+v, without who did it nor the observation", mine.History, mine.AsApplicant, []domain.Event{reviewed})
	}
}

func TestDetailListsThePendingCorrectionRequests(t *testing.T) {
	f := newDocumentsFixture(t)
	f.on(t, "P-2", "review")
	asked := domain.CorrectionRequest{Step: "review", By: reviewer, Reason: "the codes are missing", Fields: []string{"subjects"}}
	if err := f.procedures.RequestCorrection(t.Context(), asked, domain.Event{ProcedureID: "P-2", Step: "review", By: reviewer}); err != nil {
		t.Fatal(err)
	}

	withoutWho := asked
	withoutWho.By = ""
	for user, want := range map[string]domain.CorrectionRequest{reviewer: asked, applicant: withoutWho} {
		detail, err := f.details.Detail(t.Context(), user, "P-2")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(detail.CorrectionRequests, []domain.CorrectionRequest{want}) {
			t.Errorf("%s sees the correction requests %+v, want %+v", user, detail.CorrectionRequests, want)
		}
	}
}

// Someone who cannot see a procedure gets the same answer as for one that
// does not exist, so the detail does not tell which procedures exist.
func TestDetailAnswersNotFoundToWhoCannotSeeTheProcedure(t *testing.T) {
	f := newDocumentsFixture(t)
	if _, err := f.memberships.Grant(t.Context(), "admin", authinput.MembershipGrant{UserID: outsider, Area: office.ID, Permissions: []string{"review"}, Scope: []int64{2}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ user, procedure string }{
		{"nadie@fi.uba.ar", "P-1"},
		{"nadie@fi.uba.ar", "P-missing"},
		{outsider, "P-1"},
	} {
		if _, err := f.details.Detail(t.Context(), tc.user, tc.procedure); !errors.Is(err, domain.ErrProcedureNotFound) {
			t.Errorf("%s on %s: err = %v, want ErrProcedureNotFound", tc.user, tc.procedure, err)
		}
	}
}

func TestDetailFailsIfTheOpenStepIsNotInTheVersion(t *testing.T) {
	f := newDocumentsFixture(t)
	f.on(t, "P-2", "missing")
	_, err := f.details.Detail(t.Context(), applicant, "P-2")
	if err == nil || errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("err = %v, want an error other than ErrProcedureNotFound", err)
	}
}

func TestReceiptIsTheDetailOfTheApplicantWhenItIsIssued(t *testing.T) {
	f := newDocumentsFixture(t)
	f.on(t, "P-2", "review")
	before := time.Now()
	receipt, err := f.details.Receipt(t.Context(), applicant, "P-2")
	if err != nil {
		t.Fatal(err)
	}
	detail, err := f.details.Detail(t.Context(), applicant, "P-2")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(receipt.ProcedureDetail, detail) || receipt.IssuedAt.Before(before) || receipt.IssuedAt.After(time.Now()) {
		t.Errorf("receipt = %+v, want the detail of the applicant issued now", receipt)
	}
}

func TestReceiptIsNotFoundForStaffOfTheProcedure(t *testing.T) {
	f := newDocumentsFixture(t)
	f.on(t, "P-2", "review")
	for _, user := range []string{reviewer, "nadie@fi.uba.ar"} {
		if _, err := f.details.Receipt(t.Context(), user, "P-2"); !errors.Is(err, domain.ErrProcedureNotFound) {
			t.Errorf("%s: err = %v, want ErrProcedureNotFound", user, err)
		}
	}
}

func TestMineListsTheProceduresOfTheApplicantMostRecentlyMovedFirst(t *testing.T) {
	f := newDocumentsFixture(t)
	day := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	for _, p := range []domain.Procedure{
		{ID: "P-old", OpenStep: "review", MovedAt: day},
		{ID: "P-new", OpenStep: "documents", MovedAt: day.Add(48 * time.Hour)},
		{ID: "P-closed", OpenStep: "", MovedAt: day.Add(72 * time.Hour)},
		{ID: "P-other", OpenStep: "review", MovedAt: day.Add(96 * time.Hour), Applicant: reviewer},
	} {
		p.ProcedureType, p.Version, p.VersionDigest = 1, 1, f.digest
		if p.Applicant == "" {
			p.Applicant = applicant
		}
		if err := f.procedures.Add(t.Context(), output.Start{Procedure: p}); err != nil {
			t.Fatal(err)
		}
	}

	mine, err := f.details.Mine(t.Context(), applicant)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, m := range mine {
		area := ""
		if m.Area != nil {
			area = m.Area.Name
		}
		got = append(got, m.Procedure.ID+" "+m.StepName+" "+area+" "+string(m.State))
	}
	// P-1 is the fixture's own procedure, never moved. The steps of the
	// applicant and a closed procedure have no area.
	want := []string{"P-closed   closed", "P-new Documentación  with_applicant", "P-old Revisión Secretaría in_office", "P-1 Documentación  with_applicant"}
	if !slices.Equal(got, want) {
		t.Errorf("mine = %v, want %v", got, want)
	}
	for _, m := range mine {
		if m.Name != "Trámite con documentos" {
			t.Errorf("%s: name %q", m.Procedure.ID, m.Name)
		}
	}
}

func TestMineIsEmptyForSomeoneWithoutProcedures(t *testing.T) {
	f := newDocumentsFixture(t)
	mine, err := f.details.Mine(t.Context(), "nobody@fi.uba.ar")
	if err != nil || mine == nil || len(mine) != 0 {
		t.Errorf("mine = %#v, %v; want an empty list", mine, err)
	}
}
