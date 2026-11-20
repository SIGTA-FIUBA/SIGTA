package postgres_test

import (
	"testing"

	"backend/src/auth/adapters/postgres"
	"backend/src/auth/ports/output"
	"backend/src/auth/ports/output/outputtest"
	"backend/src/migrations/dbtest"
)

func TestAreaRepository(t *testing.T) {
	for _, isolation := range []dbtest.Isolation{dbtest.ReadCommitted, dbtest.Serializable} {
		t.Run(string(isolation), func(t *testing.T) {
			outputtest.TestAreaRepository(t, func(t *testing.T) output.AreaRepository {
				return postgres.NewAreaRepository(dbtest.NewSchema(t).Pool(t, dbtest.AppRole, isolation))
			})
		})
	}
}
