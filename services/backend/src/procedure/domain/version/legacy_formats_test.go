package version_test

import (
	"bytes"
	"io/fs"
	"os"
	"reflect"
	"testing"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

// formatOneContent is what testdata/format1.json reads as today: format 1
// used every key and every value of its time, and is converted when read
// (legacy_formats.go). The applicant had an area of kind "applicant", and the
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

// formatThreeContent uses every key of format 3, read as current content: the
// role of each step, areas only for staff, and the permissions of the area
// that each staff step asks for.
func formatThreeContent() version.Content {
	return formatTwoContent()
}

// formatFourContent is format 3 plus the key that format 4 added: an outcome
// that is a send back.
func formatFourContent() version.Content {
	content := formatThreeContent()
	content.Definition.Steps[2].Outcomes[1].SendBack = true
	return content
}

// formatFiveContent is format 4 plus the key that format 5 added: a document
// that is the resolution of the procedure.
func formatFiveContent() version.Content {
	content := formatFourContent()
	content.Definition.Documents[2].Resolution = true
	return content
}

// readFixture returns the bytes of testdata/format<n>.json. Those files hold
// versions as they were stored and are never edited: published versions keep
// their bytes forever. If a test fails, the encoding or the meaning of a key
// changed, and ContentFormat has to go up with a new fixture instead.
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	stored, err := fs.ReadFile(os.DirFS("testdata"), name)
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

// expectCurrent checks that an older format from format 3 on reads with the
// current shape, as the current format, and that this shape still encodes it
// byte for byte.
func expectCurrent(t *testing.T, name string, format int, want version.Content) {
	t.Helper()
	stored := readFixture(t, name)
	got, err := version.ProcedureTypeVersion{ProcedureTypeID: 1, Number: 1, Digest: version.Digest(stored), EncodedContent: stored}.DecodeContent()
	if err != nil {
		t.Fatalf("DecodeContent() = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DecodeContent() =\n%+v\nwant\n%+v", got, want)
	}
	got.Format = format
	encoded, err := got.Encode()
	if err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	if !bytes.Equal(encoded, stored) {
		t.Errorf("Encode() =\n%s\nwant\n%s", encoded, stored)
	}
	if err := want.Definition.Validate(want.Areas); err != nil {
		t.Errorf("Validate() = %v, want the fixture to be a valid definition", err)
	}
}

func TestFormatThree(t *testing.T) {
	expectCurrent(t, "format3.json", 3, formatThreeContent())
}

func TestFormatFour(t *testing.T) {
	expectCurrent(t, "format4.json", 4, formatFourContent())
}

func TestFormatFive(t *testing.T) {
	expectCurrent(t, "format5.json", 5, formatFiveContent())
}

// A format 3 version cannot carry a send back: format 4 added it.
func TestFormatThreeWithSendBackIsRejected(t *testing.T) {
	stored := bytes.Replace(readFixture(t, "format3.json"), []byte(`"next":"documents"}`), []byte(`"next":"documents","send_back":true}`), 1)
	if bytes.Equal(stored, readFixture(t, "format3.json")) {
		t.Fatal("the fixture has no outcome to mark")
	}
	if _, err := (version.ProcedureTypeVersion{EncodedContent: stored}).DecodeContent(); err == nil {
		t.Error("DecodeContent() = nil, want an error")
	}
}

// A version before format 5 cannot carry a resolution: format 5 added it.
func TestFormatBeforeFiveWithResolutionIsRejected(t *testing.T) {
	withResolution := []byte(`"max_size":10485760,"resolution":true}`)
	for _, encoded := range [][]byte{
		bytes.Replace(readFixture(t, "format5.json"), []byte(`{"format":5,`), []byte(`{"format":4,`), 1),
		bytes.Replace(readFixture(t, "format3.json"), []byte(`"max_size":10485760}`), withResolution, 1),
		bytes.Replace(readFixture(t, "format2.json"), []byte(`"max_size":10485760}`), withResolution, 1),
	} {
		if !bytes.Contains(encoded, []byte(`"resolution":true`)) {
			t.Fatalf("%s does not carry a resolution", encoded)
		}
		if _, err := (version.ProcedureTypeVersion{EncodedContent: encoded}).DecodeContent(); err == nil {
			t.Errorf("DecodeContent(%s) = nil, want an error", encoded)
		}
	}
}

// A legacy version cannot carry a role: format 3 added it.
func TestLegacyFormatWithRoleIsRejected(t *testing.T) {
	encoded := []byte(`{"format":2,"definition":{"name":"X","steps":[{"id":"a","name":"A","role":"staff","area":"office","action":"decide"}]},"areas":[{"id":"office","name":"Oficina","kind":"office"}]}`)
	if _, err := (version.ProcedureTypeVersion{EncodedContent: encoded}).DecodeContent(); err == nil {
		t.Error("DecodeContent() = nil, want an error")
	}
}
