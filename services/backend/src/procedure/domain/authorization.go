package domain

import (
	"errors"
	"slices"
	"time"
)

var (
	ErrAccessDenied      = errors.New("access denied")
	ErrProcedureNotFound = errors.New("procedure not found")
)

// PassesThrough tells whether a staff step of the content is in the area.
// Staff see the procedures that pass through their areas (ADR-017).
func (c Content) PassesThrough(area string) bool {
	return slices.ContainsFunc(c.Definition.Steps, func(s Step) bool { return s.Area == area })
}

type DeniedAttempt struct {
	UserID    string
	Action    string
	Procedure string
	At        time.Time
}
