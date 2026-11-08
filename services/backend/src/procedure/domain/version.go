package domain

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
// Format 3 gives each step a role and leaves areas to staff (ADR-017): formats
// 1 and 2 are converted when read (see legacy.go).
const ContentFormat = 3

// Content carries the staff areas its steps use, so that a version reads the
// same after an area is renamed.
type Content struct {
	Format     int        `json:"format"`
	Definition Definition `json:"definition"`
	Areas      []Area     `json:"areas,omitempty"`
}

func NewContent(d Definition, areas []Area) Content {
	used := map[string]bool{}
	for _, s := range d.Steps {
		if s.Area != "" {
			used[s.Area] = true
		}
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
// what the version says without failing its digest. Formats before 3 are
// converted to the current shape.
func (v Version) Content() (Content, error) {
	var head struct {
		Format int `json:"format"`
	}
	if err := json.Unmarshal(v.Encoded, &head); err != nil {
		return Content{}, fmt.Errorf("decode content: %w", err)
	}
	switch {
	case head.Format < 1 || head.Format > ContentFormat:
		return Content{}, fmt.Errorf("%w %d, want 1 to %d", ErrUnknownFormat, head.Format, ContentFormat)
	case head.Format < 3:
		var legacy legacyContent
		if err := decodeStrict(v.Encoded, &legacy); err != nil {
			return Content{}, err
		}
		return legacy.convert()
	}
	var c Content
	if err := decodeStrict(v.Encoded, &c); err != nil {
		return Content{}, err
	}
	return c, nil
}

func decodeStrict(encoded []byte, into any) error {
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		return fmt.Errorf("decode content: %w", err)
	}
	return nil
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
