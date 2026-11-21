package postgres_test

import (
	"testing"

	"backend/src/auth/adapters/postgres"
	"backend/src/auth/ports/output"
	"backend/src/auth/ports/output/outputtest"
	"backend/src/migrations/dbtest"
)

func TestSessionRepository(t *testing.T) {
	for _, isolation := range []dbtest.Isolation{dbtest.ReadCommitted, dbtest.Serializable} {
		t.Run(string(isolation), func(t *testing.T) {
			outputtest.TestSessionRepository(t, func(t *testing.T) output.SessionRepository {
				return postgres.NewSessionRepository(dbtest.NewSchema(t).Pool(t, dbtest.AppRole, isolation))
			})
		})
	}
}
