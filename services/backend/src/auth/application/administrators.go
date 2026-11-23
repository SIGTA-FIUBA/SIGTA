package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
	"backend/src/auth/ports/output"
)

var _ input.Administrators = (*AdministratorService)(nil)

const (
	actionListAdministrators = "list administrators"
	actionGrantAdministrator = "grant administrator"
	actionEndAdministrator   = "end administrator"
)

var (
	// ErrUserNotSignedIn means the email has no user yet. SIGTA keeps no
	// password, so a person becomes a user by signing in.
	ErrUserNotSignedIn = errors.New("nobody signed in with that email")
	// ErrAdministratorExists means the first administrator was granted
	// already: the others are granted in SIGTA, which records who did it.
	ErrAdministratorExists = errors.New("there is an administrator already: this command only grants the first one")
)

// GrantFirstAdministrator makes the user with email the first administrator,
// as the initial grant of the system. It takes no author: nothing the one who
// runs it types would prove who did. The server never calls it.
func GrantFirstAdministrator(ctx context.Context, users output.UserRepository, administrators output.FirstAdministratorRepository, email string, newID func() string, at time.Time) (domain.User, error) {
	email = domain.NormalizeEmail(email)
	user, err := users.GetByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, fmt.Errorf("%w: %s has to sign in to SIGTA once first", ErrUserNotSignedIn, email)
	}
	if err != nil {
		return domain.User{}, err
	}
	granted, err := administrators.GrantFirst(ctx, domain.Administrator{ID: newID(), UserID: user.ID, From: at})
	if err != nil {
		return domain.User{}, err
	}
	if !granted {
		return domain.User{}, ErrAdministratorExists
	}
	return user, nil
}

// AdministratorService records who granted and who ended each administrator:
// the user of the session.
type AdministratorService struct {
	administratorGuard
	periods output.AdministratorPeriodRepository
	users   output.UserRepository
	newID   func() string
	now     func() time.Time
}

func NewAdministratorService(administrators output.AdministratorPeriodRepository, users output.UserRepository, newID func() string, now func() time.Time) *AdministratorService {
	return &AdministratorService{
		administratorGuard: administratorGuard{administrators: administrators},
		periods:            administrators,
		users:              users,
		newID:              newID,
		now:                now,
	}
}

func (s *AdministratorService) List(ctx context.Context, administratorID string) ([]domain.Administrator, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionListAdministrators); err != nil {
		return nil, err
	}
	administrators, err := s.periods.ListActive(ctx, s.now().UTC())
	if err != nil {
		return nil, fmt.Errorf("list administrators: %w", err)
	}
	return administrators, nil
}

// Grant fails with domain.ErrUserNotFound if the person never signed in.
func (s *AdministratorService) Grant(ctx context.Context, administratorID, userID string) (domain.Administrator, error) {
	if err := s.requireAdministrator(ctx, administratorID, actionGrantAdministrator); err != nil {
		return domain.Administrator{}, err
	}
	if _, err := s.users.Get(ctx, userID); err != nil {
		return domain.Administrator{}, fmt.Errorf("grant administrator to %s: %w", userID, err)
	}
	a := domain.Administrator{ID: s.newID(), UserID: userID, From: s.now().UTC(), GrantedBy: administratorID}
	if err := s.periods.Grant(ctx, a); err != nil {
		return domain.Administrator{}, err
	}
	return a, nil
}

// End keeps the period, with who ended it. An administrator may end their
// own, as long as another remains.
func (s *AdministratorService) End(ctx context.Context, administratorID, id string) error {
	if err := s.requireAdministrator(ctx, administratorID, actionEndAdministrator); err != nil {
		return err
	}
	return s.periods.End(ctx, id, s.now().UTC(), administratorID)
}
