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
				return postgres.NewAreaRepository(withoutAreas(t).Pool(t, dbtest.AppRole, isolation))
			})
		})
	}
}

// withoutAreas returns a migrated schema without the Secretariat, which the
// migrations add, so that the repository starts empty as the contract asks.
// Only the superuser can delete it.
func withoutAreas(t *testing.T) *dbtest.Schema {
	t.Helper()
	schema := dbtest.NewSchema(t)
	if _, err := schema.Admin(t).Exec(t.Context(), "DELETE FROM area_permissions; DELETE FROM areas"); err != nil {
		t.Fatal(err)
	}
	return schema
}
