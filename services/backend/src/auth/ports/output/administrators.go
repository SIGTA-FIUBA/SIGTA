package output

import "context"

type Administrators interface {
	IsAdministrator(ctx context.Context, userID string) (bool, error)
}
