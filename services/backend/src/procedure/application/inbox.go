package application

import (
	"context"
	"fmt"
	"time"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.Inbox = (*InboxService)(nil)

type InboxService struct {
	summaries output.ProcedureSummaries
	now       func() time.Time
}

func NewInboxService(summaries output.ProcedureSummaries, now func() time.Time) *InboxService {
	return &InboxService{summaries: summaries, now: now}
}

func (s *InboxService) List(ctx context.Context, viewer definition.Profile, q inbox.Query) (inbox.List, error) {
	summaries, err := s.summaries.InProgress(ctx)
	if err != nil {
		return inbox.List{}, fmt.Errorf("list procedures in progress: %w", err)
	}
	return inbox.Build(summaries, viewer, q, s.now()), nil
}
