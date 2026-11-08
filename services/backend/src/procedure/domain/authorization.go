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

// PassesThrough tells whether a step of the content is in the area. Staff see
// the procedures that pass through their areas (ADR-017).
func (c Content) PassesThrough(area string) bool {
	return slices.ContainsFunc(c.Definition.Steps, func(s Step) bool { return s.Area == area })
}

// AreaOf returns the copy of the area of a step that the version keeps.
func (c Content) AreaOf(s Step) (Area, bool) {
	i := slices.IndexFunc(c.Areas, func(a Area) bool { return a.ID == s.Area })
	if i < 0 {
		return Area{}, false
	}
	return c.Areas[i], true
}

type DeniedAttempt struct {
	UserID    string
	Action    string
	Procedure string
	At        time.Time
}
