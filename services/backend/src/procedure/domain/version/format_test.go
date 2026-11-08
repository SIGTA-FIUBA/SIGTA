package version_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

// formatOneContent is what testdata/format1.json reads as today: format 1
// used every key and every value of its time, and is converted when read
// (legacy.go). The applicant had an area of kind "applicant", and the
// commission one of kind "external", which is staff.
func formatOneContent() version.Content {
	return version.Content{
		Format: version.ContentFormat,
		Definition: definition.ProcedureTypeDefinition{
			Name: "Equivalencias & <pases>",
			Fields: []definition.Field{
				{ID: "full_name", Name: "Nombre completo", Type: definition.FieldText},
				{ID: "reason", Name: "Motivo", Type: definition.FieldLongText},
				{ID: "credits", Name: "Créditos", Type: definition.FieldNumber},
				{ID: "signed_on", Name: "Fecha de firma", Type: definition.FieldDate},
				{ID: "shift", Name: "Turno", Type: definition.FieldOption, Options: []definition.Option{{ID: "morning", Name: "Mañana"}, {ID: "evening", Name: "Noche"}}},
				{ID: "degree", Name: "Carrera", Type: definition.FieldDegree},
				{ID: "subjects", Name: "Materias", Type: definition.FieldSubjectList},
			},
			Documents: []definition.Document{
				{ID: "id_card", Name: "Documento de identidad"},
				{ID: "transcript", Name: "Certificado analítico"},
				{ID: "minutes", Name: "Acta"},
			},
			Steps: []definition.Step{
				{ID: "request", Name: "Solicitud", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Fields: []string{"full_name", "degree", "subjects"}, OptionalFields: []string{"reason"}, Next: "documents"},
				{ID: "documents", Name: "Documentación", Role: definition.RoleApplicant, Action: definition.ActionUploadDocument, Documents: []string{"id_card"}, OptionalDocuments: []string{"transcript"}, Next: "review"},
				{ID: "review", Name: "Revisión", Role: definition.RoleStaff, Area: "office", Permissions: []string{"staff_operative"}, Action: definition.ActionDecide, Fields: []string{"shift"}, Outcomes: []definition.Outcome{
					{ID: "complete", Name: "Completa", Next: "evaluation"},
					{ID: "incomplete", Name: "Incompleta", Next: "documents"},
				}},
				{ID: "evaluation", Name: "Evaluación", Role: definition.RoleStaff, Area: "commission", Permissions: []string{"staff_approver"}, Action: definition.ActionUploadDocument, Documents: []string{"minutes"}, Next: "signature"},
				{ID: "signature", Name: "Firma", Role: definition.RoleStaff, Area: "office", Permissions: []string{"staff_approver"}, Action: definition.ActionConfirmExternalFact, Fields: []string{"signed_on"}, OptionalFields: []string{"credits"}},
			},
		},
		Areas: []definition.Area{
			{ID: "commission", Name: "Comisión Curricular", Permissions: []definition.Permission{{ID: "staff_approver", Name: "Aprobación"}}},
			{ID: "office", Name: "Dirección de Alumnos", Permissions: []definition.Permission{{ID: "staff_approver", Name: "Aprobación"}, {ID: "staff_operative", Name: "Operativo"}}},
		},
	}
}

// formatTwoContent is format 1 plus every key that format 2 added.
func formatTwoContent() version.Content {
	content := formatOneContent()
	content.Definition.Description = "Pedí que te reconozcan una materia"
	content.Definition.Deadlines = []string{"Hasta 30 días antes de cursar"}
	content.Definition.Admissibility = []string{"Ser estudiante regular"}
	for index := range content.Definition.Documents {
		content.Definition.Documents[index].AcceptedTypes = []string{"application/pdf", "image/png"}
		content.Definition.Documents[index].MaxSizeBytes = 10 << 20
	}
	return content
}

// formatThreeContent uses every key of format 3: the role of each step, areas
// only for staff, and the permissions of the area that each staff step asks
// for.
func formatThreeContent() version.Content {
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
func expectLegacy(t *testing.T, stored []byte, want version.Content) {
	t.Helper()
	again, err := version.ReencodeLegacy(stored)
	if err != nil {
		t.Fatalf("decode = %v", err)
	}
	if !bytes.Equal(again, stored) {
		t.Errorf("legacyContent encodes as\n%s\nwant\n%s", again, stored)
	}
	got, err := version.ProcedureTypeVersion{ProcedureTypeID: 1, Number: 1, Digest: version.Digest(stored), EncodedContent: stored}.DecodeContent()
	if err != nil {
		t.Fatalf("DecodeContent() = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DecodeContent() =\n%+v\nwant\n%+v", got, want)
	}
}

func TestFormatOne(t *testing.T) {
	stored := readFixture(t, "format1.json")
	const digest = "sha256:578100272af5c07ec7b09d123715d58a5547324e5a0fdfd845345e0b96a75ce6"
	if got := version.Digest(stored); got != digest {
		t.Errorf("version.Digest() = %s, want %s", got, digest)
	}
	expectLegacy(t, stored, formatOneContent())
}

func TestFormatTwo(t *testing.T) {
	expectLegacy(t, readFixture(t, "format2.json"), formatTwoContent())
}

func TestFormatThree(t *testing.T) {
	stored := readFixture(t, "format3.json")
	got, err := version.ProcedureTypeVersion{ProcedureTypeID: 1, Number: 1, Digest: version.Digest(stored), EncodedContent: stored}.DecodeContent()
	if err != nil {
		t.Fatalf("DecodeContent() = %v", err)
	}
	if want := formatThreeContent(); !reflect.DeepEqual(got, want) {
		t.Errorf("DecodeContent() =\n%+v\nwant\n%+v", got, want)
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
	if _, err := (version.ProcedureTypeVersion{EncodedContent: encoded}).DecodeContent(); err == nil {
		t.Error("DecodeContent() = nil, want an error")
	}
}
