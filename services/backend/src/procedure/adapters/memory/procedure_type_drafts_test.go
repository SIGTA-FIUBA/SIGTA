package memory

import (
	"testing"

	"backend/src/procedure/ports/output"
	"backend/src/procedure/ports/output/outputtest"
)

func TestProcedureTypeDraftRepository(t *testing.T) {
	outputtest.TestProcedureTypeDraftRepository(t, func(*testing.T) output.ProcedureTypeDraftRepository { return NewProcedureTypeDraftRepository() })
}
