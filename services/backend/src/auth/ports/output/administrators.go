package output

import "context"

type AdministratorRepository interface {
	IsAdministrator(ctx context.Context, userID string) (bool, error)
}
