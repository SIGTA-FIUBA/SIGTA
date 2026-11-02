package memory

import (
	"context"
	"slices"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var _ output.AreaCatalog = (*AreaCatalog)(nil)

type AreaCatalog struct {
	areas []domain.Area
}

func NewAreaCatalog(areas ...domain.Area) *AreaCatalog {
	return &AreaCatalog{areas: slices.Clone(areas)}
}

func (c *AreaCatalog) List(context.Context) ([]domain.Area, error) {
	return slices.Clone(c.areas), nil
}
