package engine_test

import (
	"errors"
	"slices"
	"testing"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
)

var withRequirements = func() definition.ProcedureTypeDefinition {
	d := equivalence
	d.Steps = slices.Clone(d.Steps)
	d.Steps[0].Fields, d.Steps[0].OptionalFields, d.Steps[0].Documents = []string{"subjects"}, []string{"plan"}, []string{"transcript"}
	d.Steps[2].Documents = []string{"ruling"}
	d.Fields = []definition.Field{{ID: "subjects"}, {ID: "plan"}, {ID: "grades"}}
	d.Documents = []definition.Document{{ID: "transcript"}, {ID: "ruling"}}
	return d
}()

func correction(step string, fields, documents []string) domain.CorrectionRequest {
	return domain.CorrectionRequest{Step: step, By: "staff@fi.uba.ar", At: monday, Reason: "the codes are missing", Fields: fields, Documents: documents}
}

func TestRequestCorrectionRecordsWhatWhyWhoAndWhenWithoutMoving(t *testing.T) {
	asked := correction("file", []string{"subjects"}, []string{"transcript"})
	request, event, err := engine.RequestCorrection(on("file"), withRequirements, asked)
	if err != nil {
		t.Fatal(err)
	}
	if request.Step != "file" || request.Reason != asked.Reason || !slices.Equal(request.Fields, asked.Fields) || !slices.Equal(request.Documents, asked.Documents) {
		t.Errorf("request = %+v, want %+v", request, asked)
	}
	record := event.CorrectionRequest
	if event.ProcedureID != "P-1" || event.Step != "file" || event.By != "staff@fi.uba.ar" || !event.At.Equal(monday) || event.Observation != asked.Reason || event.Next != "" ||
		record == nil || !slices.Equal(record.Fields, []string{"subjects"}) || !slices.Equal(record.Documents, []string{"transcript"}) {
		t.Errorf("event = %+v, want the request on file with its reason and requirements", event)
	}
}

func TestRequestCorrectionWhileReferred(t *testing.T) {
	p := on("evaluate")
	p.Referral = &domain.Referral{From: "affairs", To: "commission", Since: monday}
	if _, _, err := engine.RequestCorrection(p, withRequirements, correction("evaluate", []string{"plan"}, nil)); err != nil {
		t.Errorf("RequestCorrection = %v, want it recorded", err)
	}
}

func TestRequestCorrectionRejects(t *testing.T) {
	noReason := correction("file", []string{"subjects"}, nil)
	noReason.Reason = "  "
	tests := []struct {
		name      string
		procedure domain.Procedure
		request   domain.CorrectionRequest
		want      error
	}{
		{"a step that is not open", on("file"), correction("approve", []string{"subjects"}, nil), engine.ErrStepNotOpen},
		{"a closed procedure", on(""), correction("approve", []string{"subjects"}, nil), engine.ErrStepNotOpen},
		{"a step of the applicant", on("request"), correction("request", []string{"subjects"}, nil), engine.ErrCorrectionOnApplicant},
		{"no reason", on("file"), noReason, engine.ErrNoReason},
		{"nothing asked for", on("file"), correction("file", nil, nil), engine.ErrNothingRequested},
		{"a field the type does not have", on("file"), correction("file", []string{"age"}, nil), engine.ErrNotAnApplicantRequirement},
		{"a field no step asks for", on("file"), correction("file", []string{"grades"}, nil), engine.ErrNotAnApplicantRequirement},
		{"a document only a staff step asks for", on("file"), correction("file", nil, []string{"ruling"}), engine.ErrNotAnApplicantRequirement},
		{"a document asked for as a field", on("file"), correction("file", []string{"transcript"}, nil), engine.ErrNotAnApplicantRequirement},
		{"a field asked for as a document", on("file"), correction("file", nil, []string{"subjects"}), engine.ErrNotAnApplicantRequirement},
		{"the same field twice", on("file"), correction("file", []string{"subjects", "subjects"}, nil), engine.ErrRepeatedRequirement},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := engine.RequestCorrection(tt.procedure, withRequirements, tt.request); !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestConfirmWaitsForThePendingCorrections(t *testing.T) {
	p := on("file")
	p.CorrectionRequests = []domain.CorrectionRequest{correction("file", []string{"subjects"}, nil)}
	moved, _, err := engine.Confirm(p, equivalence, confirm("file", ""))
	if !errors.Is(err, engine.ErrAwaitingApplicant) || !errors.Is(err, engine.ErrInvalidConfirmation) {
		t.Fatalf("err = %v, want ErrAwaitingApplicant", err)
	}
	if moved.OpenStep != "file" {
		t.Errorf("open step = %q, want it unchanged", moved.OpenStep)
	}
}
