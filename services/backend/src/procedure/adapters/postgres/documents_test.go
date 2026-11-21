package postgres_test

import (
	"testing"

	"backend/src/migrations/dbtest"
	"backend/src/procedure/ports/output/outputtest"
)

func TestDocumentRepository(t *testing.T) {
	for _, isolation := range []dbtest.Isolation{dbtest.ReadCommitted, dbtest.Serializable} {
		t.Run(string(isolation), func(t *testing.T) {
			outputtest.TestDocumentRepository(t, storesAt(isolation))
		})
	}
}
