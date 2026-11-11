package domain

import "errors"

var (
	ErrAccessDenied      = errors.New("access denied")
	ErrProcedureNotFound = errors.New("procedure not found")
)

type Procedure struct {
	ID            string
	ProcedureType int64
	Version       int
	Applicant     string
	OpenStep      string
}
