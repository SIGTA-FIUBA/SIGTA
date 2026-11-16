package submission

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"backend/src/procedure/domain/definition"
)

func testDefinition() definition.ProcedureTypeDefinition {
	return definition.ProcedureTypeDefinition{
		Fields: []definition.Field{
			{ID: "reason", Type: definition.FieldLongText},
			{ID: "credits", Type: definition.FieldNumber},
			{ID: "since", Type: definition.FieldDate},
			{ID: "modality", Type: definition.FieldOption, Options: []definition.Option{{ID: "full", Name: "Full"}, {ID: "partial", Name: "Partial"}}},
			{ID: "subjects", Type: definition.FieldSubjectList},
			{ID: "comment", Type: definition.FieldText},
			{ID: "grade", Type: definition.FieldNumber},
		},
		Documents: []definition.Document{{ID: "transcript"}, {ID: "program"}, {ID: "ruling"}},
		Steps: []definition.Step{
			{ID: "request", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Fields: []string{"reason", "credits", "since", "modality", "subjects"}, OptionalFields: []string{"comment"}, Next: "documents"},
			{ID: "documents", Role: definition.RoleApplicant, Action: definition.ActionUploadDocument, Documents: []string{"transcript"}, OptionalDocuments: []string{"program"}, Next: "review"},
			{ID: "review", Role: definition.RoleStaff, Area: "office", Action: definition.ActionEnterData, Fields: []string{"grade"}, Next: "ruling"},
			{ID: "ruling", Role: definition.RoleStaff, Area: "office", Action: definition.ActionUploadDocument, Documents: []string{"ruling"}, Next: "late"},
			{ID: "late", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Fields: []string{"comment"}},
		},
	}
}

func complete() map[string]Value {
	return map[string]Value{
		"reason":   {Text: " Cursé afuera "},
		"credits":  {Text: "12.5"},
		"since":    {Text: "2026-03-01"},
		"modality": {Text: "full"},
		"subjects": {Items: []string{" Análisis I", "Física I "}},
	}
}

func TestOpeningStepsStopAtTheFirstStepOfTheStaff(t *testing.T) {
	def := testDefinition()
	var ids []string
	for _, s := range OpeningSteps(def) {
		ids = append(ids, s.ID)
	}
	if !slices.Equal(ids, []string{"request", "documents"}) {
		t.Errorf("OpeningSteps() = %v, want [request documents]", ids)
	}
	if got := OpenStep(def); got != "review" {
		t.Errorf("OpenStep() = %q, want review", got)
	}
}

func TestOpenStepWhenTheStaffOpenTheProcedure(t *testing.T) {
	def := testDefinition()
	def.Steps = def.Steps[2:]
	if got := OpenStep(def); got != "review" {
		t.Errorf("OpenStep() = %q, want review", got)
	}
}

func TestRequirementsAreOnlyThoseOfTheOpeningSteps(t *testing.T) {
	fields, documents := Requirements(testDefinition())
	var got []string
	for _, f := range fields {
		got = append(got, f.Field.ID+"@"+f.Step)
	}
	want := []string{"reason@request", "credits@request", "since@request", "modality@request", "subjects@request", "comment@request"}
	if !slices.Equal(got, want) {
		t.Errorf("fields = %v, want %v", got, want)
	}
	if len(documents) != 2 || !documents[0].Required || documents[1].Required {
		t.Errorf("documents = %+v, want a mandatory transcript and an optional program", documents)
	}
}

func TestCheckValuesTrimsWhatItKeeps(t *testing.T) {
	values, problems := CheckValues(testDefinition(), complete())
	if problems != nil {
		t.Fatalf("CheckValues() problems = %v", problems)
	}
	if values["reason"].Text != "Cursé afuera" {
		t.Errorf("reason = %q", values["reason"].Text)
	}
	if !slices.Equal(values["subjects"].Items, []string{"Análisis I", "Física I"}) {
		t.Errorf("subjects = %q", values["subjects"].Items)
	}
	if _, ok := values["comment"]; ok {
		t.Error("an optional field left out has a value")
	}
}

func TestCheckValuesRejects(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value *Value
		want  Problem
	}{
		{"a missing mandatory field", "reason", nil, FieldProblem(CodeMissing, "reason")},
		{"blank text", "reason", &Value{Text: "  "}, FieldProblem(CodeInvalid, "reason")},
		{"text too long", "comment", &Value{Text: strings.Repeat("a", MaxText+1)}, FieldProblem(CodeInvalid, "comment")},
		{"a word for a number", "credits", &Value{Text: "doce"}, FieldProblem(CodeInvalid, "credits")},
		{"a number with an exponent", "credits", &Value{Text: "1e3"}, FieldProblem(CodeInvalid, "credits")},
		{"a date in another format", "since", &Value{Text: "31/12/2026"}, FieldProblem(CodeInvalid, "since")},
		{"a date that does not exist", "since", &Value{Text: "2026-02-30"}, FieldProblem(CodeInvalid, "since")},
		{"the name of an option", "modality", &Value{Text: "Full"}, FieldProblem(CodeInvalid, "modality")},
		{"an empty list", "subjects", &Value{Items: []string{}}, FieldProblem(CodeInvalid, "subjects")},
		{"a blank subject", "subjects", &Value{Items: []string{"Física I", " "}}, FieldProblem(CodeInvalid, "subjects")},
		{"a list for a text", "reason", &Value{Items: []string{"a"}}, FieldProblem(CodeInvalid, "reason")},
		{"a field of the staff", "grade", &Value{Text: "10"}, FieldProblem(CodeNotAsked, "grade")},
		{"a field that does not exist", "unknown", &Value{Text: "x"}, FieldProblem(CodeNotAsked, "unknown")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := complete()
			if tt.value == nil {
				delete(values, tt.field)
			} else {
				values[tt.field] = *tt.value
			}
			got, problems := CheckValues(testDefinition(), values)
			if got != nil || !slices.Equal(problems, []Problem{tt.want}) {
				t.Errorf("CheckValues() = %v, %v, want only %v", got, problems, tt.want)
			}
		})
	}
}

func TestMissingDocuments(t *testing.T) {
	_, documents := Requirements(testDefinition())
	if got := MissingDocuments(documents, map[string]bool{"program": true}); !slices.Equal(got, []Problem{DocumentProblem(CodeMissing, "transcript")}) {
		t.Errorf("MissingDocuments() = %v", got)
	}
	if got := MissingDocuments(documents, map[string]bool{"transcript": true}); got != nil {
		t.Errorf("MissingDocuments() = %v, want none", got)
	}
}

func TestRejectMatchesErrInvalid(t *testing.T) {
	if Reject(nil) != nil {
		t.Error("Reject(nil) is not nil")
	}
	err := Reject([]Problem{FieldProblem(CodeMissing, "reason")})
	var rejected *Error
	if !errors.Is(err, ErrInvalid) || !errors.As(err, &rejected) || len(rejected.Problems) != 1 {
		t.Errorf("Reject() = %v", err)
	}
}
