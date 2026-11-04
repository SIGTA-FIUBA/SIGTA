package version_test

import (
	"errors"
	"strings"
	"testing"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/definition/definitiontest"
	"backend/src/procedure/domain/version"
)

func testVersion(t *testing.T, number int, d definition.ProcedureTypeDefinition) version.ProcedureTypeVersion {
	t.Helper()
	encoded := encode(t, d, definitiontest.Areas())
	return version.ProcedureTypeVersion{ProcedureTypeID: 1, Number: number, Digest: version.Digest(encoded), EncodedContent: encoded}
}

func TestVerifyDigestDetectsAlteredContent(t *testing.T) {
	v := testVersion(t, 1, definitiontest.Definition())
	if err := v.VerifyDigest(); err != nil {
		t.Fatalf("VerifyDigest() = %v, want nil", err)
	}
	v.EncodedContent = []byte(strings.Replace(string(v.EncodedContent), "Solicitud", "Pedido", 1))
	if err := v.VerifyDigest(); !errors.Is(err, version.ErrAlteredContent) {
		t.Errorf("VerifyDigest() = %v, want ErrAlteredContent", err)
	}
}

func TestVerifyDigestIgnoresLaterValidationRules(t *testing.T) {
	d := definitiontest.Definition()
	d.Steps[0].Name = ""
	if err := testVersion(t, 1, d).VerifyDigest(); err != nil {
		t.Errorf("VerifyDigest() = %v, want nil for a version that only breaks a validation rule", err)
	}
}

func TestDecodeContentRejectsUnknownKeys(t *testing.T) {
	known := []byte(`{"format":1,"definition":{"name":"Trámite","steps":[]}}`)
	if _, err := (version.ProcedureTypeVersion{EncodedContent: known}).DecodeContent(); err != nil {
		t.Fatalf("DecodeContent(%s) = %v, want nil", known, err)
	}
	encoded := []byte(`{"format":1,"definition":{"name":"Trámite","steps":[]},"deadline_days":10}`)
	v := version.ProcedureTypeVersion{ProcedureTypeID: 1, Number: 1, Digest: version.Digest(encoded), EncodedContent: encoded}
	if err := v.VerifyDigest(); err != nil {
		t.Errorf("VerifyDigest() = %v, want nil", err)
	}
	if _, err := v.DecodeContent(); err == nil {
		t.Errorf("DecodeContent(%s) = nil error, want the unknown key rejected", encoded)
	}
}

func TestDecodeContentRejectsAnotherFormat(t *testing.T) {
	for _, encoded := range []string{
		`{"format":3,"definition":{"name":"Trámite"}}`,
		`{"definition":{"name":"Trámite"}}`,
	} {
		v := version.ProcedureTypeVersion{Number: 1, Digest: version.Digest([]byte(encoded)), EncodedContent: []byte(encoded)}
		if _, err := v.DecodeContent(); !errors.Is(err, version.ErrUnknownFormat) {
			t.Errorf("DecodeContent(%s) = %v, want ErrUnknownFormat", encoded, err)
		}
	}
}
