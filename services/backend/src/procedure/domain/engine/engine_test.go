package engine_test

import (
	"errors"
	"testing"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
)

var equivalence = definition.ProcedureTypeDefinition{
	Steps: []definition.Step{
		{ID: "request", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Next: "check"},
		{ID: "check", Role: definition.RoleStaff, Area: "affairs", Action: definition.ActionDecide, Outcomes: []definition.Outcome{
			{ID: "accept", Next: "evaluate"},
			{ID: "send_back", Next: "request"},
		}},
		{ID: "evaluate", Role: definition.RoleStaff, Area: "commission", Action: definition.ActionConfirmExternalFact, Next: "review"},
		{ID: "review", Role: definition.RoleStaff, Area: "board", Action: definition.ActionConfirmExternalFact, Next: "file"},
		{ID: "file", Role: definition.RoleStaff, Area: "affairs", Action: definition.ActionConfirmExternalFact, Next: "approve"},
		{ID: "approve", Role: definition.RoleStaff, Area: "affairs", Action: definition.ActionEnterData},
	},
}

var (
	monday   = time.Date(2026, 11, 16, 10, 30, 0, 0, time.UTC)
	thursday = time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)
)

func on(step string) domain.Procedure {
	return domain.Procedure{ID: "P-1", Applicant: "applicant@fi.uba.ar", OpenStep: step}
}

func confirm(step, outcome string) engine.Confirmation {
	return engine.Confirmation{Step: step, By: "staff@fi.uba.ar", At: monday, Observation: "done", Outcome: outcome}
}

func TestConfirmRecordsWhoWhenAndObservation(t *testing.T) {
	_, event, err := engine.Confirm(on("file"), equivalence, confirm("file", ""))
	if err != nil {
		t.Fatal(err)
	}
	want := engine.Event{
		ProcedureID: "P-1",
		Step:        "file",
		Action:      definition.ActionConfirmExternalFact,
		By:          "staff@fi.uba.ar",
		At:          monday,
		Observation: "done",
		Declared:    true,
		Next:        "approve",
	}
	if event != want {
		t.Errorf("event = %+v, want %+v", event, want)
	}
}

func TestConfirmOpensTheNextStep(t *testing.T) {
	tests := []struct {
		name    string
		open    string
		outcome string
		next    string
	}{
		{"configured next", "file", "", "approve"},
		{"chosen outcome", "check", "accept", "evaluate"},
		{"send back", "check", "send_back", "request"},
		{"last step closes", "approve", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moved, event, err := engine.Confirm(on(tt.open), equivalence, confirm(tt.open, tt.outcome))
			if err != nil {
				t.Fatal(err)
			}
			if moved.OpenStep != tt.next || event.Next != tt.next {
				t.Errorf("open step = %q and event next = %q, want %q", moved.OpenStep, event.Next, tt.next)
			}
			if moved.Closed() != (tt.next == "") {
				t.Errorf("closed = %v, want %v", moved.Closed(), tt.next == "")
			}
		})
	}
}

func TestConfirmRejects(t *testing.T) {
	closed := on("")
	tests := []struct {
		name         string
		procedure    domain.Procedure
		confirmation engine.Confirmation
		want         error
	}{
		{"a step that is not open", on("file"), confirm("approve", ""), engine.ErrStepNotOpen},
		{"a closed procedure", closed, confirm("approve", ""), engine.ErrStepNotOpen},
		{"no observation", on("file"), engine.Confirmation{Step: "file", Observation: "  "}, engine.ErrNoObservation},
		{"a decision without outcome", on("check"), confirm("check", ""), engine.ErrInvalidOutcome},
		{"an unknown outcome", on("check"), confirm("check", "fast_track"), engine.ErrInvalidOutcome},
		{"an outcome outside a decision", on("file"), confirm("file", "accept"), engine.ErrInvalidOutcome},
		{"a file number without referral", on("file"), engine.Confirmation{Step: "file", Observation: "done", ExternalFileNumber: "EX-1"}, engine.ErrNotAReferral},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moved, _, err := engine.Confirm(tt.procedure, equivalence, tt.confirmation)
			if !errors.Is(err, tt.want) || !errors.Is(err, engine.ErrInvalidConfirmation) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if moved.OpenStep != tt.procedure.OpenStep {
				t.Errorf("open step = %q, want it unchanged at %q", moved.OpenStep, tt.procedure.OpenStep)
			}
		})
	}
}

func TestConfirmOnAnOpenStepMissingFromTheVersionFails(t *testing.T) {
	_, _, err := engine.Confirm(on("gone"), equivalence, confirm("gone", ""))
	if err == nil || errors.Is(err, engine.ErrInvalidConfirmation) {
		t.Errorf("err = %v, want an inconsistency that is not the person's fault", err)
	}
}

func TestOnlyExternalFactsAreDeclared(t *testing.T) {
	_, event, err := engine.Confirm(on("check"), equivalence, confirm("check", "send_back"))
	if err != nil {
		t.Fatal(err)
	}
	if event.Declared {
		t.Error("a decision is recorded as declared, want it only for external facts")
	}
}

func TestMovingToAnotherAreaRefers(t *testing.T) {
	c := confirm("check", "accept")
	c.ExternalFileNumber = "EX-2026-00001"
	moved, event, err := engine.Confirm(on("check"), equivalence, c)
	if err != nil {
		t.Fatal(err)
	}
	wantRecord := engine.ReferralRecord{From: "affairs", To: "commission", ExternalFileNumber: "EX-2026-00001"}
	if event.Referral == nil || *event.Referral != wantRecord {
		t.Errorf("event referral = %+v, want %+v", event.Referral, wantRecord)
	}
	wantReferral := domain.Referral{From: "affairs", To: "commission", Since: monday}
	if moved.Referral == nil || *moved.Referral != wantReferral {
		t.Errorf("procedure referral = %+v, want %+v", moved.Referral, wantReferral)
	}
}

func TestStayingInTheAreaOrGoingToTheApplicantDoesNotRefer(t *testing.T) {
	for _, tt := range []struct{ open, outcome string }{{"file", ""}, {"check", "send_back"}} {
		moved, event, err := engine.Confirm(on(tt.open), equivalence, confirm(tt.open, tt.outcome))
		if err != nil {
			t.Fatal(err)
		}
		if event.Referral != nil || moved.Referral != nil || event.Returned {
			t.Errorf("confirming %q: referral = %+v, procedure referral = %+v, returned = %v; want none", tt.open, event.Referral, moved.Referral, event.Returned)
		}
	}
}

func TestAReferralFromOutsideKeepsWhereItLeftAndSince(t *testing.T) {
	p := on("evaluate")
	p.Referral = &domain.Referral{From: "affairs", To: "commission", Since: monday}
	c := confirm("evaluate", "")
	c.At = thursday

	moved, event, err := engine.Confirm(p, equivalence, c)
	if err != nil {
		t.Fatal(err)
	}
	if event.Referral == nil || event.Referral.From != "commission" || event.Referral.To != "board" {
		t.Errorf("event referral = %+v, want from commission to board", event.Referral)
	}
	want := domain.Referral{From: "affairs", To: "board", Since: monday}
	if moved.Referral == nil || *moved.Referral != want {
		t.Errorf("procedure referral = %+v, want %+v", moved.Referral, want)
	}
}

func TestGoingBackToTheAreaItLeftReturns(t *testing.T) {
	p := on("review")
	p.Referral = &domain.Referral{From: "affairs", To: "board", Since: monday}

	moved, event, err := engine.Confirm(p, equivalence, confirm("review", ""))
	if err != nil {
		t.Fatal(err)
	}
	if !event.Returned || event.Referral != nil {
		t.Errorf("returned = %v and referral = %+v, want a return without a new referral", event.Returned, event.Referral)
	}
	if moved.Referral != nil {
		t.Errorf("procedure referral = %+v, want none after returning", moved.Referral)
	}
}

func TestConfirmDoesNotChangeTheProcedureItReceives(t *testing.T) {
	p := on("check")
	if _, _, err := engine.Confirm(p, equivalence, confirm("check", "accept")); err != nil {
		t.Fatal(err)
	}
	if p.OpenStep != "check" || p.Referral != nil {
		t.Errorf("procedure = %+v, want it untouched", p)
	}
}
