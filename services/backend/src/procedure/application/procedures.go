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
	types      input.ProcedureTypes
	procedures output.Procedures
	access     output.Access
	denied     output.DeniedAttempts
	newID      func() string
	now        func() time.Time
}

func NewProcedureService(types input.ProcedureTypes, procedures output.Procedures, access output.Access, denied output.DeniedAttempts, newID func() string, now func() time.Time) *ProcedureService {
	return &ProcedureService{types: types, procedures: procedures, access: access, denied: denied, newID: newID, now: now}
}

func (s *ProcedureService) Register(ctx context.Context, userID string, kind domain.UserKind) error {
	if !kind.Known() {
		return fmt.Errorf("%w: %q", domain.ErrUnknownUserKind, kind)
	}
	if err := s.access.Register(ctx, userID, kind); err != nil {
		return fmt.Errorf("register %s: %w", userID, err)
	}
	return nil
}

func (s *ProcedureService) Grant(ctx context.Context, p domain.Permission) error {
	actor, err := s.access.Actor(ctx, p.UserID)
	if errors.Is(err, output.ErrUnknownUser) {
		return fmt.Errorf("%w: %s is not registered", domain.ErrInvalidPermission, p.UserID)
	}
	if err != nil {
		return fmt.Errorf("grant to %s: %w", p.UserID, err)
	}
	if actor.Kind != domain.KindStaff {
		return fmt.Errorf("%w: %s is %s, not staff", domain.ErrInvalidPermission, p.UserID, actor.Kind)
	}
	if p.Profile != domain.ProfileStaffOperative && p.Profile != domain.ProfileStaffApprover {
		return fmt.Errorf("%w: unknown profile %q", domain.ErrInvalidPermission, p.Profile)
	}
	if _, err := s.latest(ctx, p.ProcedureType); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidPermission, err)
	}
	if err := s.access.Grant(ctx, p); err != nil {
		return fmt.Errorf("grant to %s: %w", p.UserID, err)
	}
	return nil
}

func (s *ProcedureService) Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error) {
	actor, err := s.actor(ctx, userID)
	if err != nil {
		return domain.Procedure{}, err
	}
	if actor.Kind != domain.KindApplicant {
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
	actor, p, err := s.visible(ctx, userID, procedureID, actionOpen)
	if err != nil {
		return input.ProcedureView{}, err
	}
	view := input.ProcedureView{Procedure: p}
	allowed, err := s.canAct(ctx, actor, p, p.OpenStep)
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
	return slices.DeleteFunc(all, func(p domain.Procedure) bool { return !actor.CanSee(p) }), nil
}

// Act only checks who can act for now: what acting records and where the
// procedure goes next comes with the step engine.
func (s *ProcedureService) Act(ctx context.Context, userID, procedureID, stepID string) error {
	actor, p, err := s.visible(ctx, userID, procedureID, actionAct)
	if err != nil {
		return err
	}
	allowed, err := s.canAct(ctx, actor, p, stepID)
	if err != nil {
		return err
	}
	if !allowed {
		return s.deny(ctx, actor, actionAct+" "+stepID, p.ID)
	}
	return nil
}

func (s *ProcedureService) actor(ctx context.Context, userID string) (domain.Actor, error) {
	actor, err := s.access.Actor(ctx, userID)
	if errors.Is(err, output.ErrUnknownUser) {
		// Someone SIGTA does not know has no permission at all.
		return domain.Actor{UserID: userID}, nil
	}
	if err != nil {
		return domain.Actor{}, fmt.Errorf("actor %s: %w", userID, err)
	}
	return actor, nil
}

// visible returns the procedure if the actor can see it, and records the
// attempt otherwise.
func (s *ProcedureService) visible(ctx context.Context, userID, procedureID, action string) (domain.Actor, domain.Procedure, error) {
	actor, err := s.actor(ctx, userID)
	if err != nil {
		return domain.Actor{}, domain.Procedure{}, err
	}
	p, err := s.procedures.Get(ctx, procedureID)
	if err != nil {
		return domain.Actor{}, domain.Procedure{}, fmt.Errorf("procedure %s: %w", procedureID, err)
	}
	if !actor.CanSee(p) {
		if err := s.deny(ctx, actor, action, p.ID); !errors.Is(err, domain.ErrAccessDenied) {
			return domain.Actor{}, domain.Procedure{}, err
		}
		return domain.Actor{}, domain.Procedure{}, fmt.Errorf("procedure %s: %w", procedureID, domain.ErrProcedureNotFound)
	}
	return actor, p, nil
}

func (s *ProcedureService) canAct(ctx context.Context, actor domain.Actor, p domain.Procedure, stepID string) (bool, error) {
	if stepID != p.OpenStep {
		return false, nil
	}
	version, err := s.types.Version(ctx, p.ProcedureType, p.Version)
	if err != nil {
		return false, err
	}
	content, err := version.Content()
	if err != nil {
		return false, fmt.Errorf("procedure type %d version %d: %w", p.ProcedureType, p.Version, err)
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

// deny records the attempt and returns domain.ErrAccessDenied, or the error
// that kept it from being recorded: a refusal is never left unrecorded.
func (s *ProcedureService) deny(ctx context.Context, actor domain.Actor, action, procedureID string) error {
	attempt := domain.DeniedAttempt{UserID: actor.UserID, Action: action, Procedure: procedureID, At: s.now().UTC()}
	if err := s.denied.Record(ctx, attempt); err != nil {
		return fmt.Errorf("record denied attempt: %w", err)
	}
	return fmt.Errorf("%s %s: %w", actor.UserID, action, domain.ErrAccessDenied)
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
