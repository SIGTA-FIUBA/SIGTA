package postgres

import "context"

const Attempts = attempts

func SetBeforeCommit(s *ProcedureTypeVersionRepository, f func(ctx context.Context)) {
	s.beforeCommit = f
}
