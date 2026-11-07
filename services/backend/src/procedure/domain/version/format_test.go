package version_test

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

// formatOneContent uses every key and every value of format 1.
func formatOneContent() version.Content {
	return version.Content{
		Format: 1,
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
				{ID: "request", Name: "Solicitud", Area: "applicant", Action: definition.ActionEnterData, Fields: []string{"full_name", "degree", "subjects"}, OptionalFields: []string{"reason"}, Next: "documents"},
				{ID: "documents", Name: "Documentación", Area: "applicant", Action: definition.ActionUploadDocument, Documents: []string{"id_card"}, OptionalDocuments: []string{"transcript"}, Next: "review"},
				{ID: "review", Name: "Revisión", Area: "office", Profile: definition.ProfileStaffOperative, Action: definition.ActionDecide, Fields: []string{"shift"}, Outcomes: []definition.Outcome{
					{ID: "complete", Name: "Completa", Next: "evaluation"},
					{ID: "incomplete", Name: "Incompleta", Next: "documents"},
				}},
				{ID: "evaluation", Name: "Evaluación", Area: "commission", Profile: definition.ProfileStaffApprover, Action: definition.ActionUploadDocument, Documents: []string{"minutes"}, Next: "signature"},
				{ID: "signature", Name: "Firma", Area: "office", Profile: definition.ProfileStaffApprover, Action: definition.ActionConfirmExternalFact, Fields: []string{"signed_on"}, OptionalFields: []string{"credits"}},
			},
		},
		Areas: []definition.Area{
			{ID: "applicant", Name: "Titular", Kind: definition.AreaApplicant},
			{ID: "commission", Name: "Comisión Curricular", Kind: definition.AreaExternal},
			{ID: "office", Name: "Dirección de Alumnos", Kind: definition.AreaOffice},
		},
	}
}

// testdata/format1.json holds the bytes of a format 1 version and is never
// edited: published versions keep those bytes forever. If this test fails,
// the encoding or the meaning of a key changed, and ContentFormat has to go up
// with a new fixture instead.
func TestFormatOne(t *testing.T) {
	stored, err := os.ReadFile("testdata/format1.json")
	if err != nil {
		t.Fatal(err)
	}
	stored = bytes.TrimSuffix(stored, []byte("\n"))
	const digest = "sha256:578100272af5c07ec7b09d123715d58a5547324e5a0fdfd845345e0b96a75ce6"
	if got := version.Digest(stored); got != digest {
		t.Errorf("Digest() = %s, want %s", got, digest)
	}

	v := version.ProcedureTypeVersion{ProcedureTypeID: 1, Number: 1, Digest: digest, EncodedContent: stored}
	got, err := v.DecodeContent()
	if err != nil {
		t.Fatalf("DecodeContent() = %v", err)
	}
	if want := formatOneContent(); !reflect.DeepEqual(got, want) {
		t.Errorf("DecodeContent() =\n%+v\nwant\n%+v", got, want)
	}

	encoded, err := formatOneContent().Encode()
	if err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	if !bytes.Equal(encoded, stored) {
		t.Errorf("Encode() =\n%s\nwant\n%s", encoded, stored)
	}
}

// formatTwoContent is format 1 plus every key that format 2 added.
func formatTwoContent() version.Content {
	content := formatOneContent()
	content.Format = 2
	content.Definition.Description = "Pedí que te reconozcan una materia"
	content.Definition.Deadlines = []string{"Hasta 30 días antes de cursar"}
	content.Definition.Admissibility = []string{"Ser estudiante regular"}
	for index := range content.Definition.Documents {
		content.Definition.Documents[index].AcceptedTypes = []string{"application/pdf", "image/png"}
		content.Definition.Documents[index].MaxSize = 10 << 20
	}
	return content
}

// testdata/format2.json follows the same rule as format1.json.
func TestFormatTwo(t *testing.T) {
	stored, err := os.ReadFile("testdata/format2.json")
	if err != nil {
		t.Fatal(err)
	}
	stored = bytes.TrimSuffix(stored, []byte("\n"))
	v := version.ProcedureTypeVersion{ProcedureTypeID: 1, Number: 1, Digest: version.Digest(stored), EncodedContent: stored}
	got, err := v.DecodeContent()
	if err != nil {
		t.Fatalf("DecodeContent() = %v", err)
	}
	if want := formatTwoContent(); !reflect.DeepEqual(got, want) {
		t.Errorf("DecodeContent() =\n%+v\nwant\n%+v", got, want)
	}
	encoded, err := formatTwoContent().Encode()
	if err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	if !bytes.Equal(encoded, stored) {
		t.Errorf("Encode() =\n%s\nwant\n%s", encoded, stored)
	}
	if err := formatTwoContent().Definition.Validate(formatTwoContent().Areas); err != nil {
		t.Errorf("Validate() = %v, want the fixture to be a valid definition", err)
	}
}
