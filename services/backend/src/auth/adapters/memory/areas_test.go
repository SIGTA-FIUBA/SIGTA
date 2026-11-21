package memory

import (
	"testing"

	"backend/src/auth/ports/output"
	"backend/src/auth/ports/output/outputtest"
)

func TestAreaRepository(t *testing.T) {
	outputtest.TestAreaRepository(t, func(*testing.T) output.AreaRepository { return NewAreaRepository() })
}
