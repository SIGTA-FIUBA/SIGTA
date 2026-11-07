package memory

import (
	"context"
	"slices"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/ports/output"
)

var _ output.AreaCatalog = (*AreaCatalog)(nil)

type AreaCatalog struct {
	areas []definition.Area
}

func NewAreaCatalog(areas ...definition.Area) *AreaCatalog {
	return &AreaCatalog{areas: slices.Clone(areas)}
}

func (c *AreaCatalog) List(context.Context) ([]definition.Area, error) {
	return slices.Clone(c.areas), nil
}
