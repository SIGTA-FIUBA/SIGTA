package memory

import (
	"context"
	"slices"

	"backend/src/domain/procedure"
	"backend/src/ports/output"
)

var _ output.AreaCatalog = (*AreaCatalog)(nil)

type AreaCatalog struct {
	areas []procedure.Area
}

func NewAreaCatalog(areas ...procedure.Area) *AreaCatalog {
	return &AreaCatalog{areas: slices.Clone(areas)}
}

func (c *AreaCatalog) List(context.Context) ([]procedure.Area, error) {
	return slices.Clone(c.areas), nil
}
