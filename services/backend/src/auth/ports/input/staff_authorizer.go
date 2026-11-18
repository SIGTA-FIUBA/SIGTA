package input

import "context"

// An empty Scope covers every procedure type.
type AreaReach struct {
	Area  string
	Scope []int64
}

// Other modules ask StaffAuthorizer instead of reading memberships (ADR-021).
type StaffAuthorizer interface {
	Holds(ctx context.Context, userID, area string, procedureType int64, permissions []string) (bool, error)
	Reach(ctx context.Context, userID string) ([]AreaReach, error)
}
