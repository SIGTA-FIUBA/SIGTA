package procedure

import "testing"

func TestApplicantRequirements(t *testing.T) {
	c := NewContent(Definition{
		Fields: []Field{
			{ID: "shared", Name: "Dato", Type: FieldText},
			{ID: "note", Name: "Nota", Type: FieldText},
			{ID: "both", Name: "En dos pasos", Type: FieldText},
			{ID: "staff", Name: "Del personal", Type: FieldText},
		},
		Documents: []Document{{ID: "shared", Name: "Documento"}, {ID: "minutes", Name: "Acta"}},
		Steps: []Step{
			{ID: "data", Area: "applicant", Fields: []string{"shared"}, OptionalFields: []string{"note", "both"}},
			{ID: "again", Area: "applicant", Fields: []string{"both"}},
			{ID: "documents", Area: "applicant", OptionalDocuments: []string{"shared"}},
			{ID: "review", Area: "office", Fields: []string{"staff"}, Documents: []string{"minutes"}},
		},
	}, testAreas())

	fields, documents := c.ApplicantRequirements()

	wantFields := []struct {
		id       string
		required bool
	}{{"shared", true}, {"note", false}, {"both", true}}
	if len(fields) != len(wantFields) {
		t.Fatalf("fields = %+v, want %+v", fields, wantFields)
	}
	for i, want := range wantFields {
		if fields[i].Field.ID != want.id || fields[i].Required != want.required {
			t.Errorf("fields[%d] = %+v, want %+v", i, fields[i], want)
		}
	}
	// A document can share its id with a field and still be optional.
	if len(documents) != 1 || documents[0].Document.ID != "shared" || documents[0].Required {
		t.Errorf("documents = %+v, want only an optional shared", documents)
	}
}
