package application

import (
	"errors"
	"slices"
	"testing"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
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

// Someone who cannot see a procedure gets the same answer as for one that
// does not exist, so the detail does not tell which procedures exist.
func TestDetailAnswersNotFoundToWhoCannotSeeTheProcedure(t *testing.T) {
	f := newDocumentsFixture(t)
	other := "otra@fi.uba.ar"
	if _, err := f.memberships.Grant(t.Context(), "admin", authinput.MembershipGrant{UserID: other, Area: office.ID, Permissions: []string{"review"}, Scope: []int64{2}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ user, procedure string }{
		{"nadie@fi.uba.ar", "P-1"},
		{"nadie@fi.uba.ar", "P-missing"},
		{other, "P-1"},
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
