package catalog_test

import (
	"testing"

	"backend/src/procedure/domain/catalog"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/definition/definitiontest"
	"backend/src/procedure/domain/version"
)

func TestApplicantRequirements(t *testing.T) {
	content := version.NewContent(definition.ProcedureTypeDefinition{
		Fields: []definition.Field{
			{ID: "shared", Name: "Dato", Type: definition.FieldText},
			{ID: "note", Name: "Nota", Type: definition.FieldText},
			{ID: "both", Name: "En dos pasos", Type: definition.FieldText},
			{ID: "staff", Name: "Del personal", Type: definition.FieldText},
		},
		Documents: []definition.Document{{ID: "shared", Name: "Documento"}, {ID: "minutes", Name: "Acta"}},
		Steps: []definition.Step{
			{ID: "data", Area: "applicant", Fields: []string{"shared"}, OptionalFields: []string{"note", "both"}},
			{ID: "again", Area: "applicant", Fields: []string{"both"}, OptionalFields: []string{"shared"}},
			{ID: "documents", Area: "applicant", OptionalDocuments: []string{"shared"}},
			{ID: "review", Area: "office", Fields: []string{"staff"}, Documents: []string{"minutes"}},
		},
	}, definitiontest.Areas())

	fields, documents := catalog.ApplicantRequirements(content)

	wantFields := []struct {
		id       string
		required bool
	}{{"shared", true}, {"note", false}, {"both", true}}
	if len(fields) != len(wantFields) {
		t.Fatalf("fields = %+v, want %+v", fields, wantFields)
	}
	for index, want := range wantFields {
		if fields[index].Field.ID != want.id || fields[index].Required != want.required {
			t.Errorf("fields[%d] = %+v, want %+v", index, fields[index], want)
		}
	}
	// A document can share its id with a field and still be optional.
	if len(documents) != 1 || documents[0].Document.ID != "shared" || documents[0].Required {
		t.Errorf("documents = %+v, want only an optional shared", documents)
	}
}
