package application

import (
	"context"
	"fmt"
	"slices"
	"time"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.Procedures = (*ProcedureService)(nil)

const (
	actionAct               = "act"
	actionConfirm           = "confirm"
	actionRequestCorrection = "request a correction on"
)

type ProcedureService struct {
	staffAuthorizer authinput.StaffAuthorizer
	procedureTypes  input.ProcedureTypes
	procedures      output.ProcedureRepository
	documents       output.DocumentRepository
	newID           func() string
	now             func() time.Time
}

func NewProcedureService(procedureTypes input.ProcedureTypes, procedures output.ProcedureRepository, documents output.DocumentRepository, staffAuthorizer authinput.StaffAuthorizer, newID func() string, now func() time.Time) *ProcedureService {
	return &ProcedureService{
		staffAuthorizer: staffAuthorizer,
		procedureTypes:  procedureTypes,
		procedures:      procedures,
		documents:       documents,
		newID:           newID,
		now:             now,
	}
}

func (s *ProcedureService) Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error) {
	latest, err := s.latest(ctx, procedureType)
	if err != nil {
		return domain.Procedure{}, err
	}
	content, err := latest.DecodeContent()
	if err != nil {
		return domain.Procedure{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, latest.Number, err)
	}
	now := s.now()
	p := domain.Procedure{
		ID:            s.newID(),
		ProcedureType: procedureType,
		Version:       latest.Number,
		VersionDigest: latest.Digest,
		Applicant:     userID,
		OpenStep:      content.Definition.Steps[0].ID,
		StartedAt:     now,
		MovedAt:       now,
	}
	if err := s.procedures.Add(ctx, output.Start{Procedure: p}); err != nil {
		return domain.Procedure{}, fmt.Errorf("start procedure: %w", err)
	}
	return p, nil
}

func (s *ProcedureService) Open(ctx context.Context, userID, procedureID string) (input.ProcedureView, error) {
	p, content, err := s.visible(ctx, userID, procedureID)
	if err != nil {
		return input.ProcedureView{}, err
	}
	view := input.ProcedureView{Procedure: p}
	allowed, err := s.canAct(ctx, userID, p, content, p.OpenStep)
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
	contents := map[version.Pin]version.Content{}
	var seen []domain.Procedure
	for _, p := range all {
		key := p.Pin()
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
	p, content, err := s.visible(ctx, userID, procedureID)
	if err != nil {
		return err
	}
	allowed, err := s.canAct(ctx, userID, p, content, stepID)
	if err != nil {
		return err
	}
	if !allowed {
		return denied(userID, actionAct+" "+stepID)
	}
	return nil
}

func (s *ProcedureService) Confirm(ctx context.Context, userID, procedureID string, c input.StepConfirmation) (domain.Event, error) {
	p, content, err := s.visible(ctx, userID, procedureID)
	if err != nil {
		return domain.Event{}, err
	}
	if p.Closed() || c.Step != p.OpenStep {
		return domain.Event{}, fmt.Errorf("procedure %s step %q: %w", p.ID, c.Step, engine.ErrStepNotOpen)
	}
	allowed, err := s.canAct(ctx, userID, p, content, c.Step)
	if err != nil {
		return domain.Event{}, err
	}
	if !allowed {
		return domain.Event{}, denied(userID, actionConfirm+" "+c.Step)
	}
	attached, err := s.attachedTo(ctx, p.ID, c.Step)
	if err != nil {
		return domain.Event{}, err
	}
	moved, event, err := engine.Confirm(p, content.Definition, engine.Confirmation{
		Step:               c.Step,
		By:                 userID,
		At:                 s.now(),
		Observation:        c.Observation,
		Outcome:            c.Outcome,
		ExternalFileNumber: c.ExternalFileNumber,
		Attached:           attached,
	})
	if err != nil {
		return domain.Event{}, fmt.Errorf("procedure %s step %q: %w", p.ID, c.Step, err)
	}
	if err := s.procedures.Advance(ctx, moved, event); err != nil {
		return domain.Event{}, fmt.Errorf("advance procedure %s: %w", p.ID, err)
	}
	return event, nil
}

func (s *ProcedureService) RequestCorrection(ctx context.Context, userID, procedureID string, r input.CorrectionRequest) (domain.Event, error) {
	p, content, err := s.visible(ctx, userID, procedureID)
	if err != nil {
		return domain.Event{}, err
	}
	if p.Closed() || r.Step != p.OpenStep {
		return domain.Event{}, fmt.Errorf("procedure %s step %q: %w", p.ID, r.Step, engine.ErrStepNotOpen)
	}
	allowed, err := s.canAct(ctx, userID, p, content, r.Step)
	if err != nil {
		return domain.Event{}, err
	}
	if !allowed {
		return domain.Event{}, denied(userID, actionRequestCorrection+" "+r.Step)
	}
	request, event, err := engine.RequestCorrection(p, content.Definition, domain.CorrectionRequest{
		Step:      r.Step,
		By:        userID,
		At:        s.now(),
		Reason:    r.Reason,
		Fields:    r.Fields,
		Documents: r.Documents,
	})
	if err != nil {
		return domain.Event{}, fmt.Errorf("procedure %s step %q: %w", p.ID, r.Step, err)
	}
	if err := s.procedures.RequestCorrection(ctx, request, event); err != nil {
		return domain.Event{}, fmt.Errorf("request a correction on procedure %s: %w", p.ID, err)
	}
	return event, nil
}

func (s *ProcedureService) memberOf(ctx context.Context, userID string) ([]authinput.AreaScope, error) {
	memberOf, err := s.staffAuthorizer.Scopes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("areas of %s: %w", userID, err)
	}
	return memberOf, nil
}

// visible returns the procedure if the person can see it, and answers as if
// it did not exist otherwise.
func (s *ProcedureService) visible(ctx context.Context, userID, procedureID string) (domain.Procedure, version.Content, error) {
	memberOf, err := s.memberOf(ctx, userID)
	if err != nil {
		return domain.Procedure{}, version.Content{}, err
	}
	p, err := s.procedures.Get(ctx, procedureID)
	if err != nil {
		return domain.Procedure{}, version.Content{}, fmt.Errorf("procedure %s: %w", procedureID, err)
	}
	content, err := s.content(ctx, p)
	if err != nil {
		return domain.Procedure{}, version.Content{}, err
	}
	if !canSee(userID, memberOf, p, content) {
		return domain.Procedure{}, version.Content{}, fmt.Errorf("procedure %s: %w", procedureID, domain.ErrProcedureNotFound)
	}
	return p, content, nil
}

func (s *ProcedureService) content(ctx context.Context, p domain.Procedure) (version.Content, error) {
	stored, err := s.procedureTypes.Pinned(ctx, p.Pin())
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
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d: %w", procedureType, output.ErrVersionNotFound)
	}
	return s.procedureTypes.Version(ctx, procedureType, entries[i].Version)
}

// attachedTo returns the documents of the procedure type that have a file
// attached to the step.
func (s *ProcedureService) attachedTo(ctx context.Context, procedureID, stepID string) ([]string, error) {
	documents, err := s.documents.List(ctx, procedureID)
	if err != nil {
		return nil, fmt.Errorf("documents of %s: %w", procedureID, err)
	}
	var attached []string
	for _, d := range documents {
		if d.Step == stepID {
			attached = append(attached, d.Requirement)
		}
	}
	return attached, nil
}
