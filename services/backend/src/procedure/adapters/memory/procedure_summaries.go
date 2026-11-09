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

// InProgress filters by the access, as a query would.
func (s *ProcedureSummaries) InProgress(_ context.Context, access inbox.Access) ([]inbox.Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	visible := []inbox.Summary{}
	for _, summary := range s.summaries {
		if access.CanSee(summary) {
			visible = append(visible, summary)
		}
	}
	return visible, nil
}
