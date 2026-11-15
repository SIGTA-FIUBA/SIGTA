package application

import (
	"context"
	"errors"
	"fmt"

	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
	out "backend/src/auth/ports/output"
)

var _ input.People = (*PeopleService)(nil)

type PeopleService struct {
	users out.UserRepository
}

func NewPeopleService(users out.UserRepository) *PeopleService {
	return &PeopleService{users: users}
}

func (s *PeopleService) Names(ctx context.Context, ids []string) (map[string]string, error) {
	names := make(map[string]string, len(ids))
	for _, id := range ids {
		if _, done := names[id]; done {
			continue
		}
		user, err := s.users.FindByID(ctx, id)
		if errors.Is(err, domain.ErrUnknownUser) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("user %s: %w", id, err)
		}
		names[id] = user.Name
	}
	return names, nil
}
