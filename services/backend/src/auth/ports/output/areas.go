package output

import (
	"context"

	"backend/src/auth/domain"
)

// AreaRepository keeps every area, deleted or not. List returns only those
// not deleted, and Get fails with domain.ErrAreaNotFound if there is no such
// area.
type AreaRepository interface {
	List(ctx context.Context) ([]domain.Area, error)
	Get(ctx context.Context, id string) (domain.Area, error)
	Save(ctx context.Context, a domain.Area) error
}
