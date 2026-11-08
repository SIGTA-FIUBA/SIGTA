package memory

import (
	"context"
	"slices"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var _ output.AreaRepository = (*AreaRepository)(nil)

type AreaRepository struct {
	areas []domain.Area
}

func NewAreaRepository(areas ...domain.Area) *AreaRepository {
	return &AreaRepository{areas: slices.Clone(areas)}
}

func (r *AreaRepository) List(context.Context) ([]domain.Area, error) {
	return slices.Clone(r.areas), nil
}
