package application

import (
	"context"
	"fmt"

	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
	"backend/src/auth/ports/output"
)

var _ input.Areas = (*AreaService)(nil)

type AreaService struct {
	areas output.AreaRepository
}

func NewAreaService(areas output.AreaRepository) *AreaService {
	return &AreaService{areas: areas}
}

func (s *AreaService) List(ctx context.Context) ([]domain.Area, error) {
	areas, err := s.areas.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list areas: %w", err)
	}
	return areas, nil
}
