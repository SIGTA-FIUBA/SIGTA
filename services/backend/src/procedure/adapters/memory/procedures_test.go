package memory

import (
	"testing"

	"backend/src/procedure/ports/output/outputtest"
)

func newStores(*testing.T) outputtest.ProcedureStores {
	procedures := NewProcedureRepository()
	return outputtest.ProcedureStores{
		Versions:   NewProcedureTypeVersionRepository(),
		Procedures: procedures,
		Documents:  NewDocumentRepository(procedures),
	}
}

func TestProcedureRepository(t *testing.T) {
	outputtest.TestProcedureRepository(t, newStores)
}
