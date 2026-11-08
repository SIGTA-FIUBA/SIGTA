package version

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"

	"backend/src/procedure/domain/definition"
)

// ContentFormat tells which shape of Content an encoded version has. It goes
// up whenever Content changes, so that code that does not know a format
// refuses to read it instead of reading it partly. Format 2 added what the
// catalog shows, all of it optional in the encoding, so format 1 still reads.
// Format 3 gives each step a role and leaves areas to staff (ADR-020): formats
// 1 and 2 are converted when read (see legacy.go).
const ContentFormat = 3

// Content carries the staff areas its steps use, so that a version reads the
// same after an area is renamed.
type Content struct {
	Format     int                                `json:"format"`
	Definition definition.ProcedureTypeDefinition `json:"definition"`
	Areas      []definition.Area                  `json:"areas,omitempty"`
}

func NewContent(d definition.ProcedureTypeDefinition, areas []definition.Area) Content {
	used := map[string]bool{}
	for _, s := range d.Steps {
		if s.Area != "" {
			used[s.Area] = true
		}
	}
	var referenced []definition.Area
	for _, a := range areas {
		if used[a.ID] {
			referenced = append(referenced, a)
		}
	}
	// Sorted so that reordering the areas does not change the digest.
	slices.SortFunc(referenced, func(a, b definition.Area) int { return cmp.Compare(a.ID, b.ID) })
	return Content{Format: ContentFormat, Definition: d, Areas: referenced}
}

func (c Content) PassesThrough(area string) bool {
	return slices.ContainsFunc(c.Definition.Steps, func(s definition.Step) bool { return s.Area == area })
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
