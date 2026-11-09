package application

import (
	"context"
	"fmt"
	"time"

	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.Inbox = (*InboxService)(nil)

type InboxService struct {
	permissions output.Permissions
	summaries   output.ProcedureSummaries
	now         func() time.Time
}

func NewInboxService(permissions output.Permissions, summaries output.ProcedureSummaries, now func() time.Time) *InboxService {
	return &InboxService{permissions: permissions, summaries: summaries, now: now}
}

// List asks first what the person can see, and reads only that.
func (s *InboxService) List(ctx context.Context, person inbox.Person, q inbox.Query) (inbox.List, error) {
	access, err := s.permissions.AccessOf(ctx, person)
	if err != nil {
		return inbox.List{}, fmt.Errorf("access of %s: %w", person.ID, err)
	}
	if !access.Staff() {
		return inbox.List{}, input.ErrNotStaff
	}
	summaries, err := s.summaries.InProgress(ctx, access)
	if err != nil {
		return inbox.List{}, fmt.Errorf("list procedures in progress: %w", err)
	}
	return inbox.Build(summaries, access, q, s.now()), nil
}
