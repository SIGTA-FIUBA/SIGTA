package output

import (
	"context"
	"errors"

	"backend/src/domain/procedure"
)

var (
	ErrNoVersion     = errors.New("procedure version not found")
	ErrVersionExists = errors.New("procedure version already exists")
)

// ProcedureVersions stores published versions. A stored version is never
// changed or deleted, and its encoded content is kept byte for byte, since
// the digest is computed over those bytes.
type ProcedureVersions interface {
	// Latest returns the version with the highest number, or ErrNoVersion if
	// the procedure type has none.
	Latest(ctx context.Context, procedureType int64) (procedure.Version, error)
	// Get fails with ErrNoVersion if the version does not exist.
	Get(ctx context.Context, procedureType int64, number int) (procedure.Version, error)
	// Add fails with ErrVersionExists if the procedure type already has a
	// version with that number, also when another publication added it at the
	// same time.
	Add(ctx context.Context, v procedure.Version) error
}

type AreaCatalog interface {
	List(ctx context.Context) ([]procedure.Area, error)
}
