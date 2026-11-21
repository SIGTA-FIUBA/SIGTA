package version

import (
	"errors"
	"fmt"
)

// ErrNotPinnedVersion means a version has the number a procedure follows but
// not its digest: the version was lost and its number published again.
var ErrNotPinnedVersion = errors.New("not the version the procedure follows")

// A Pin is what a procedure stores of the version it follows. The number
// alone would not notice a version replaced by another with the same number.
type Pin struct {
	ProcedureTypeID int64
	Number          int
	Digest          string
}

func (v ProcedureTypeVersion) Pin() Pin {
	return Pin{ProcedureTypeID: v.ProcedureTypeID, Number: v.Number, Digest: v.Digest}
}

// Check fails with ErrNotPinnedVersion if v is not the pinned version. It does
// not verify the digest of v against its content: VerifyDigest does that.
func (p Pin) Check(v ProcedureTypeVersion) error {
	if v.Pin() != p {
		return fmt.Errorf("%w: pinned version %d of procedure type %d with digest %s, got version %d of procedure type %d with digest %s",
			ErrNotPinnedVersion, p.Number, p.ProcedureTypeID, p.Digest, v.Number, v.ProcedureTypeID, v.Digest)
	}
	return nil
}
