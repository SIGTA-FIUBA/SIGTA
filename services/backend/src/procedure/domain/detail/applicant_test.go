package detail_test

import (
	"slices"
	"testing"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/domain/version"
)

// applicantContent has a step of the applicant, a staff step that refers the
// procedure and asks for a field, and a staff step that attaches the minutes
// and the resolution.
func applicantContent() version.Content {
	return version.Content{
		Definition: definition.ProcedureTypeDefinition{
			Fields:    []definition.Field{{ID: "reason"}, {ID: "record_number"}},
			Documents: []definition.Document{{ID: "transcript"}, {ID: "minutes"}, {ID: "resolution", Resolution: true}},
			Steps: []definition.Step{
				{ID: "request", Role: definition.RoleApplicant, Action: definition.ActionUploadDocument, Fields: []string{"reason"}, Documents: []string{"transcript"}, Next: "review"},
				{ID: "review", Role: definition.RoleStaff, Area: "office", Action: definition.ActionEnterData, Fields: []string{"reason", "record_number"}, Next: "ruling"},
				{ID: "ruling", Role: definition.RoleStaff, Area: "commission", Action: definition.ActionUploadDocument, Documents: []string{"minutes", "resolution"}},
			},
		},
	}
}

func applicantSteps(t *testing.T, closed bool) []detail.Step {
	t.Helper()
	c := applicantContent()
	steps := detail.WithValues(detail.WithFiles(detail.Steps(c), []domain.Document{
		{ID: "D-1", Requirement: "transcript", Step: "request", UploadedBy: "applicant"},
		{ID: "D-2", Requirement: "minutes", Step: "ruling"},
		{ID: "D-3", Requirement: "resolution", Step: "ruling", UploadedBy: "commission"},
	}), []domain.FieldValue{
		{Field: "reason", Step: "request", Text: "Pase de carrera"},
		{Field: "record_number", Step: "review", Text: "EX-1"},
	})
	p := domain.Procedure{OpenStep: "ruling"}
	if closed {
		p.OpenStep = ""
	}
	return detail.ForApplicant(c, p).Steps(steps)
}

func fileIDs(s detail.Step) []string {
	var ids []string
	for _, d := range s.Files {
		ids = append(ids, d.ID)
	}
	return ids
}

func TestTheApplicantSeesWhatTheyEnteredAndAttached(t *testing.T) {
	steps := applicantSteps(t, false)

	if got := fileIDs(steps[0]); !slices.Equal(got, []string{"D-1"}) || steps[0].Files[0].UploadedBy != "applicant" {
		t.Errorf("files of the request = %+v, want D-1 with who uploaded it", steps[0].Files)
	}
	if len(steps[0].Values) != 1 || steps[0].Values[0].Text != "Pase de carrera" {
		t.Errorf("values of the request = %+v, want the reason", steps[0].Values)
	}
	if len(steps[0].Fields) != 1 || len(steps[0].Documents) != 1 {
		t.Errorf("request = %+v, want its field and its document", steps[0])
	}
}

func TestTheApplicantDoesNotSeeWhatStaffEntered(t *testing.T) {
	steps := applicantSteps(t, false)

	if s := steps[1]; len(s.Fields) != 0 || len(s.Values) != 0 {
		t.Errorf("review = %+v, want no fields nor values", s)
	}
	if s := steps[2]; len(s.Documents) != 0 || len(s.Files) != 0 {
		t.Errorf("ruling = %+v, want no documents nor files while the procedure is open", s)
	}
	if s := steps[1]; s.Area == nil || s.Step.ID != "review" {
		t.Errorf("review = %+v, want the step and its area", s)
	}
}

// The reason is asked again in the review: what staff enters there is not
// what the applicant entered.
func TestTheApplicantDoesNotSeeAValueStaffEnteredForTheirField(t *testing.T) {
	c := applicantContent()
	steps := detail.WithValues(detail.Steps(c), []domain.FieldValue{
		{Field: "reason", Step: "request", Text: "Pase de carrera"},
		{Field: "reason", Step: "review", Text: "Corregido por la Dirección"},
	})

	seen := detail.ForApplicant(c, domain.Procedure{OpenStep: "ruling"}).Steps(steps)

	if len(seen[0].Values) != 0 {
		t.Errorf("values of the request = %+v, want none: the last one is from staff", seen[0].Values)
	}
}

func TestTheApplicantSeesTheResolutionOnceTheProcedureCloses(t *testing.T) {
	steps := applicantSteps(t, true)

	s := steps[2]
	if len(s.Documents) != 1 || s.Documents[0].Document.ID != "resolution" {
		t.Errorf("documents of the ruling = %+v, want only the resolution", s.Documents)
	}
	if got := fileIDs(s); !slices.Equal(got, []string{"D-3"}) {
		t.Errorf("files of the ruling = %v, want D-3", got)
	}
	if s.Files[0].UploadedBy != "" {
		t.Errorf("the resolution was uploaded by %q, want no one shown", s.Files[0].UploadedBy)
	}
}

func TestDocumentsTheApplicantSees(t *testing.T) {
	c := applicantContent()
	for _, tt := range []struct {
		document domain.Document
		closed   bool
		want     bool
	}{
		{domain.Document{Requirement: "transcript", Step: "request"}, false, true},
		{domain.Document{Requirement: "minutes", Step: "ruling"}, false, false},
		{domain.Document{Requirement: "minutes", Step: "ruling"}, true, false},
		{domain.Document{Requirement: "resolution", Step: "ruling"}, false, false},
		{domain.Document{Requirement: "resolution", Step: "ruling"}, true, true},
	} {
		p := domain.Procedure{OpenStep: "ruling"}
		if tt.closed {
			p.OpenStep = ""
		}
		if got := detail.ForApplicant(c, p).Sees(tt.document); got != tt.want {
			t.Errorf("Sees(%s, closed %t) = %t, want %t", tt.document.Requirement, tt.closed, got, tt.want)
		}
	}
}

func TestTheHistoryOfTheApplicant(t *testing.T) {
	history := []domain.Event{
		{Step: "request", By: "applicant", Observation: "Mi pedido"},
		{Step: "review", By: "staff", Observation: "Falta el analítico", Referral: &domain.ReferralRecord{From: "office", To: "commission", ExternalFileNumber: "EX-1"}},
		{Step: "ruling", By: "commission", Document: &domain.DocumentRecord{ID: "D-2", Requirement: "minutes"}},
		{Step: "ruling", By: "commission", Document: &domain.DocumentRecord{ID: "D-3", Requirement: "resolution"}},
		{Step: "ruling", By: "commission", Observation: "Dictamen enviado"},
	}

	seen := detail.ForApplicant(applicantContent(), domain.Procedure{}).History(history)

	if len(seen) != 4 {
		t.Fatalf("history = %+v, want 4 entries: the minutes left out", seen)
	}
	if seen[0].Observation != "Mi pedido" {
		t.Errorf("observation of the applicant = %q, want it kept", seen[0].Observation)
	}
	if seen[1].Observation != "" || seen[1].Referral == nil || seen[1].Referral.ExternalFileNumber != "" || seen[1].Referral.To != "commission" {
		t.Errorf("review = %+v %+v, want the referral without observation nor file number", seen[1], seen[1].Referral)
	}
	if seen[0].By != "applicant" {
		t.Errorf("request = %+v, want the applicant who did it", seen[0])
	}
	if seen[1].By != "" || seen[2].By != "" || seen[1].Step != "review" {
		t.Errorf("steps of staff = %+v, %+v; want the step without who did it", seen[1], seen[2])
	}
	if seen[2].Document == nil || seen[2].Document.ID != "D-3" {
		t.Errorf("third entry = %+v, want the resolution", seen[2])
	}
	if seen[3].Observation != "" {
		t.Errorf("last entry = %+v, want no observation", seen[3])
	}
	if history[1].Referral.ExternalFileNumber != "EX-1" || history[1].Observation == "" {
		t.Error("History changed the events it was given")
	}
}

// A send back to a step of the applicant tells them why, as ADR-025 asks; one
// that stays among staff steps does not.
func TestTheApplicantSeesTheReasonOfASendBackToTheirStep(t *testing.T) {
	history := []domain.Event{
		{Step: "review", By: "staff", Observation: "Falta el analítico", SendBack: true, Next: "request"},
		{Step: "ruling", By: "commission", Observation: "Rever la nota", SendBack: true, Next: "review"},
	}

	seen := detail.ForApplicant(applicantContent(), domain.Procedure{}).History(history)

	if seen[0].Observation != "Falta el analítico" || seen[0].By != "" {
		t.Errorf("send back to the applicant = %+v, want its reason without who did it", seen[0])
	}
	if seen[1].Observation != "" {
		t.Errorf("send back between staff steps = %+v, want no observation", seen[1])
	}
}
