package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/draft"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.ProcedureTypeDrafts = (*ProcedureTypeDraftService)(nil)

// ProcedureTypeDraftService asks for the permissions of the Secretariat: to
// edit procedure types to create, edit, propose and discard a draft, and to
// publish them to approve or reject one (ADR-020). A membership scoped to some
// procedure types only reaches their drafts, so only one without scope
// creates a procedure type.
type ProcedureTypeDraftService struct {
	drafts output.ProcedureTypeDraftRepository
	types  *ProcedureTypeService
	staff  authinput.StaffAuthorizer
	now    func() time.Time
}

func NewProcedureTypeDraftService(drafts output.ProcedureTypeDraftRepository, types *ProcedureTypeService, staff authinput.StaffAuthorizer, now func() time.Time) *ProcedureTypeDraftService {
	return &ProcedureTypeDraftService{drafts: drafts, types: types, staff: staff, now: now}
}

// newTypeAttempts bounds how many numbers Create tries when another draft
// takes the one it chose.
const newTypeAttempts = 3

// noProcedureType is what Create checks the permission against: no scope
// contains it, so only a membership without scope reaches it.
const noProcedureType = 0

func (s *ProcedureTypeDraftService) Create(ctx context.Context, userID string, def definition.ProcedureTypeDefinition) (input.DraftView, error) {
	if err := s.authorize(ctx, userID, noProcedureType, "create a procedure type", authinput.EditProcedureTypes); err != nil {
		return input.DraftView{}, err
	}
	for range newTypeAttempts {
		id, err := s.nextProcedureType(ctx)
		if err != nil {
			return input.DraftView{}, err
		}
		d := draft.New(id, 0, def, userID, s.at())
		err = s.drafts.Add(ctx, d)
		if errors.Is(err, output.ErrDraftExists) {
			continue
		}
		if err != nil {
			return input.DraftView{}, err
		}
		return s.view(ctx, d)
	}
	return input.DraftView{}, fmt.Errorf("create a procedure type: %d attempts lost to other drafts: %w", newTypeAttempts, output.ErrDraftExists)
}

// nextProcedureType is one more than the highest procedure type with a
// version or a draft. If a draft of a new procedure type is published and
// deleted between both reads, its number comes out again, and publishing the
// new draft fails with an outdated base instead of mixing them up.
func (s *ProcedureTypeDraftService) nextProcedureType(ctx context.Context) (int64, error) {
	versions, err := s.types.versions.LatestOfEach(ctx)
	if err != nil {
		return 0, fmt.Errorf("list latest versions: %w", err)
	}
	drafts, err := s.drafts.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list drafts: %w", err)
	}
	var highest int64
	for _, v := range versions {
		highest = max(highest, v.ProcedureTypeID)
	}
	for _, d := range drafts {
		highest = max(highest, d.ProcedureTypeID)
	}
	return highest + 1, nil
}

func (s *ProcedureTypeDraftService) Open(ctx context.Context, userID string, procedureType int64) (input.DraftView, error) {
	if err := s.authorize(ctx, userID, procedureType, fmt.Sprintf("edit procedure type %d", procedureType), authinput.EditProcedureTypes); err != nil {
		return input.DraftView{}, err
	}
	latest, err := s.types.latest(ctx, procedureType)
	if err != nil {
		return input.DraftView{}, err
	}
	if latest.Number == 0 {
		return input.DraftView{}, fmt.Errorf("procedure type %d: %w", procedureType, output.ErrVersionNotFound)
	}
	content, err := latest.DecodeContent()
	if err != nil {
		return input.DraftView{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, latest.Number, err)
	}
	d := draft.New(procedureType, latest.Number, content.Definition, userID, s.at())
	if err := s.drafts.Add(ctx, d); err != nil {
		return input.DraftView{}, fmt.Errorf("procedure type %d: %w", procedureType, err)
	}
	return s.view(ctx, d)
}

func (s *ProcedureTypeDraftService) Get(ctx context.Context, userID string, procedureType int64) (input.DraftView, error) {
	if err := s.authorize(ctx, userID, procedureType, fmt.Sprintf("see the draft of procedure type %d", procedureType), authinput.EditProcedureTypes, authinput.PublishProcedureTypes); err != nil {
		return input.DraftView{}, err
	}
	d, err := s.drafts.Get(ctx, procedureType)
	if err != nil {
		return input.DraftView{}, err
	}
	return s.view(ctx, d)
}

func (s *ProcedureTypeDraftService) List(ctx context.Context, userID string) ([]input.DraftView, error) {
	drafts, err := s.drafts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list drafts: %w", err)
	}
	views := make([]input.DraftView, 0, len(drafts))
	for _, d := range drafts {
		ok, err := s.holdsAny(ctx, userID, d.ProcedureTypeID, authinput.EditProcedureTypes, authinput.PublishProcedureTypes)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		view, err := s.view(ctx, d)
		if err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *ProcedureTypeDraftService) Edit(ctx context.Context, userID string, procedureType int64, revision int, def definition.ProcedureTypeDefinition) (input.DraftView, error) {
	return s.change(ctx, userID, procedureType, revision, "edit", authinput.EditProcedureTypes, func(d draft.Draft) (draft.Draft, error) {
		return d.Edit(def, userID, s.at())
	})
}

func (s *ProcedureTypeDraftService) Propose(ctx context.Context, userID string, procedureType int64, revision int) (input.DraftView, error) {
	return s.change(ctx, userID, procedureType, revision, "propose", authinput.EditProcedureTypes, func(d draft.Draft) (draft.Draft, error) {
		if err := s.validate(ctx, d); err != nil {
			return draft.Draft{}, err
		}
		return d.Propose(userID, s.at())
	})
}

func (s *ProcedureTypeDraftService) Reject(ctx context.Context, userID string, procedureType int64, revision int, reason string) (input.DraftView, error) {
	return s.change(ctx, userID, procedureType, revision, "reject", authinput.PublishProcedureTypes, func(d draft.Draft) (draft.Draft, error) {
		return d.Reject(userID, reason, s.at())
	})
}

// Approve publishes before it deletes the draft, since the stores of both are
// apart. If the draft changed in between, as when someone else rejected it at
// the same time, the version is published anyway and the draft stays, with
// the same content as the version.
func (s *ProcedureTypeDraftService) Approve(ctx context.Context, userID string, procedureType int64, revision int) (version.ProcedureTypeVersion, error) {
	d, err := s.current(ctx, userID, procedureType, revision, "approve", authinput.PublishProcedureTypes)
	if err != nil {
		return version.ProcedureTypeVersion{}, err
	}
	authorship, err := d.Approval(userID)
	if err != nil {
		return version.ProcedureTypeVersion{}, fmt.Errorf("approve the draft of procedure type %d: %w", procedureType, err)
	}
	v, err := s.types.Publish(ctx, input.Publication{ProcedureType: procedureType, Base: d.Base, Definition: d.Definition, Authorship: authorship})
	if err != nil {
		return version.ProcedureTypeVersion{}, err
	}
	err = s.drafts.Remove(ctx, procedureType, d.Revision)
	if err != nil && !errors.Is(err, output.ErrDraftChanged) && !errors.Is(err, output.ErrDraftNotFound) {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d version %d is published, but its draft was not removed: %w", procedureType, v.Number, err)
	}
	return v, nil
}

func (s *ProcedureTypeDraftService) Discard(ctx context.Context, userID string, procedureType int64, revision int) error {
	d, err := s.current(ctx, userID, procedureType, revision, "discard", authinput.EditProcedureTypes)
	if err != nil {
		return err
	}
	if d.Status == draft.StatusProposed {
		return fmt.Errorf("discard the draft of procedure type %d: %w", procedureType, draft.ErrDraftProposed)
	}
	return s.drafts.Remove(ctx, procedureType, d.Revision)
}

// change applies apply to the draft at revision and stores the result.
func (s *ProcedureTypeDraftService) change(ctx context.Context, userID string, procedureType int64, revision int, action, permission string, apply func(draft.Draft) (draft.Draft, error)) (input.DraftView, error) {
	d, err := s.current(ctx, userID, procedureType, revision, action, permission)
	if err != nil {
		return input.DraftView{}, err
	}
	next, err := apply(d)
	if err != nil {
		return input.DraftView{}, fmt.Errorf("%s the draft of procedure type %d: %w", action, procedureType, err)
	}
	if err := s.drafts.Update(ctx, next); err != nil {
		return input.DraftView{}, err
	}
	return s.view(ctx, next)
}

// current checks the permission before reading the draft, so that whoever
// lacks it does not learn whether there is one.
func (s *ProcedureTypeDraftService) current(ctx context.Context, userID string, procedureType int64, revision int, action, permission string) (draft.Draft, error) {
	if err := s.authorize(ctx, userID, procedureType, fmt.Sprintf("%s the draft of procedure type %d", action, procedureType), permission); err != nil {
		return draft.Draft{}, err
	}
	d, err := s.drafts.Get(ctx, procedureType)
	if err != nil {
		return draft.Draft{}, err
	}
	if d.Revision != revision {
		return draft.Draft{}, fmt.Errorf("%w: procedure type %d is at revision %d, not %d", output.ErrDraftChanged, procedureType, d.Revision, revision)
	}
	return d, nil
}

func (s *ProcedureTypeDraftService) authorize(ctx context.Context, userID string, procedureType int64, action string, permissions ...string) error {
	ok, err := s.holdsAny(ctx, userID, procedureType, permissions...)
	if err != nil {
		return err
	}
	if !ok {
		return denied(userID, action)
	}
	return nil
}

func (s *ProcedureTypeDraftService) holdsAny(ctx context.Context, userID string, procedureType int64, permissions ...string) (bool, error) {
	for _, p := range permissions {
		ok, err := s.staff.Holds(ctx, userID, authinput.SecretariatArea, procedureType, []string{p})
		if err != nil {
			return false, fmt.Errorf("permissions of %s in %s: %w", userID, authinput.SecretariatArea, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (s *ProcedureTypeDraftService) validate(ctx context.Context, d draft.Draft) error {
	areas, err := s.types.currentStaffAreas(ctx)
	if err != nil {
		return err
	}
	return d.Definition.Validate(areas)
}

func (s *ProcedureTypeDraftService) view(ctx context.Context, d draft.Draft) (input.DraftView, error) {
	err := s.validate(ctx, d)
	if invalid, ok := errors.AsType[*definition.ValidationError](err); ok {
		return input.DraftView{Draft: d, Problems: invalid.Problems}, nil
	}
	if err != nil {
		return input.DraftView{}, err
	}
	return input.DraftView{Draft: d}, nil
}

// at keeps microseconds, as PostgreSQL does.
func (s *ProcedureTypeDraftService) at() time.Time {
	return s.now().UTC().Truncate(time.Microsecond)
}
