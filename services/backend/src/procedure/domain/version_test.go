package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func encode(t *testing.T, d Definition, areas []Area) []byte {
	t.Helper()
	encoded, err := NewContent(d, areas).Encode()
	if err != nil {
		t.Fatalf("Encode() = %v", err)
	}
	return encoded
}

func testVersion(t *testing.T, number int, d Definition) Version {
	t.Helper()
	encoded := encode(t, d, testAreas())
	return Version{ProcedureType: 1, Number: number, Digest: Digest(encoded), Encoded: encoded}
}

func TestContentCarriesOnlyUsedAreas(t *testing.T) {
	content := NewContent(testDefinition(), testAreas())
	if ids := areaIDs(content.Areas); !slices.Equal(ids, []string{"applicant", "office"}) {
		t.Errorf("Areas = %v, want applicant and office", ids)
	}

	d := testDefinition()
	d.Steps[3].Area = "commission"
	content = NewContent(d, testAreas())
	if ids := areaIDs(content.Areas); !slices.Equal(ids, []string{"applicant", "commission", "office"}) {
		t.Errorf("Areas = %v, want the office of the decision too", ids)
	}
}

func areaIDs(areas []Area) []string {
	var ids []string
	for _, a := range areas {
		ids = append(ids, a.ID)
	}
	return ids
}

func TestDigest(t *testing.T) {
	base := Digest(encode(t, testDefinition(), testAreas()))

	unused := append(testAreas(), Area{ID: "front_desk", Name: "Mesa de Entradas", Kind: AreaOffice})
	if got := Digest(encode(t, testDefinition(), unused)); got != base {
		t.Errorf("an unused area changed the digest")
	}

	reordered := testAreas()
	reordered[0], reordered[1] = reordered[1], reordered[0]
	if got := Digest(encode(t, testDefinition(), reordered)); got != base {
		t.Errorf("reordering the areas changed the digest")
	}

	renamed := testAreas()
	renamed[1].Name = "Dirección General de Alumnos"
	if got := Digest(encode(t, testDefinition(), renamed)); got == base {
		t.Errorf("renaming a used area kept the digest")
	}

	d := testDefinition()
	d.Documents[0].Name = "Carta al Decano"
	if got := Digest(encode(t, d, testAreas())); got == base {
		t.Errorf("renaming a document kept the digest")
	}
}

func TestVerifyDetectsAlteredContent(t *testing.T) {
	v := testVersion(t, 1, testDefinition())
	if err := v.Verify(); err != nil {
		t.Fatalf("Verify() = %v, want nil", err)
	}
	v.Encoded = []byte(strings.Replace(string(v.Encoded), "Solicitud", "Pedido", 1))
	if err := v.Verify(); !errors.Is(err, ErrAlteredContent) {
		t.Errorf("Verify() = %v, want ErrAlteredContent", err)
	}
}

func TestVerifyIgnoresLaterValidationRules(t *testing.T) {
	d := testDefinition()
	d.Steps[0].Name = ""
	if err := testVersion(t, 1, d).Verify(); err != nil {
		t.Errorf("Verify() = %v, want nil for a version that only breaks a validation rule", err)
	}
}

func TestContentRejectsUnknownKeys(t *testing.T) {
	known := []byte(`{"format":1,"definition":{"name":"Trámite","steps":[]}}`)
	if _, err := (Version{Encoded: known}).Content(); err != nil {
		t.Fatalf("Content(%s) = %v, want nil", known, err)
	}
	encoded := []byte(`{"format":1,"definition":{"name":"Trámite","steps":[]},"deadline_days":10}`)
	v := Version{ProcedureType: 1, Number: 1, Digest: Digest(encoded), Encoded: encoded}
	if err := v.Verify(); err != nil {
		t.Errorf("Verify() = %v, want nil", err)
	}
	if _, err := v.Content(); err == nil {
		t.Errorf("Content(%s) = nil error, want the unknown key rejected", encoded)
	}
}

func TestContentRejectsAnotherFormat(t *testing.T) {
	for _, encoded := range []string{
		`{"format":2,"definition":{"name":"Trámite"}}`,
		`{"definition":{"name":"Trámite"}}`,
	} {
		v := Version{Number: 1, Digest: Digest([]byte(encoded)), Encoded: []byte(encoded)}
		if _, err := v.Content(); !errors.Is(err, ErrUnknownFormat) {
			t.Errorf("Content(%s) = %v, want ErrUnknownFormat", encoded, err)
		}
	}
}

func TestAuthorship(t *testing.T) {
	if err := (Authorship{ProposedBy: "ana", ApprovedBy: "beto"}).Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil for two people", err)
	}
	tests := []struct {
		name string
		a    Authorship
		want string
	}{
		{name: "no proposer", a: Authorship{ApprovedBy: "beto"}, want: "a version needs who proposed it"},
		{name: "no approver", a: Authorship{ProposedBy: "ana"}, want: "a version needs who approved it"},
		{name: "same person", a: Authorship{ProposedBy: "ana", ApprovedBy: "ana"}, want: "the approver must be someone other than the proposer"},
		{name: "same person in capitals", a: Authorship{ProposedBy: "ana", ApprovedBy: "Ana"}, want: "the approver must be someone other than the proposer"},
		{name: "proposer with a space", a: Authorship{ProposedBy: "ana ", ApprovedBy: "ana"}, want: `proposer "ana " has surrounding spaces or invisible characters`},
		{name: "approver with an invisible character", a: Authorship{ProposedBy: "ana", ApprovedBy: "ana\u200b"}, want: `approver "ana\u200b" has surrounding spaces or invisible characters`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.a.Validate()
			if !errors.Is(err, ErrInvalidAuthorship) || err.Error() != "invalid authorship: "+tt.want {
				t.Errorf("Validate() = %v, want %q", err, tt.want)
			}
		})
	}
}
