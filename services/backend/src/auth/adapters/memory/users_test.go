package memory

import (
	"testing"

	"backend/src/auth/ports/output"
	"backend/src/auth/ports/output/outputtest"
)

func TestUserRepository(t *testing.T) {
	outputtest.TestUserRepository(t, func(*testing.T) output.UserRepository { return NewUserRepository() })
}
