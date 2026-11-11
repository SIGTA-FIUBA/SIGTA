package domain

import (
	"errors"
	"strings"
)

var ErrUnknownUser = errors.New("no user registered with that identity")

// User is a person registered in SIGTA. What they may do comes from their
// memberships, never from the identity provider.
type User struct {
	ID    string
	Email string
	Name  string
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// InDomain tells whether a normalized email belongs to the given domain, such
// as "fi.uba.ar".
func InDomain(email, emailDomain string) bool {
	return strings.HasSuffix(email, "@"+emailDomain)
}
