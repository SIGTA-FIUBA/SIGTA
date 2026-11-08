package domain

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// formatOneContent is what testdata/format1.json reads as today: format 1
// used every key and every value of its time, and is converted when read
// (legacy.go). The applicant had an area of kind "applicant", and the
// commission one of kind "external", which is staff.
func formatOneContent() Content {
	return Content{
		Format: ContentFormat,
		Definition: Definition{
			Name: "Equivalencias & <pases>",
			Fields: []Field{
				{ID: "full_name", Name: "Nombre completo", Type: FieldText},
				{ID: "reason", Name: "Motivo", Type: FieldLongText},
				{ID: "credits", Name: "Créditos", Type: FieldNumber},
				{ID: "signed_on", Name: "Fecha de firma", Type: FieldDate},
				{ID: "shift", Name: "Turno", Type: FieldOption, Options: []Option{{ID: "morning", Name: "Mañana"}, {ID: "evening", Name: "Noche"}}},
				{ID: "degree", Name: "Carrera", Type: FieldDegree},
				{ID: "subjects", Name: "Materias", Type: FieldSubjectList},
			},
			Documents: []Document{
				{ID: "id_card", Name: "Documento de identidad"},
				{ID: "transcript", Name: "Certificado analítico"},
				{ID: "minutes", Name: "Acta"},
			},
			Steps: []Step{
				{ID: "request", Name: "Solicitud", Role: RoleApplicant, Action: ActionEnterData, Fields: []string{"full_name", "degree", "subjects"}, OptionalFields: []string{"reason"}, Next: "documents"},
				{ID: "documents", Name: "Documentación", Role: RoleApplicant, Action: ActionUploadDocument, Documents: []string{"id_card"}, OptionalDocuments: []string{"transcript"}, Next: "review"},
				{ID: "review", Name: "Revisión", Role: RoleStaff, Area: "office", Profile: ProfileStaffOperative, Action: ActionDecide, Fields: []string{"shift"}, Outcomes: []Outcome{
					{ID: "complete", Name: "Completa", Next: "evaluation"},
					{ID: "incomplete", Name: "Incompleta", Next: "documents"},
				}},
				{ID: "evaluation", Name: "Evaluación", Role: RoleStaff, Area: "commission", Profile: ProfileStaffApprover, Action: ActionUploadDocument, Documents: []string{"minutes"}, Next: "signature"},
				{ID: "signature", Name: "Firma", Role: RoleStaff, Area: "office", Profile: ProfileStaffApprover, Action: ActionConfirmExternalFact, Fields: []string{"signed_on"}, OptionalFields: []string{"credits"}},
			},
		},
		Areas: []Area{
			{ID: "commission", Name: "Comisión Curricular"},
			{ID: "office", Name: "Dirección de Alumnos"},
		},
	}
}

// formatTwoContent is format 1 plus every key that format 2 added.
func formatTwoContent() Content {
	content := formatOneContent()
	content.Definition.Description = "Pedí que te reconozcan una materia"
	content.Definition.Deadlines = []string{"Hasta 30 días antes de cursar"}
	content.Definition.Admissibility = []string{"Ser estudiante regular"}
	for index := range content.Definition.Documents {
		content.Definition.Documents[index].AcceptedTypes = []string{"application/pdf", "image/png"}
		content.Definition.Documents[index].MaxSize = 10 << 20
	}
	return content
}

// formatThreeContent uses every key of format 3: the role of each step, and
// areas only for staff.
func formatThreeContent() Content {
	return formatTwoContent()
}

// readFixture returns the bytes of testdata/format<n>.json. Those files hold
// versions as they were stored and are never edited: published versions keep
// their bytes forever. If a test fails, the encoding or the meaning of a key
// changed, and ContentFormat has to go up with a new fixture instead.
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	stored, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.TrimSuffix(stored, []byte("\n"))
}

// expectLegacy checks that a format before 3 is still described exactly by
// legacyContent, and that it reads as want.
func expectLegacy(t *testing.T, stored []byte, want Content) {
	t.Helper()
	var legacy legacyContent
	if err := decodeStrict(stored, &legacy); err != nil {
		t.Fatalf("decode = %v", err)
	}
	again, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, stored) {
		t.Errorf("legacyContent encodes as\n%s\nwant\n%s", again, stored)
	}
	got, err := Version{ProcedureType: 1, Number: 1, Digest: Digest(stored), Encoded: stored}.Content()
	if err != nil {
		t.Fatalf("Content() = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Content() =\n%+v\nwant\n%+v", got, want)
	}
}

func TestFormatOne(t *testing.T) {
	stored := readFixture(t, "format1.json")
	const digest = "sha256:578100272af5c07ec7b09d123715d58a5547324e5a0fdfd845345e0b96a75ce6"
	if got := Digest(stored); got != digest {
		t.Errorf("Digest() = %s, want %s", got, digest)
	}
	expectLegacy(t, stored, formatOneContent())
}

func TestFormatTwo(t *testing.T) {
	expectLegacy(t, readFixture(t, "format2.json"), formatTwoContent())
}

func TestFormatThree(t *testing.T) {
	stored := readFixture(t, "format3.json")
	got, err := Version{ProcedureType: 1, Number: 1, Digest: Digest(stored), Encoded: stored}.Content()
	if err != nil {
		t.Fatalf("Content() = %v", err)
	}
	if want := formatThreeContent(); !reflect.DeepEqual(got, want) {
		t.Errorf("Content() =\n%+v\nwant\n%+v", got, want)
	}
	encoded, err := formatThreeContent().Encode()
	if err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	if !bytes.Equal(encoded, stored) {
		t.Errorf("Encode() =\n%s\nwant\n%s", encoded, stored)
	}
	if err := formatThreeContent().Definition.Validate(formatThreeContent().Areas); err != nil {
		t.Errorf("Validate() = %v, want the fixture to be a valid definition", err)
	}
}

// A legacy version cannot carry a role: format 3 added it.
func TestLegacyFormatWithRoleIsRejected(t *testing.T) {
	encoded := []byte(`{"format":2,"definition":{"name":"X","steps":[{"id":"a","name":"A","role":"staff","area":"office","action":"decide"}]},"areas":[{"id":"office","name":"Oficina","kind":"office"}]}`)
	if _, err := (Version{Encoded: encoded}).Content(); err == nil {
		t.Error("Content() = nil, want an error")
	}
}
