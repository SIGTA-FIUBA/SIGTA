package application

import (
	"context"
	"fmt"
	"slices"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.Procedures = (*ProcedureService)(nil)

const (
	actionStart = "start"
	actionAct   = "act"
)

type ProcedureService struct {
	staffAuthorizer authinput.StaffAuthorizer
	procedureTypes  input.ProcedureTypes
	procedures      output.ProcedureRepository
	newID           func() string
}

func NewProcedureService(procedureTypes input.ProcedureTypes, procedures output.ProcedureRepository, staff authinput.StaffAuthorizer, newID func() string) *ProcedureService {
	return &ProcedureService{
		staffAuthorizer: staff,
		procedureTypes:  procedureTypes,
		procedures:      procedures,
		newID:           newID,
	}
}

func (s *ProcedureService) Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error) {
	staff, err := s.staffAuthorizer.IsStaff(ctx, userID)
	if err != nil {
		return domain.Procedure{}, fmt.Errorf("role of %s: %w", userID, err)
	}
	if staff {
		return domain.Procedure{}, denied(userID, actionStart)
	}
	latest, err := s.latest(ctx, procedureType)
	if err != nil {
		return domain.Procedure{}, err
	}
	content, err := latest.DecodeContent()
	if err != nil {
		return domain.Procedure{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, latest.Number, err)
	}
	p := domain.Procedure{
		ID:            s.newID(),
		ProcedureType: procedureType,
		Version:       latest.Number,
		Applicant:     userID,
		OpenStep:      content.Definition.Steps[0].ID,
	}
	if err := s.procedures.Add(ctx, p); err != nil {
		return domain.Procedure{}, fmt.Errorf("start procedure: %w", err)
	}
	return p, nil
}

func (s *ProcedureService) Open(ctx context.Context, userID, procedureID string) (input.ProcedureView, error) {
	memberOf, p, content, err := s.visible(ctx, userID, procedureID)
	if err != nil {
		return input.ProcedureView{}, err
	}
	view := input.ProcedureView{Procedure: p}
	allowed, err := s.canAct(ctx, userID, len(memberOf) == 0, p, content, p.OpenStep)
	if err != nil {
		return input.ProcedureView{}, err
	}
	if allowed {
		view.Actions = []string{p.OpenStep}
	}
	return view, nil
}

func (s *ProcedureService) List(ctx context.Context, userID string) ([]domain.Procedure, error) {
	memberOf, err := s.memberOf(ctx, userID)
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
	contents := map[versionKey]version.Content{}
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
		if canSee(userID, memberOf, p, content) {
			seen = append(seen, p)
		}
	}
	return seen, nil
}

func (s *ProcedureService) Act(ctx context.Context, userID, procedureID, stepID string) error {
	memberOf, p, content, err := s.visible(ctx, userID, procedureID)
	if err != nil {
		return err
	}
	allowed, err := s.canAct(ctx, userID, len(memberOf) == 0, p, content, stepID)
	if err != nil {
		return err
	}
	if !allowed {
		return denied(userID, actionAct+" "+stepID)
	}
	return nil
}

func (s *ProcedureService) memberOf(ctx context.Context, userID string) ([]authinput.AreaReach, error) {
	memberOf, err := s.staffAuthorizer.Reach(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("areas of %s: %w", userID, err)
	}
	return memberOf, nil
}

// visible returns the procedure if the person can see it, and answers as if
// it did not exist otherwise.
func (s *ProcedureService) visible(ctx context.Context, userID, procedureID string) ([]authinput.AreaReach, domain.Procedure, version.Content, error) {
	memberOf, err := s.memberOf(ctx, userID)
	if err != nil {
		return nil, domain.Procedure{}, version.Content{}, err
	}
	p, err := s.procedures.Get(ctx, procedureID)
	if err != nil {
		return nil, domain.Procedure{}, version.Content{}, fmt.Errorf("procedure %s: %w", procedureID, err)
	}
	content, err := s.content(ctx, p)
	if err != nil {
		return nil, domain.Procedure{}, version.Content{}, err
	}
	if !canSee(userID, memberOf, p, content) {
		return nil, domain.Procedure{}, version.Content{}, fmt.Errorf("procedure %s: %w", procedureID, domain.ErrProcedureNotFound)
	}
	return memberOf, p, content, nil
}

func (s *ProcedureService) content(ctx context.Context, p domain.Procedure) (version.Content, error) {
	stored, err := s.procedureTypes.Version(ctx, p.ProcedureType, p.Version)
	if err != nil {
		return version.Content{}, err
	}
	content, err := stored.DecodeContent()
	if err != nil {
		return version.Content{}, fmt.Errorf("procedure type %d version %d: %w", p.ProcedureType, p.Version, err)
	}
	return content, nil
}

func (s *ProcedureService) latest(ctx context.Context, procedureType int64) (version.ProcedureTypeVersion, error) {
	entries, err := s.procedureTypes.Catalog(ctx)
	if err != nil {
		return version.ProcedureTypeVersion{}, err
	}
	i := slices.IndexFunc(entries, func(e input.CatalogEntry) bool { return e.ProcedureType == procedureType })
	if i < 0 {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d: %w", procedureType, output.ErrNoVersion)
	}
	return s.procedureTypes.Version(ctx, procedureType, entries[i].Version)
}
