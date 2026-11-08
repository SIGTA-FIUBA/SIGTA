package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	authinput "backend/src/auth/ports/input"
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
	authorizer
	procedureTypes input.ProcedureTypes
	procedures     output.ProcedureRepository
	newID          func() string
}

func NewProcedureService(procedureTypes input.ProcedureTypes, procedures output.ProcedureRepository, staff authinput.StaffAuthorizer, deniedAttempts output.DeniedAttemptLog, newID func() string, clock func() time.Time) *ProcedureService {
	return &ProcedureService{
		authorizer:     authorizer{staff: staff, deniedAttempts: deniedAttempts, clock: clock},
		procedureTypes: procedureTypes,
		procedures:     procedures,
		newID:          newID,
	}
}

func (s *ProcedureService) Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error) {
	staff, err := s.staff.IsStaff(ctx, userID)
	if err != nil {
		return domain.Procedure{}, fmt.Errorf("role of %s: %w", userID, err)
	}
	if staff {
		return domain.Procedure{}, s.deny(ctx, userID, actionStart, "")
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
	reach, p, content, err := s.visible(ctx, userID, procedureID, actionOpen)
	if err != nil {
		return input.ProcedureView{}, err
	}
	view := input.ProcedureView{Procedure: p}
	allowed, err := s.canAct(ctx, userID, len(reach) > 0, p, content, p.OpenStep)
	if err != nil {
		return input.ProcedureView{}, err
	}
	if allowed {
		view.Actions = []string{p.OpenStep}
	}
	return view, nil
}

func (s *ProcedureService) List(ctx context.Context, userID string) ([]domain.Procedure, error) {
	reach, err := s.reach(ctx, userID)
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
		if s.canSee(userID, reach, p, content) {
			seen = append(seen, p)
		}
	}
	return seen, nil
}

func (s *ProcedureService) Act(ctx context.Context, userID, procedureID, stepID string) error {
	reach, p, content, err := s.visible(ctx, userID, procedureID, actionAct)
	if err != nil {
		return err
	}
	allowed, err := s.canAct(ctx, userID, len(reach) > 0, p, content, stepID)
	if err != nil {
		return err
	}
	if !allowed {
		return s.deny(ctx, userID, actionAct+" "+stepID, p.ID)
	}
	return nil
}

func (s *ProcedureService) reach(ctx context.Context, userID string) ([]authinput.AreaReach, error) {
	reach, err := s.staff.Reach(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("areas of %s: %w", userID, err)
	}
	return reach, nil
}

// visible returns the procedure if the person can see it, and answers as if
// it did not exist otherwise.
func (s *ProcedureService) visible(ctx context.Context, userID, procedureID, action string) ([]authinput.AreaReach, domain.Procedure, domain.Content, error) {
	reach, err := s.reach(ctx, userID)
	if err != nil {
		return nil, domain.Procedure{}, domain.Content{}, err
	}
	p, err := s.procedures.Get(ctx, procedureID)
	if err != nil {
		return nil, domain.Procedure{}, domain.Content{}, fmt.Errorf("procedure %s: %w", procedureID, err)
	}
	content, err := s.content(ctx, p)
	if err != nil {
		return nil, domain.Procedure{}, domain.Content{}, err
	}
	if !s.canSee(userID, reach, p, content) {
		if err := s.deny(ctx, userID, action, p.ID); !errors.Is(err, domain.ErrAccessDenied) {
			return nil, domain.Procedure{}, domain.Content{}, err
		}
		return nil, domain.Procedure{}, domain.Content{}, fmt.Errorf("procedure %s: %w", procedureID, domain.ErrProcedureNotFound)
	}
	return reach, p, content, nil
}

func (s *ProcedureService) content(ctx context.Context, p domain.Procedure) (domain.Content, error) {
	version, err := s.procedureTypes.Version(ctx, p.ProcedureType, p.Version)
	if err != nil {
		return domain.Content{}, err
	}
	content, err := version.Content()
	if err != nil {
		return domain.Content{}, fmt.Errorf("procedure type %d version %d: %w", p.ProcedureType, p.Version, err)
	}
	return content, nil
}

func (s *ProcedureService) latest(ctx context.Context, procedureType int64) (domain.Version, error) {
	entries, err := s.procedureTypes.Catalog(ctx)
	if err != nil {
		return domain.Version{}, err
	}
	i := slices.IndexFunc(entries, func(e input.CatalogEntry) bool { return e.ProcedureType == procedureType })
	if i < 0 {
		return domain.Version{}, fmt.Errorf("procedure type %d: %w", procedureType, output.ErrNoVersion)
	}
	return s.procedureTypes.Version(ctx, procedureType, entries[i].Version)
}
