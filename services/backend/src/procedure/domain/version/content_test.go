package version_test

import (
	"slices"
	"testing"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/definition/definitiontest"
	"backend/src/procedure/domain/version"
)

func encode(t *testing.T, d definition.ProcedureTypeDefinition, areas []definition.Area) []byte {
	t.Helper()
	encoded, err := version.NewContent(d, areas).Encode()
	if err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	return encoded
}

func TestContentCarriesOnlyUsedAreas(t *testing.T) {
	content := version.NewContent(definitiontest.Definition(), definitiontest.Areas())
	if ids := areaIDs(content.Areas); !slices.Equal(ids, []string{"office"}) {
		t.Errorf("Areas = %v, want only office: the applicant has no area", ids)
	}

	d := definitiontest.Definition()
	d.Steps[3].Area = "commission"
	content = version.NewContent(d, definitiontest.Areas())
	if ids := areaIDs(content.Areas); !slices.Equal(ids, []string{"commission", "office"}) {
		t.Errorf("Areas = %v, want the office of the decision too", ids)
	}
}

func areaIDs(areas []definition.Area) []string {
	var ids []string
	for _, a := range areas {
		ids = append(ids, a.ID)
	}
	return ids
}

func TestDigest(t *testing.T) {
	base := version.Digest(encode(t, definitiontest.Definition(), definitiontest.Areas()))

	unused := append(definitiontest.Areas(), definition.Area{ID: "front_desk", Name: "Mesa de Entradas"})
	if got := version.Digest(encode(t, definitiontest.Definition(), unused)); got != base {
		t.Errorf("an unused area changed the digest")
	}

	reordered := definitiontest.Areas()
	reordered[0], reordered[1] = reordered[1], reordered[0]
	if got := version.Digest(encode(t, definitiontest.Definition(), reordered)); got != base {
		t.Errorf("reordering the areas changed the digest")
	}

	renamed := definitiontest.Areas()
	renamed[0].Name = "Dirección General de Alumnos"
	if got := version.Digest(encode(t, definitiontest.Definition(), renamed)); got == base {
		t.Errorf("renaming a used area kept the digest")
	}

	d := definitiontest.Definition()
	d.Documents[0].Name = "Carta al Decano"
	if got := version.Digest(encode(t, d, definitiontest.Areas())); got == base {
		t.Errorf("renaming a document kept the digest")
	}
}
