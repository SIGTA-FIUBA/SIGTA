package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.Procedures = (*ProcedureService)(nil)

const (
	actionStart = "start"
	actionOpen  = "open"
	actionAct   = "act"
)

type ProcedureService struct {
	gate
	types      input.ProcedureTypes
	procedures output.ProcedureRepository
	newID      func() string
}

func NewProcedureService(types input.ProcedureTypes, procedures output.ProcedureRepository, access output.Access, denied output.DeniedAttempts, newID func() string, now func() time.Time) *ProcedureService {
	return &ProcedureService{gate: gate{access: access, denied: denied, now: now}, types: types, procedures: procedures, newID: newID}
}

func (s *ProcedureService) Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error) {
	actor, err := s.actor(ctx, userID)
	if err != nil {
		return domain.Procedure{}, err
	}
	if actor.Role() != domain.RoleApplicant {
		return domain.Procedure{}, s.deny(ctx, actor, actionStart, "")
	}
	version, err := s.latest(ctx, procedureType)
	if err != nil {
		return domain.Procedure{}, err
	}
	content, err := version.Content()
	if err != nil {
		return domain.Procedure{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, version.Number, err)
	}
	p := domain.Procedure{
		ID:            s.newID(),
		ProcedureType: procedureType,
		Version:       version.Number,
		Applicant:     userID,
		OpenStep:      content.Definition.Steps[0].ID,
	}
	if err := s.procedures.Add(ctx, p); err != nil {
		return domain.Procedure{}, fmt.Errorf("start procedure: %w", err)
	}
	return p, nil
}

func (s *ProcedureService) Open(ctx context.Context, userID, procedureID string) (input.ProcedureView, error) {
	actor, p, content, err := s.visible(ctx, userID, procedureID, actionOpen)
	if err != nil {
		return input.ProcedureView{}, err
	}
	view := input.ProcedureView{Procedure: p}
	allowed, err := canAct(actor, p, content, p.OpenStep)
	if err != nil {
		return input.ProcedureView{}, err
	}
	if allowed {
		view.Actions = []string{p.OpenStep}
	}
	return view, nil
}

func (s *ProcedureService) List(ctx context.Context, userID string) ([]domain.Procedure, error) {
	actor, err := s.actor(ctx, userID)
	if err != nil {
		return nil, err
	}
	all, err := s.procedures.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list procedures: %w", err)
	}
	type versionKey struct {
		procedureType int64
		number        int
	}
	contents := map[versionKey]domain.Content{}
	var seen []domain.Procedure
	for _, p := range all {
		key := versionKey{p.ProcedureType, p.Version}
		content, ok := contents[key]
		if !ok {
			if content, err = s.content(ctx, p); err != nil {
				return nil, err
			}
			contents[key] = content
		}
		if actor.CanSee(p, content) {
			seen = append(seen, p)
		}
	}
	return seen, nil
}

func (s *ProcedureService) Act(ctx context.Context, userID, procedureID, stepID string) error {
	actor, p, content, err := s.visible(ctx, userID, procedureID, actionAct)
	if err != nil {
		return err
	}
	allowed, err := canAct(actor, p, content, stepID)
	if err != nil {
		return err
	}
	if !allowed {
		return s.deny(ctx, actor, actionAct+" "+stepID, p.ID)
	}
	return nil
}

func (s *ProcedureService) visible(ctx context.Context, userID, procedureID, action string) (domain.Actor, domain.Procedure, domain.Content, error) {
	actor, err := s.actor(ctx, userID)
	if err != nil {
		return domain.Actor{}, domain.Procedure{}, domain.Content{}, err
	}
	p, err := s.procedures.Get(ctx, procedureID)
	if err != nil {
		return domain.Actor{}, domain.Procedure{}, domain.Content{}, fmt.Errorf("procedure %s: %w", procedureID, err)
	}
	content, err := s.content(ctx, p)
	if err != nil {
		return domain.Actor{}, domain.Procedure{}, domain.Content{}, err
	}
	if !actor.CanSee(p, content) {
		if err := s.deny(ctx, actor, action, p.ID); !errors.Is(err, domain.ErrAccessDenied) {
			return domain.Actor{}, domain.Procedure{}, domain.Content{}, err
		}
		return domain.Actor{}, domain.Procedure{}, domain.Content{}, fmt.Errorf("procedure %s: %w", procedureID, domain.ErrProcedureNotFound)
	}
	return actor, p, content, nil
}

func (s *ProcedureService) content(ctx context.Context, p domain.Procedure) (domain.Content, error) {
	version, err := s.types.Version(ctx, p.ProcedureType, p.Version)
	if err != nil {
		return domain.Content{}, err
	}
	content, err := version.Content()
	if err != nil {
		return domain.Content{}, fmt.Errorf("procedure type %d version %d: %w", p.ProcedureType, p.Version, err)
	}
	return content, nil
}

func canAct(actor domain.Actor, p domain.Procedure, content domain.Content, stepID string) (bool, error) {
	if stepID != p.OpenStep {
		return false, nil
	}
	i := slices.IndexFunc(content.Definition.Steps, func(st domain.Step) bool { return st.ID == stepID })
	if i < 0 {
		return false, fmt.Errorf("procedure %s: open step %q is not in version %d", p.ID, stepID, p.Version)
	}
	step := content.Definition.Steps[i]
	j := slices.IndexFunc(content.Areas, func(a domain.Area) bool { return a.ID == step.Area })
	if j < 0 {
		return false, fmt.Errorf("procedure %s: area %q is not in version %d", p.ID, step.Area, p.Version)
	}
	return actor.CanActOn(p, step, content.Areas[j]), nil
}

func (s *ProcedureService) latest(ctx context.Context, procedureType int64) (domain.Version, error) {
	entries, err := s.types.Catalog(ctx)
	if err != nil {
		return domain.Version{}, err
	}
	i := slices.IndexFunc(entries, func(e input.CatalogEntry) bool { return e.ProcedureType == procedureType })
	if i < 0 {
		return domain.Version{}, fmt.Errorf("procedure type %d: %w", procedureType, output.ErrNoVersion)
	}
	return s.types.Version(ctx, procedureType, entries[i].Version)
}
