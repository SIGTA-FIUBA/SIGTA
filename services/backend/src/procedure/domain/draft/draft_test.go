package draft_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"backend/src/procedure/domain/definition/definitiontest"
	"backend/src/procedure/domain/draft"
	"backend/src/procedure/domain/version"
)

var at = time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)

func proposed(t *testing.T) draft.Draft {
	t.Helper()
	d, err := draft.New(3, 0, definitiontest.Definition(), "ana", at).Propose("ana", at.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestNewDraftIsBeingEdited(t *testing.T) {
	d := draft.New(3, 2, definitiontest.Definition(), "ana", at)
	if d.Status != draft.StatusEditing || d.Revision != 1 || d.Base != 2 || d.EditedBy != "ana" || !slices.Equal(d.Editors, []string{"ana"}) {
		t.Errorf("New() = %+v", d)
	}
}

func TestEditAndProposeBumpTheRevision(t *testing.T) {
	def := definitiontest.Definition()
	def.Name = "Otro nombre"
	d, err := draft.New(3, 0, definitiontest.Definition(), "ana", at).Edit(def, "beto", at.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if d.Revision != 2 || d.EditedBy != "beto" || d.Definition.Name != "Otro nombre" || !slices.Equal(d.Editors, []string{"ana", "beto"}) {
		t.Errorf("Edit() = %+v", d)
	}
	d, err = d.Propose("beto", at.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != draft.StatusProposed || d.Revision != 3 || d.ProposedBy != "beto" {
		t.Errorf("Propose() = %+v", d)
	}
}

func TestEditorsAreRememberedOnce(t *testing.T) {
	d := draft.New(3, 0, definitiontest.Definition(), "ana", at)
	for _, by := range []string{"beto", "ANA", "beto"} {
		var err error
		if d, err = d.Edit(definitiontest.Definition(), by, at); err != nil {
			t.Fatal(err)
		}
	}
	if !slices.Equal(d.Editors, []string{"ana", "beto"}) {
		t.Errorf("Editors = %v, want [ana beto]", d.Editors)
	}
}

func TestProposedDraftIsNotEditedNorProposedAgain(t *testing.T) {
	d := proposed(t)
	if _, err := d.Edit(definitiontest.Definition(), "beto", at); !errors.Is(err, draft.ErrDraftProposed) {
		t.Errorf("Edit() = %v, want ErrDraftProposed", err)
	}
	if _, err := d.Propose("beto", at); !errors.Is(err, draft.ErrDraftProposed) {
		t.Errorf("Propose() = %v, want ErrDraftProposed", err)
	}
}

func TestApproval(t *testing.T) {
	got, err := proposed(t).Approval("beto")
	if err != nil || got != (version.Authorship{ProposedBy: "ana", ApprovedBy: "beto"}) {
		t.Errorf("Approval() = %+v, %v", got, err)
	}
	for _, by := range []string{"ana", "ANA"} {
		if _, err := proposed(t).Approval(by); !errors.Is(err, draft.ErrSelfApproval) {
			t.Errorf("Approval(%q) = %v, want ErrSelfApproval", by, err)
		}
	}
	edited, err := draft.New(3, 0, definitiontest.Definition(), "ana", at).Edit(definitiontest.Definition(), "beto", at)
	if err != nil {
		t.Fatal(err)
	}
	edited, err = edited.Propose("ana", at)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := edited.Approval("Beto"); !errors.Is(err, draft.ErrSelfApproval) {
		t.Errorf("Approval() by an editor = %v, want ErrSelfApproval", err)
	}
	if _, err := edited.Approval("carla"); err != nil {
		t.Errorf("Approval() by someone who did not edit it = %v", err)
	}
	editing := draft.New(3, 0, definitiontest.Definition(), "ana", at)
	if _, err := editing.Approval("beto"); !errors.Is(err, draft.ErrDraftNotProposed) {
		t.Errorf("Approval() of a draft being edited = %v, want ErrDraftNotProposed", err)
	}
}

func TestReject(t *testing.T) {
	d, err := proposed(t).Reject("beto", "Falta el plazo", at.Add(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != draft.StatusEditing || d.ProposedBy != "" || d.Rejection.Reason != "Falta el plazo" || d.Rejection.By != "beto" || d.Revision != 3 {
		t.Errorf("Reject() = %+v", d)
	}
	if _, err := d.Edit(definitiontest.Definition(), "ana", at); err != nil {
		t.Errorf("Edit() after a rejection = %v", err)
	}

	tests := []struct {
		name   string
		d      draft.Draft
		by     string
		reason string
		want   error
	}{
		{name: "by whoever proposed it", d: proposed(t), by: "ana", reason: "No", want: draft.ErrSelfApproval},
		{name: "without a reason", d: proposed(t), by: "beto", reason: " ", want: draft.ErrRejectionReasonMissing},
		{name: "not proposed", d: draft.New(3, 0, definitiontest.Definition(), "ana", at), by: "beto", reason: "No", want: draft.ErrDraftNotProposed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.d.Reject(tt.by, tt.reason, at); !errors.Is(err, tt.want) {
				t.Errorf("Reject() = %v, want %v", err, tt.want)
			}
		})
	}
}
