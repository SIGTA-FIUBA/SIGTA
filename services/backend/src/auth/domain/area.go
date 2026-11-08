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

// ApplicantAreaID is reserved: procedure types use it for the steps of whoever
// started the procedure, so no staff area can take it.
const ApplicantAreaID = "applicant"

var areaIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Area is a part of the organization that staff belong to, such as the
// Secretaría or a curricular commission. Areas are data: an administrator
// creates them, and procedure types refer to them by ID (ADR-017). A deleted
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
	return areaIDPattern.MatchString(id) && id != ApplicantAreaID
}
