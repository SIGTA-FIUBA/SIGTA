package postgres_test

import (
	"testing"

	"backend/src/migrations/dbtest"
	"backend/src/procedure/adapters/postgres"
	"backend/src/procedure/ports/output/outputtest"
)

func storesAt(isolation dbtest.Isolation) func(t *testing.T) outputtest.ProcedureStores {
	return func(t *testing.T) outputtest.ProcedureStores {
		pool := dbtest.NewSchema(t).Pool(t, dbtest.AppRole, isolation)
		return outputtest.ProcedureStores{
			Versions:   postgres.NewProcedureTypeVersionRepository(pool),
			Procedures: postgres.NewProcedureRepository(pool),
			Documents:  postgres.NewDocumentRepository(pool),
		}
	}
}

func TestProcedureRepository(t *testing.T) {
	for _, isolation := range []dbtest.Isolation{dbtest.ReadCommitted, dbtest.Serializable} {
		t.Run(string(isolation), func(t *testing.T) {
			outputtest.TestProcedureRepository(t, storesAt(isolation))
		})
	}
}
