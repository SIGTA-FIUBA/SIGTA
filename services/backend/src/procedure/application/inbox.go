package application

import (
	"context"
	"fmt"
	"slices"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
)

var _ input.Inbox = (*InboxService)(nil)

// InboxService reads the same procedures and decides who sees and who acts as
// ProcedureService does, so the inbox and the detail answer the same to the
// same person.
type InboxService struct {
	procedures *ProcedureService
	now        func() time.Time
}

func NewInboxService(procedures *ProcedureService, now func() time.Time) *InboxService {
	return &InboxService{procedures: procedures, now: now}
}

// List leaves out the closed procedures and the person's own: the inbox is
// the work of the staff, and nobody moves their own procedure forward as
// staff. It fails if the open step of a procedure is not in its version, as
// the detail does.
func (s *InboxService) List(ctx context.Context, userID string, q inbox.Query) (inbox.List, error) {
	memberOf, err := s.procedures.memberOf(ctx, userID)
	if err != nil {
		return inbox.List{}, err
	}
	if len(memberOf) == 0 {
		return inbox.List{}, input.ErrNotStaff
	}
	all, err := s.procedures.procedures.List(ctx)
	if err != nil {
		return inbox.List{}, fmt.Errorf("list procedures: %w", err)
	}
	contents := map[version.Pin]version.Content{}
	var summaries []inbox.Summary
	for _, p := range all {
		if p.Closed() || p.Applicant == userID {
			continue
		}
		content, ok := contents[p.Pin()]
		if !ok {
			if content, err = s.procedures.content(ctx, p); err != nil {
				return inbox.List{}, err
			}
			contents[p.Pin()] = content
		}
		if !canSee(userID, memberOf, p, content) {
			continue
		}
		summary, err := s.summary(ctx, userID, p, content)
		if err != nil {
			return inbox.List{}, err
		}
		summaries = append(summaries, summary)
	}
	return inbox.Build(summaries, q, s.now()), nil
}

func (s *InboxService) summary(ctx context.Context, userID string, p domain.Procedure, c version.Content) (inbox.Summary, error) {
	state, err := domain.StateOf(p, c.Definition)
	if err != nil {
		return inbox.Summary{}, err
	}
	actionable, err := s.procedures.canAct(ctx, userID, p, c, p.OpenStep)
	if err != nil {
		return inbox.Summary{}, err
	}
	// StateOf already found the open step in the definition.
	step := c.Definition.Steps[slices.IndexFunc(c.Definition.Steps, func(st definition.Step) bool { return st.ID == p.OpenStep })]
	return inbox.Summary{
		ID:                p.ID,
		ProcedureType:     p.ProcedureType,
		ProcedureTypeName: c.Definition.Name,
		ApplicantID:       p.Applicant,
		StepName:          step.Name,
		Area:              areaOf(step, c.Areas),
		State:             state,
		Actionable:        actionable,
		StartedAt:         p.StartedAt,
		LastMovedAt:       p.MovedAt,
	}, nil
}

// areaOf names the area of a step with the name the version keeps, or its id
// if the version does not name it. It is nil on the steps of the applicant.
func areaOf(step definition.Step, areas []definition.Area) *inbox.Option[string] {
	if step.Role == definition.RoleApplicant || step.Area == "" {
		return nil
	}
	area := inbox.Option[string]{ID: step.Area, Name: step.Area}
	if i := slices.IndexFunc(areas, func(a definition.Area) bool { return a.ID == step.Area }); i >= 0 && areas[i].Name != "" {
		area.Name = areas[i].Name
	}
	return &area
}
