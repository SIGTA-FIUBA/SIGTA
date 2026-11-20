package domain

import (
	"errors"
	"time"
)

var (
	ErrAdministratorNotFound = errors.New("administrator not found")
	ErrAlreadyAdministrator  = errors.New("the person is an administrator already")
	// ErrLastAdministrator means ending the period would leave SIGTA without
	// administrators, and then nobody could grant one again but the console.
	ErrLastAdministrator = errors.New("SIGTA would be left without administrators")
)

// Administrator is a period in which a person administers areas and
// memberships. A zero Until means it holds until it is ended. GrantedBy is
// empty only in the initial grant of the system, which the grant-admin
// command adds.
type Administrator struct {
	ID        string
	UserID    string
	From      time.Time
	Until     time.Time
	GrantedBy string
	EndedBy   string
}

func (a Administrator) ActiveAt(t time.Time) bool {
	return !a.From.After(t) && (a.Until.IsZero() || a.Until.After(t))
}
