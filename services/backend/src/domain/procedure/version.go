package procedure

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// ContentFormat tells which shape of Content an encoded version has. It goes
// up whenever Content changes, so that code that does not know a format
// refuses to read it instead of reading it partly. Format 2 added what the
// catalog shows, all of it optional in the encoding, so format 1 still reads.
const ContentFormat = 2

// Content carries the areas its steps use, so that a version reads the same
// after an area is renamed.
type Content struct {
	Format     int        `json:"format"`
	Definition Definition `json:"definition"`
	Areas      []Area     `json:"areas,omitempty"`
}

func NewContent(d Definition, areas []Area) Content {
	used := map[string]bool{}
	for _, s := range d.Steps {
		used[s.Area] = true
	}
	var referenced []Area
	for _, a := range areas {
		if used[a.ID] {
			referenced = append(referenced, a)
		}
	}
	// Sorted so that reordering the areas does not change the digest.
	slices.SortFunc(referenced, func(a, b Area) int { return cmp.Compare(a.ID, b.ID) })
	return Content{Format: ContentFormat, Definition: d, Areas: referenced}
}

func (c Content) Encode() ([]byte, error) {
	encoded, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("encode content: %w", err)
	}
	return encoded, nil
}

// Digest hashes the encoded bytes as they are stored, never a re-encoding, so
// that changing how Content is encoded does not change the digest of the
// versions already published.
func Digest(encoded []byte) string {
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

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
	case a.ProposedBy != strings.TrimSpace(a.ProposedBy) || strings.ContainsFunc(a.ProposedBy, invisible):
		err = fmt.Errorf("proposer %q has surrounding spaces or invisible characters", a.ProposedBy)
	case a.ApprovedBy != strings.TrimSpace(a.ApprovedBy) || strings.ContainsFunc(a.ApprovedBy, invisible):
		err = fmt.Errorf("approver %q has surrounding spaces or invisible characters", a.ApprovedBy)
	case strings.EqualFold(a.ProposedBy, a.ApprovedBy):
		err = errors.New("the approver must be someone other than the proposer")
	default:
		return nil
	}
	return fmt.Errorf("%w: %w", ErrInvalidAuthorship, err)
}

// A Version keeps its content encoded, exactly as it was hashed when it was
// published, and is never changed afterwards.
type Version struct {
	ProcedureType int64
	Number        int
	Digest        string
	Encoded       []byte
	Authorship    Authorship
	PublishedAt   time.Time
}

// Verify checks only the digest and not the validation rules, so that a rule
// added later does not invalidate versions that procedures in progress follow.
func (v Version) Verify() error {
	if got := Digest(v.Encoded); got != v.Digest {
		return fmt.Errorf("%w: digest is %s, published as %s", ErrAlteredContent, got, v.Digest)
	}
	return nil
}

// Content rejects keys it does not know: they mean the version was published
// by a newer model, or that a key was renamed, and ignoring them would change
// what the version says without failing its digest.
func (v Version) Content() (Content, error) {
	var c Content
	decoder := json.NewDecoder(bytes.NewReader(v.Encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&c); err != nil {
		return Content{}, fmt.Errorf("decode content: %w", err)
	}
	if c.Format < 1 || c.Format > ContentFormat {
		return Content{}, fmt.Errorf("content format %d is not between 1 and %d", c.Format, ContentFormat)
	}
	return c, nil
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
