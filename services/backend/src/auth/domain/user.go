package domain

import (
	"errors"
	"strings"
)

var ErrUserNotFound = errors.New("no user registered with that identity")

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

// SearchAccents are the accented letters that a search of people takes as
// the letter in the same place of SearchPlain, so that "perez" finds "Pérez".
const (
	SearchAccents = "áéíóúüñÁÉÍÓÚÜÑàèìòùÀÈÌÒÙ"
	SearchPlain   = "aeiouunAEIOUUNaeiouAEIOU"
)

var searchFolding = func() *strings.Replacer {
	accents, plain := []rune(SearchAccents), []rune(SearchPlain)
	pairs := make([]string, 0, 2*len(accents))
	for i := range accents {
		pairs = append(pairs, string(accents[i]), string(plain[i]))
	}
	return strings.NewReplacer(pairs...)
}()

// FoldForSearch is text as a search of people compares it: without the
// accents of SearchAccents and in lowercase.
func FoldForSearch(text string) string {
	return strings.ToLower(searchFolding.Replace(text))
}

// InDomain tells whether a normalized email belongs to the given domain, such
// as "fi.uba.ar".
func InDomain(email, emailDomain string) bool {
	return strings.HasSuffix(email, "@"+emailDomain)
}
