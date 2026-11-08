package memory

import (
	"context"
	"slices"
	"sync"

	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureSummaries = (*ProcedureSummaries)(nil)

// ProcedureSummaries keeps the summaries it is given. Until procedures can be
// started and moved, it holds the synthetic ones of local development and
// those of the tests.
type ProcedureSummaries struct {
	mu        sync.Mutex
	summaries []inbox.Summary
}

func NewProcedureSummaries(summaries ...inbox.Summary) *ProcedureSummaries {
	return &ProcedureSummaries{summaries: slices.Clone(summaries)}
}

func (s *ProcedureSummaries) Add(summary inbox.Summary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.summaries = append(s.summaries, summary)
}

func (s *ProcedureSummaries) InProgress(context.Context) ([]inbox.Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.summaries), nil
}
