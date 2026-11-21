package memory

import (
	"testing"

	"backend/src/auth/ports/output"
	"backend/src/auth/ports/output/outputtest"
)

func TestAdministratorRepository(t *testing.T) {
	outputtest.TestAdministratorRepository(t, func(*testing.T) output.FirstAdministratorRepository { return NewAdministratorRepository() })
	outputtest.TestAdministratorPeriodRepository(t, func(*testing.T) output.AdministratorPeriodRepository { return NewAdministratorRepository() })
}
