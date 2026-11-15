package domain

import (
	"errors"

	"backend/src/procedure/domain/version"
)

var (
	ErrAccessDenied      = errors.New("access denied")
	ErrProcedureNotFound = errors.New("procedure not found")
)

// A Procedure keeps the digest of its version as well as the number, so that
// a version lost and published again with the same number is noticed.
type Procedure struct {
	ID            string
	ProcedureType int64
	Version       int
	VersionDigest string
	Applicant     string
	OpenStep      string
}

func (p Procedure) Pin() version.Pin {
	return version.Pin{ProcedureTypeID: p.ProcedureType, Number: p.Version, Digest: p.VersionDigest}
}
