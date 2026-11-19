package memory

import (
	"context"

	"backend/src/auth/ports/output"
)

var _ output.AdministratorRepository = (*AdministratorRepository)(nil)

type AdministratorRepository struct {
	ids map[string]bool
}

func NewAdministratorRepository(ids ...string) *AdministratorRepository {
	a := &AdministratorRepository{ids: map[string]bool{}}
	for _, id := range ids {
		a.ids[id] = true
	}
	return a
}

func (a *AdministratorRepository) IsAdministrator(_ context.Context, userID string) (bool, error) {
	return a.ids[userID], nil
}
