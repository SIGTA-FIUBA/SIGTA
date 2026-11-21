package version_test

import (
	"errors"
	"testing"

	"backend/src/procedure/domain/definition/definitiontest"
	"backend/src/procedure/domain/version"
)

func TestPinCheck(t *testing.T) {
	pinned := testVersion(t, 1, definitiontest.Definition())
	if err := pinned.Pin().Check(pinned); err != nil {
		t.Fatalf("Check() of the pinned version = %v, want nil", err)
	}

	republished := definitiontest.Definition()
	republished.Name = "Otro trámite"
	sameNumber := testVersion(t, 1, republished)
	otherType := pinned
	otherType.ProcedureTypeID = 2
	for name, v := range map[string]version.ProcedureTypeVersion{
		"another digest":         sameNumber,
		"another procedure type": otherType,
		"another number":         testVersion(t, 2, definitiontest.Definition()),
	} {
		if err := pinned.Pin().Check(v); !errors.Is(err, version.ErrNotPinnedVersion) {
			t.Errorf("Check() of %s = %v, want ErrNotPinnedVersion", name, err)
		}
	}
}
