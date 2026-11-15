package memory

import (
	"context"
	"slices"
	"sync"
	"time"

	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureSummaries = (*ProcedureSummaries)(nil)

// ProcedureSummaries keeps the summaries it is given, in progress and closed.
// Until procedures can be started and moved, it holds the synthetic ones of
// local development and those of the tests.
type ProcedureSummaries struct {
	mu        sync.Mutex
	summaries []inbox.Summary
}

func NewProcedureSummaries(summaries ...inbox.Summary) *ProcedureSummaries {
	return &ProcedureSummaries{summaries: slices.Clone(summaries)}
}

// Add keeps the summary, in place of the one with the same id if there is one.
func (s *ProcedureSummaries) Add(summary inbox.Summary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, kept := range s.summaries {
		if kept.ID == summary.ID {
			s.summaries[i] = summary
			return
		}
	}
	s.summaries = append(s.summaries, summary)
}

// InProgress filters by the access, as a query would.
func (s *ProcedureSummaries) InProgress(_ context.Context, access inbox.Access) ([]inbox.Summary, error) {
	return s.visible(access, func(summary inbox.Summary) bool { return summary.ClosedAt == nil }), nil
}

func (s *ProcedureSummaries) ClosedSince(_ context.Context, access inbox.Access, since time.Time) ([]inbox.Summary, error) {
	return s.visible(access, func(summary inbox.Summary) bool {
		return summary.ClosedAt != nil && !summary.ClosedAt.Before(since)
	}), nil
}

func (s *ProcedureSummaries) Find(_ context.Context, access inbox.Access, id string) (inbox.Summary, bool, error) {
	found := s.visible(access, func(summary inbox.Summary) bool { return summary.ID == id })
	if len(found) == 0 {
		return inbox.Summary{}, false, nil
	}
	return found[0], true, nil
}

func (s *ProcedureSummaries) visible(access inbox.Access, keep func(inbox.Summary) bool) []inbox.Summary {
	s.mu.Lock()
	defer s.mu.Unlock()
	visible := []inbox.Summary{}
	for _, summary := range s.summaries {
		if access.CanSee(summary) && keep(summary) {
			visible = append(visible, summary)
		}
	}
	return visible
}
