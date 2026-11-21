package postgres_test

import (
	"testing"

	"backend/src/migrations/dbtest"
	"backend/src/procedure/adapters/postgres"
	"backend/src/procedure/ports/output"
	"backend/src/procedure/ports/output/outputtest"
)

func TestProcedureTypeDraftRepository(t *testing.T) {
	for _, isolation := range []dbtest.Isolation{dbtest.ReadCommitted, dbtest.Serializable} {
		t.Run(string(isolation), func(t *testing.T) {
			outputtest.TestProcedureTypeDraftRepository(t, func(t *testing.T) output.ProcedureTypeDraftRepository {
				return postgres.NewProcedureTypeDraftRepository(dbtest.NewSchema(t).Pool(t, dbtest.AppRole, isolation))
			})
		})
	}
}
