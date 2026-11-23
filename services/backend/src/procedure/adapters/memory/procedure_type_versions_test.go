package memory

import (
	"testing"

	"backend/src/procedure/ports/output"
	"backend/src/procedure/ports/output/outputtest"
)

func TestProcedureTypeVersionRepository(t *testing.T) {
	outputtest.TestProcedureTypeVersionRepository(t, func(*testing.T) output.ProcedureTypeVersionRepository { return NewProcedureTypeVersionRepository() })
}
