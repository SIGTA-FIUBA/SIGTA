package output

import (
	"context"
	"errors"

	"backend/src/procedure/domain"
)

var (
	ErrNoVersion     = errors.New("procedure version not found")
	ErrVersionExists = errors.New("procedure version already exists")
	// ErrWrongVersion means the store returned another version than the one
	// asked for. The digest covers the content, not where it is stored, so
	// only the caller can notice it.
	ErrWrongVersion = errors.New("the store returned another version")
)

// ProcedureVersions stores published versions. A stored version is never
// changed or deleted, and its encoded content is kept byte for byte, since
// the digest is computed over those bytes.
type ProcedureVersions interface {
	// Latest returns the version with the highest number, or ErrNoVersion if
	// the procedure type has none.
	Latest(ctx context.Context, procedureType int64) (domain.Version, error)
	// LatestOfEach returns the latest version of every procedure type that has
	// one, ordered by procedure type.
	LatestOfEach(ctx context.Context) ([]domain.Version, error)
	// Get fails with ErrNoVersion if the version does not exist.
	Get(ctx context.Context, procedureType int64, number int) (domain.Version, error)
	// Add fails with ErrVersionExists if the procedure type already has a
	// version with that number, also when another publication added it at the
	// same time.
	Add(ctx context.Context, v domain.Version) error
}
