package application

import (
	"context"
	"errors"
	"fmt"

	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
	"backend/src/auth/ports/output"
)

var (
	_ input.People    = (*PeopleService)(nil)
	_ input.Directory = (*PeopleService)(nil)
)

const actionFindPerson = "find person"

type PeopleService struct {
	administratorGuard
	users output.UserRepository
}

func NewPeopleService(users output.UserRepository, administrators output.AdministratorRepository) *PeopleService {
	return &PeopleService{administratorGuard: administratorGuard{administrators: administrators}, users: users}
}

func (s *PeopleService) Names(ctx context.Context, ids []string) (map[string]string, error) {
	users, err := s.Users(ctx, ids)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(users))
	for id, user := range users {
		names[id] = user.Name
	}
	return names, nil
}

func (s *PeopleService) Users(ctx context.Context, ids []string) (map[string]domain.User, error) {
	users := make(map[string]domain.User, len(ids))
	for _, id := range ids {
		if _, done := users[id]; done {
			continue
		}
		user, err := s.users.Get(ctx, id)
		if errors.Is(err, domain.ErrUserNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("user %s: %w", id, err)
		}
		users[id] = user
	}
	return users, nil
}

// Find matches the whole email, without telling uppercase from lowercase, so
// that an administrator finds only someone whose email they know.
func (s *PeopleService) Find(ctx context.Context, administratorID, email string) (domain.User, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionFindPerson); err != nil {
		return domain.User{}, err
	}
	user, err := s.users.GetByEmail(ctx, domain.NormalizeEmail(email))
	if err != nil {
		return domain.User{}, fmt.Errorf("person with email %s: %w", email, err)
	}
	return user, nil
}
