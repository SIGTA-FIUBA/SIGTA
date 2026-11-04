package input

import (
	"context"
	"errors"
	"fmt"

	"backend/src/procedure/domain/catalog"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

// ErrOutdatedBase means the definition was edited from a version that is no
// longer the latest one, so publishing it would undo the changes in between.
var ErrOutdatedBase = errors.New("the definition was edited from an outdated version")

// OutdatedBaseError is the ErrOutdatedBase that Publish returns. Latest is the
// version to edit from again; after losing a race to another publication it
// may already be outdated too.
type OutdatedBaseError struct {
	Base   int
	Latest int
}

func (e *OutdatedBaseError) Error() string {
	return fmt.Sprintf("%v: edited from version %d, the latest is %d", ErrOutdatedBase, e.Base, e.Latest)
}

func (e *OutdatedBaseError) Unwrap() error { return ErrOutdatedBase }

type Publication struct {
	ProcedureType int64
	// Base is the number of the version the definition was edited from, or 0
	// for the first version.
	Base       int
	Definition definition.ProcedureTypeDefinition
	Authorship version.Authorship
}

// CatalogEntry is what an applicant sees of a procedure type before starting it.
type CatalogEntry struct {
	ProcedureType int64
	Version       int
	Name          string
	Description   string
	Deadlines     []string
	Admissibility []string
	Fields        []catalog.FieldRequirement
	Documents     []catalog.DocumentRequirement
}

type ProcedureTypes interface {
	// Publish returns the version as it is stored. If the content is the same
	// as the latest version, whatever the base, it publishes nothing and
	// returns that version with its own authorship.
	//
	// It fails with version.ErrInvalidAuthorship, with a
	// *definition.ValidationError (definition.ErrInvalidAreas or
	// definition.ErrInvalidDefinition) whose problems ValidationProblemsOf
	// reads, with an *OutdatedBaseError if the base is not the latest version,
	// and with version.ErrAlteredContent, output.ErrWrongVersion or
	// version.ErrUnknownFormat if the latest version cannot be trusted or read.
	Publish(ctx context.Context, p Publication) (version.ProcedureTypeVersion, error)
	// Version returns a published version after checking its digest, for the
	// procedures that follow it. It fails with output.ErrNoVersion if the
	// version does not exist, and with version.ErrAlteredContent or
	// output.ErrWrongVersion if it cannot be trusted.
	Version(ctx context.Context, procedureType int64, number int) (version.ProcedureTypeVersion, error)
	// Catalog returns the latest version of every procedure type, after checking
	// their digests. It fails with version.ErrAlteredContent,
	// output.ErrWrongVersion or version.ErrUnknownFormat if any of them cannot be
	// trusted or read.
	Catalog(ctx context.Context) ([]CatalogEntry, error)
}
