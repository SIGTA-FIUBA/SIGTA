package domain

import "time"

// FieldValue is what was entered for a field of the procedure type, on the
// step that asked for it. Items holds a list of subjects; Text, any other
// value, as it was entered once checked against the type of the field.
type FieldValue struct {
	Procedure string
	Field     string
	Step      string
	Text      string
	Items     []string
	EnteredBy string
	EnteredAt time.Time
}
