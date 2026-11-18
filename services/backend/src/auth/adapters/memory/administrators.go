package memory

import (
	"context"

	"backend/src/auth/ports/output"
)

var _ output.Administrators = (*Administrators)(nil)

type Administrators struct {
	ids map[string]bool
}

func NewAdministrators(ids ...string) *Administrators {
	a := &Administrators{ids: map[string]bool{}}
	for _, id := range ids {
		a.ids[id] = true
	}
	return a
}

func (a *Administrators) IsAdministrator(_ context.Context, userID string) (bool, error) {
	return a.ids[userID], nil
}
