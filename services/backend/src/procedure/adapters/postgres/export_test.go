package postgres

import "context"

const Attempts = attempts

func SetBeforeCommit(s *ProcedureVersions, f func(ctx context.Context)) { s.beforeCommit = f }
