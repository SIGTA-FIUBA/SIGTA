package output

import (
	"context"

	"backend/src/procedure/domain"
)

type DeniedAttemptLog interface {
	Record(ctx context.Context, a domain.DeniedAttempt) error
}
