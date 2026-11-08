package auth

import (
	"context"
	"fmt"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var _ output.AreaCatalog = (*AreaCatalog)(nil)

// AreaCatalog lists the areas of auth as office areas, after the applicant
// area, which is not part of the organization but holds the steps of whoever
// started the procedure.
type AreaCatalog struct {
	areas     authinput.Areas
	applicant domain.Area
}

func NewAreaCatalog(areas authinput.Areas, applicant domain.Area) *AreaCatalog {
	return &AreaCatalog{areas: areas, applicant: applicant}
}

func (c *AreaCatalog) List(ctx context.Context) ([]domain.Area, error) {
	areas, err := c.areas.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("areas: %w", err)
	}
	listed := make([]domain.Area, 0, len(areas)+1)
	listed = append(listed, c.applicant)
	for _, a := range areas {
		listed = append(listed, domain.Area{ID: a.ID, Name: a.Name, Kind: domain.AreaOffice})
	}
	return listed, nil
}
