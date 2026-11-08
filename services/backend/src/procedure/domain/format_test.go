package domain

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

// formatOneContent uses every key and every value of format 1.
func formatOneContent() Content {
	return Content{
		Format: 1,
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
				{ID: "request", Name: "Solicitud", Area: "applicant", Action: ActionEnterData, Fields: []string{"full_name", "degree", "subjects"}, OptionalFields: []string{"reason"}, Next: "documents"},
				{ID: "documents", Name: "Documentación", Area: "applicant", Action: ActionUploadDocument, Documents: []string{"id_card"}, OptionalDocuments: []string{"transcript"}, Next: "review"},
				{ID: "review", Name: "Revisión", Area: "office", Profile: ProfileStaffOperative, Action: ActionDecide, Fields: []string{"shift"}, Outcomes: []Outcome{
					{ID: "complete", Name: "Completa", Next: "evaluation"},
					{ID: "incomplete", Name: "Incompleta", Next: "documents"},
				}},
				{ID: "evaluation", Name: "Evaluación", Area: "commission", Profile: ProfileStaffApprover, Action: ActionUploadDocument, Documents: []string{"minutes"}, Next: "signature"},
				{ID: "signature", Name: "Firma", Area: "office", Profile: ProfileStaffApprover, Action: ActionConfirmExternalFact, Fields: []string{"signed_on"}, OptionalFields: []string{"credits"}},
			},
		},
		Areas: []Area{
			{ID: "applicant", Name: "Titular", Kind: AreaApplicant},
			{ID: "commission", Name: "Comisión Curricular", Kind: AreaExternal},
			{ID: "office", Name: "Dirección de Alumnos", Kind: AreaOffice},
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
	if got := Digest(stored); got != digest {
		t.Errorf("Digest() = %s, want %s", got, digest)
	}

	v := Version{ProcedureType: 1, Number: 1, Digest: digest, Encoded: stored}
	got, err := v.Content()
	if err != nil {
		t.Fatalf("Content() = %v", err)
	}
	if want := formatOneContent(); !reflect.DeepEqual(got, want) {
		t.Errorf("Content() =\n%+v\nwant\n%+v", got, want)
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
func formatTwoContent() Content {
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
	version := Version{ProcedureType: 1, Number: 1, Digest: Digest(stored), Encoded: stored}
	got, err := version.Content()
	if err != nil {
		t.Fatalf("Content() = %v", err)
	}
	if want := formatTwoContent(); !reflect.DeepEqual(got, want) {
		t.Errorf("Content() =\n%+v\nwant\n%+v", got, want)
	}
	encoded, err := formatTwoContent().Encode()
	if err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	if !bytes.Equal(encoded, stored) {
		t.Errorf("Encode() =\n%s\nwant\n%s", encoded, stored)
	}
	// The fixture predates ADR-017, which keeps the external kind only for
	// reading published versions: today the same area would be an office one.
	areas := formatTwoContent().Areas
	for i := range areas {
		if areas[i].Kind == AreaExternal {
			areas[i].Kind = AreaOffice
		}
	}
	if err := formatTwoContent().Definition.Validate(areas); err != nil {
		t.Errorf("Validate() = %v, want the fixture to be a valid definition", err)
	}
}
