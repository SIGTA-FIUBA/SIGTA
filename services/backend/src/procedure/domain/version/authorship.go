package version

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var ErrInvalidAuthorship = errors.New("invalid authorship")

// Authorship is who published a version: one person proposes it and another
// approves it.
type Authorship struct {
	ProposedBy string
	ApprovedBy string
}

func (a Authorship) Validate() error {
	var err error
	switch {
	case a.ProposedBy == "":
		err = errors.New("a version needs who proposed it")
	case a.ApprovedBy == "":
		err = errors.New("a version needs who approved it")
	case !plain(a.ProposedBy):
		err = fmt.Errorf("proposer %q has surrounding spaces or invisible characters", a.ProposedBy)
	case !plain(a.ApprovedBy):
		err = fmt.Errorf("approver %q has surrounding spaces or invisible characters", a.ApprovedBy)
	case strings.EqualFold(a.ProposedBy, a.ApprovedBy):
		err = errors.New("the approver must be someone other than the proposer")
	default:
		return nil
	}
	return fmt.Errorf("%w: %w", ErrInvalidAuthorship, err)
}

func plain(name string) bool {
	return name == strings.TrimSpace(name) && !strings.ContainsFunc(name, invisible)
}

// invisible rejects what cannot be seen on screen, so that two people who look
// the same are not told apart by an invisible character. It also catches
// invalid UTF-8, which decodes to the replacement character.
func invisible(r rune) bool {
	return r == utf8.RuneError ||
		unicode.IsControl(r) ||
		unicode.In(r, unicode.Cf, unicode.Co, unicode.Cs, unicode.Noncharacter_Code_Point) ||
		unicode.IsSpace(r) && r != ' '
}
