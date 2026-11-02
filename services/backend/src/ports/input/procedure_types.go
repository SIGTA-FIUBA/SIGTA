package input

import (
	"context"
	"errors"

	"backend/src/domain/procedure"
)

// ErrOutdatedBase means the definition was edited from a version that is no
// longer the latest one, so publishing it would undo the changes in between.
var ErrOutdatedBase = errors.New("the definition was edited from an outdated version")

type Publication struct {
	ProcedureType int64
	// Base is the number of the version the definition was edited from, or 0
	// for the first version.
	Base       int
	Definition procedure.Definition
	Authorship procedure.Authorship
}

type ProcedureTypes interface {
	// Publish returns the version as it is stored. If the content is the same
	// as the latest version, whatever the base, it publishes nothing and
	// returns that version with its own authorship.
	Publish(ctx context.Context, p Publication) (procedure.Version, error)
	// Version returns a published version after checking its digest, for the
	// procedures that follow it.
	Version(ctx context.Context, procedureType int64, number int) (procedure.Version, error)
}
