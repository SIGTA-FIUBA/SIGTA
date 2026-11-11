package memory

import (
	"testing"

	"backend/src/procedure/ports/output"
	"backend/src/procedure/ports/output/outputtest"
)

func TestProcedureVersions(t *testing.T) {
	outputtest.TestProcedureVersions(t, func(*testing.T) output.ProcedureVersions { return NewProcedureVersions() })
}
