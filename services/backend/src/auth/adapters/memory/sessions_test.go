package memory

import (
	"testing"

	"backend/src/auth/ports/output"
	"backend/src/auth/ports/output/outputtest"
)

func TestSessionRepository(t *testing.T) {
	outputtest.TestSessionRepository(t, func(*testing.T) output.SessionRepository { return NewSessionRepository() })
}
