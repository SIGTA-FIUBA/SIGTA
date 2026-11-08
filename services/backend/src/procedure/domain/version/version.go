package version

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
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
	switch {
	case head.Format < 1 || head.Format > ContentFormat:
		return Content{}, fmt.Errorf("%w %d, want 1 to %d", ErrUnknownFormat, head.Format, ContentFormat)
	case head.Format < 3:
		var legacy legacyContent
		if err := decodeStrict(v.EncodedContent, &legacy); err != nil {
			return Content{}, err
		}
		return legacy.convert()
	}
	var c Content
	if err := decodeStrict(v.EncodedContent, &c); err != nil {
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
