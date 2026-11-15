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
	views       output.ProcedureViews
	now         func() time.Time
}

func NewInboxService(permissions output.Permissions, summaries output.ProcedureSummaries, views output.ProcedureViews, now func() time.Time) *InboxService {
	return &InboxService{permissions: permissions, summaries: summaries, views: views, now: now}
}

// List asks first what the person can see, and reads only that.
func (s *InboxService) List(ctx context.Context, person inbox.Person, q inbox.Query) (inbox.List, error) {
	access, err := s.staffAccess(ctx, person)
	if err != nil {
		return inbox.List{}, err
	}
	now := s.now()
	summaries, err := s.summaries.InProgress(ctx, access)
	if err != nil {
		return inbox.List{}, fmt.Errorf("list procedures in progress: %w", err)
	}
	closed, err := s.summaries.ClosedSince(ctx, access, inbox.MonthStart(now))
	if err != nil {
		return inbox.List{}, fmt.Errorf("list procedures closed this month: %w", err)
	}
	seen, err := s.views.SeenBy(ctx, person.ID)
	if err != nil {
		return inbox.List{}, fmt.Errorf("views of %s: %w", person.ID, err)
	}
	list := inbox.Build(summaries, access, seen, q, now)
	list.ClosedThisMonth = inbox.ClosedThisMonth(closed, access, now)
	return list, nil
}

// MarkSeen records the view only of a procedure the person can see.
func (s *InboxService) MarkSeen(ctx context.Context, person inbox.Person, procedureID string) error {
	access, err := s.staffAccess(ctx, person)
	if err != nil {
		return err
	}
	summary, ok, err := s.summaries.Find(ctx, access, procedureID)
	if err != nil {
		return fmt.Errorf("find procedure %s: %w", procedureID, err)
	}
	if !ok || !access.CanSee(summary) {
		return input.ErrProcedureNotFound
	}
	if err := s.views.See(ctx, person.ID, procedureID, s.now()); err != nil {
		return fmt.Errorf("record view of %s: %w", procedureID, err)
	}
	return nil
}

func (s *InboxService) staffAccess(ctx context.Context, person inbox.Person) (inbox.Access, error) {
	access, err := s.permissions.AccessOf(ctx, person)
	if err != nil {
		return inbox.Access{}, fmt.Errorf("access of %s: %w", person.ID, err)
	}
	if !access.Staff() {
		return inbox.Access{}, input.ErrNotStaff
	}
	return access, nil
}
