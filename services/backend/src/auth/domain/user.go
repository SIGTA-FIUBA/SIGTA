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

// SearchFrom are the letters that a search of people takes as the letter in
// the same place of SearchTo: capitals and accented Latin letters, so that
// "perez" finds "PÉREZ" and "joao" finds "João". Any other letter is compared
// as it is. The fold is a mapping of single characters, which a database
// applies with translate whatever its locale, so both adapters agree.
const (
	SearchFrom = "áàâäãåéèêëíìîïóòôöõúùûüñçýÿ" + "ÁÀÂÄÃÅÉÈÊËÍÌÎÏÓÒÔÖÕÚÙÛÜÑÇÝŸ" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	SearchTo   = "aaaaaaeeeeiiiiooooouuuuncyy" + "aaaaaaeeeeiiiiooooouuuuncyy" + "abcdefghijklmnopqrstuvwxyz"
)

var searchFolding = func() map[rune]rune {
	from, to := []rune(SearchFrom), []rune(SearchTo)
	folding := make(map[rune]rune, len(from))
	for i := range from {
		folding[from[i]] = to[i]
	}
	return folding
}()

// FoldForSearch is text as a search of people compares it: each letter of
// SearchFrom replaced by its letter in SearchTo.
func FoldForSearch(text string) string {
	return strings.Map(func(r rune) rune {
		if folded, ok := searchFolding[r]; ok {
			return folded
		}
		return r
	}, text)
}

// InDomain tells whether a normalized email belongs to the given domain, such
// as "fi.uba.ar".
func InDomain(email, emailDomain string) bool {
	return strings.HasSuffix(email, "@"+emailDomain)
}
