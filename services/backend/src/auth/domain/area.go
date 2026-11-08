package domain

import (
	"errors"
	"regexp"
	"time"
)

var (
	ErrInvalidArea  = errors.New("invalid area")
	ErrAreaNotFound = errors.New("area not found")
)

var areaIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Area exists only within staff: the applicant has none (ADR-020). A deleted
// area is kept, so that memberships and published versions still name it.
type Area struct {
	ID        string
	Name      string
	DeletedAt time.Time
	DeletedBy string
}

func (a Area) Deleted() bool {
	return !a.DeletedAt.IsZero()
}

func ValidAreaID(id string) bool {
	return areaIDPattern.MatchString(id)
}
