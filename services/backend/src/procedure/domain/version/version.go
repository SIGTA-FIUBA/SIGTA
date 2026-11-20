package version

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"backend/src/procedure/domain/definition"
)

var (
	ErrAlteredContent = errors.New("content does not match its digest")
	ErrUnknownFormat  = errors.New("unknown content format")
)

// A ProcedureTypeVersion keeps its content encoded, exactly as it was hashed
// when it was published, and is never changed afterwards.
type ProcedureTypeVersion struct {
	ProcedureTypeID int64
	Number          int
	Digest          string
	EncodedContent  []byte
	Authorship      Authorship
	PublishedAt     time.Time
}

// VerifyDigest checks only the digest and not the validation rules, so that a
// rule added later does not invalidate versions that procedures in progress
// follow.
func (v ProcedureTypeVersion) VerifyDigest() error {
	if got := Digest(v.EncodedContent); got != v.Digest {
		return fmt.Errorf("%w: digest is %s, published as %s", ErrAlteredContent, got, v.Digest)
	}
	return nil
}

// DecodeContent rejects keys it does not know: they mean the version was
// published by a newer model, or that a key was renamed, and ignoring them
// would change what the version says without failing its digest. Formats
// before 3 are converted to the current shape.
func (v ProcedureTypeVersion) DecodeContent() (Content, error) {
	var head struct {
		Format int `json:"format"`
	}
	if err := json.Unmarshal(v.EncodedContent, &head); err != nil {
		return Content{}, fmt.Errorf("decode content: %w", err)
	}
	var c Content
	switch {
	case head.Format < 1 || head.Format > ContentFormat:
		return Content{}, fmt.Errorf("%w %d, want 1 to %d", ErrUnknownFormat, head.Format, ContentFormat)
	case head.Format < 3:
		var legacy legacyContent
		if err := decodeStrict(v.EncodedContent, &legacy); err != nil {
			return Content{}, err
		}
		c = legacy.convert()
	default:
		if err := decodeStrict(v.EncodedContent, &c); err != nil {
			return Content{}, err
		}
		// Read as the current format, so that publishing the same
		// definition again does not add a version that changes nothing.
		c.Format = ContentFormat
	}
	// Compared with the format that added each key, not with the current one,
	// so that a later format does not reject what a version could say.
	if head.Format < sendBackFormat && hasSendBack(c.Definition) {
		return Content{}, fmt.Errorf("decode content: format %d has no send backs, format %d added them", head.Format, sendBackFormat)
	}
	if head.Format < resolutionFormat && slices.ContainsFunc(c.Definition.Documents, func(d definition.Document) bool { return d.Resolution }) {
		return Content{}, fmt.Errorf("decode content: format %d has no resolution, format %d added it", head.Format, resolutionFormat)
	}
	return c, nil
}

func hasSendBack(d definition.ProcedureTypeDefinition) bool {
	return slices.ContainsFunc(d.Steps, func(s definition.Step) bool {
		return slices.ContainsFunc(s.Outcomes, func(o definition.Outcome) bool { return o.SendBack })
	})
}

func decodeStrict(encoded []byte, into any) error {
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil {
		return fmt.Errorf("decode content: %w", err)
	}
	return nil
}
