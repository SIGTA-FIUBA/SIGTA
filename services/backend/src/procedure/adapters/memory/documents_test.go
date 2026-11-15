package memory

import (
	"testing"

	"backend/src/procedure/ports/output"
	"backend/src/procedure/ports/output/outputtest"
)

func TestDocumentRepository(t *testing.T) {
	outputtest.TestDocumentRepository(t, func(*testing.T) output.DocumentRepository { return NewDocumentRepository() })
}
