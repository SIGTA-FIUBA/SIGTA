package memory

import (
	"testing"

	"backend/src/auth/ports/output"
	"backend/src/auth/ports/output/outputtest"
)

func TestMembershipRepository(t *testing.T) {
	outputtest.TestMembershipRepository(t, func(*testing.T) output.MembershipRepository { return NewMembershipRepository() })
}
