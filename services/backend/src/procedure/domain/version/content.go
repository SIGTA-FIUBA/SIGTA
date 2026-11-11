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
// Format 3 gives each step a role, leaves areas to staff, and has staff steps
// ask for permissions of their area instead of a fixed profile (ADR-020):
// formats 1 and 2 are converted when read (see legacy.go).
const ContentFormat = 3

// Content carries the staff areas its steps use, with the permissions they ask
// for, so that a version reads the same after either is renamed.
type Content struct {
	Format     int                                `json:"format"`
	Definition definition.ProcedureTypeDefinition `json:"definition"`
	Areas      []definition.Area                  `json:"areas,omitempty"`
}

func NewContent(d definition.ProcedureTypeDefinition, areas []definition.Area) Content {
	asked := map[string]map[string]bool{}
	for _, s := range d.Steps {
		if s.Area == "" {
			continue
		}
		if asked[s.Area] == nil {
			asked[s.Area] = map[string]bool{}
		}
		for _, p := range s.Permissions {
			asked[s.Area][p] = true
		}
	}
	var referenced []definition.Area
	for _, a := range areas {
		permissions, used := asked[a.ID]
		if !used {
			continue
		}
		copied := definition.Area{ID: a.ID, Name: a.Name}
		for _, p := range a.Permissions {
			if permissions[p.ID] {
				copied.Permissions = append(copied.Permissions, p)
			}
		}
		// Sorted so that reordering the areas or their permissions does not
		// change the digest.
		slices.SortFunc(copied.Permissions, func(a, b definition.Permission) int { return cmp.Compare(a.ID, b.ID) })
		referenced = append(referenced, copied)
	}
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
