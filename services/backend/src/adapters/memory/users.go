package memory

import (
	"context"
	"domain"
	"sync"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]domain.User
}

func NewUserRepository(users ...domain.User) *UserRepository {
	r := &UserRepository{users: make(map[string]domain.User)}
	for _, u := range users {
		r.Add(u)
	}
	return r
}

func (r *UserRepository) Add(user domain.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user.Email = domain.NormalizeEmail(user.Email)
	r.users[user.ID] = user
}

func (r *UserRepository) FindByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrUnknownUser
}

func (r *UserRepository) FindByID(_ context.Context, id string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return domain.User{}, domain.ErrUnknownUser
	}
	return u, nil
}
