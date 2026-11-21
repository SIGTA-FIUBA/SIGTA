package domain

import "time"

// Administrator is a period in which a person administers areas and
// memberships.
type Administrator struct {
	ID     string
	UserID string
	From   time.Time
}
